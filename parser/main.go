package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"baliance.com/gooxml/document"
	"baliance.com/gooxml/schema/soo/wml"
	"github.com/xuri/excelize/v2"
	ptime "github.com/yaa110/go-persian-calendar"
)

type Event struct {
	Timestamp time.Time
	Type      string
	Title     string
}

var f = "B Nazanin"
var ft = "Times New Roman"

type WorkSession struct {
	Start       time.Time
	End         time.Time
	SolarStart  ptime.Time
	Subtasks    []string
	Duration    time.Duration
	PauseTimes  []time.Time // Tracks start of each pause
	ResumeTimes []time.Time // Tracks end of each pause
}

type DailyTask struct {
	SolarDate  ptime.Time
	TotalHours float64
	Sessions   []WorkSession
}

func main() {
	events, err := parseExcel("time_tracker.xlsx")
	if err != nil {
		log.Fatal(err)
	}

	dailyTasks := groupSessions(events)
	doc := createDocument(dailyTasks)

	if err := saveDocument(doc, "output.docx"); err != nil {
		log.Fatal(err)
	}
	log.Println("✅ Document generated successfully")
}

func parseExcel(path string) ([]Event, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel: %w", err)
	}
	defer f.Close()

	var events []Event
	for _, sheet := range f.GetSheetList() {
		rows, err := f.GetRows(sheet)
		if err != nil {
			return nil, fmt.Errorf("error reading sheet %s: %w", sheet, err)
		}

		for i, row := range rows {
			if i == 0 || len(row) < 3 {
				continue
			}

			event, err := parseRow(row)
			if err != nil {
				log.Printf("Skipping row %d: %v", i+1, err)
				continue
			}

			events = append(events, event)
		}
	}
	log.Printf("Parsed %d events from Excel", len(events))
	return events, nil
}

func parseRow(row []string) (Event, error) {
	timeStr := strings.TrimSpace(row[0])
	eventType := strings.TrimSpace(row[1])
	title := strings.TrimSpace(row[2])

	parsedTime, err := time.Parse("1/2/06 15:04", timeStr)
	if err != nil {
		return Event{}, fmt.Errorf("invalid time format: %w", err)
	}

	return Event{
		Timestamp: parsedTime,
		Type:      eventType,
		Title:     title,
	}, nil
}

func groupSessions(events []Event) []DailyTask {
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	var sessions []WorkSession
	var currentSession *WorkSession
	var lastEventTime time.Time
	var isPaused bool

	for _, event := range events {
		switch event.Type {
		case "START":
			if currentSession != nil {
				// End the previous session if not paused
				if !isPaused {
					currentSession.End = lastEventTime
					currentSession.Duration = calculateDuration(currentSession)
				}
				sessions = append(sessions, *currentSession)
			}
			currentSession = &WorkSession{
				Start:       event.Timestamp,
				SolarStart:  ptime.New(event.Timestamp),
				Subtasks:    []string{event.Title},
				PauseTimes:  []time.Time{},
				ResumeTimes: []time.Time{},
			}
			isPaused = false
		case "PAUSE":
			if currentSession != nil && !isPaused {
				currentSession.PauseTimes = append(currentSession.PauseTimes, event.Timestamp)
				currentSession.Subtasks = append(currentSession.Subtasks, event.Title)
				isPaused = true
			}
		case "RESUME":
			if currentSession != nil && isPaused {
				currentSession.ResumeTimes = append(currentSession.ResumeTimes, event.Timestamp)
				currentSession.Subtasks = append(currentSession.Subtasks, event.Title)
				isPaused = false
			}
		case "STOP":
			if currentSession != nil {
				currentSession.End = event.Timestamp
				if !isPaused {
					currentSession.Duration = calculateDuration(currentSession)
				} else {
					// If paused, duration stops at last pause
					currentSession.Duration = calculateDuration(currentSession)
				}
				sessions = append(sessions, *currentSession)
				currentSession = nil
				isPaused = false
			}
		default:
			if currentSession != nil {
				currentSession.Subtasks = append(currentSession.Subtasks, event.Title)
			}
		}
		lastEventTime = event.Timestamp
	}

	if currentSession != nil {
		currentSession.End = lastEventTime
		currentSession.Duration = calculateDuration(currentSession)
		sessions = append(sessions, *currentSession)
	}

	dailyMap := make(map[string]*DailyTask)
	for _, session := range sessions {
		solarDate := session.SolarStart
		dateKey := fmt.Sprintf("%d-%02d-%02d", solarDate.Year(), solarDate.Month(), solarDate.Day())

		if _, exists := dailyMap[dateKey]; !exists {
			dailyMap[dateKey] = &DailyTask{
				SolarDate: solarDate,
			}
		}

		daily := dailyMap[dateKey]
		daily.TotalHours += session.Duration.Hours()
		daily.Sessions = append(daily.Sessions, session)
	}

	var dailyTasks []DailyTask
	for _, daily := range dailyMap {
		dailyTasks = append(dailyTasks, *daily)
	}
	sort.Slice(dailyTasks, func(i, j int) bool {
		return dailyTasks[i].SolarDate.Time().Before(dailyTasks[j].SolarDate.Time())
	})

	log.Printf("Generated %d daily tasks", len(dailyTasks))
	return dailyTasks
}

func calculateDuration(session *WorkSession) time.Duration {
	if session.End.IsZero() {
		return 0
	}

	totalDuration := time.Duration(0)
	start := session.Start

	// Handle all pause-resume pairs
	for i := 0; i < len(session.PauseTimes); i++ {
		pauseTime := session.PauseTimes[i]
		if pauseTime.Before(start) {
			continue
		}

		// Add time from start (or last resume) to pause
		totalDuration += pauseTime.Sub(start)

		// Set start to resume time, if available
		if i < len(session.ResumeTimes) {
			resumeTime := session.ResumeTimes[i]
			if resumeTime.After(pauseTime) {
				start = resumeTime
			}
		}
	}

	// If not paused, add time from last resume (or start) to end
	if len(session.PauseTimes) == len(session.ResumeTimes) {
		totalDuration += session.End.Sub(start)
	}

	return totalDuration
}

func createDocument(dailyTasks []DailyTask) *document.Document {
	doc := document.New()
	tbl := doc.AddTable()
	tbl.Properties().SetWidthPercent(100)

	log.Printf("Processing %d daily tasks", len(dailyTasks))

	// Create header
	header := tbl.AddRow()
	for _, text := range []string{"تاریخ (شمسی)", "کل ساعات تا ۱۱ام", "فعالیت‌ها", "مدت زمان", "نتایج"} {
		cell := header.AddCell()
		cell.Properties().SetVerticalAlignment(wml.ST_VerticalJcCenter)
		para := cell.AddParagraph()
		para.Properties().SetAlignment(wml.ST_JcCenter)
		paraProps := para.Properties().X()
		if paraProps != nil {
			paraProps.Bidi = wml.NewCT_OnOff()
		} else {
			log.Println("Warning: para.Properties().X() is nil for header")
		}
		run := para.AddRun()
		run.Properties().SetBold(true)
		runProps := run.Properties().X()
		if runProps != nil {
			rFonts := wml.NewCT_Fonts()
			rFonts.AsciiAttr = &f
			rFonts.HAnsiAttr = &f
			rFonts.CsAttr = &f
			runProps.RFonts = rFonts
		} else {
			log.Println("Warning: run.Properties().X() is nil for header")
		}
		run.AddText(text)
	}

	cumulative := make(map[string]float64)

	for i, daily := range dailyTasks {
		log.Printf("Processing daily task %d: Date %s, Sessions %d", i, daily.SolarDate.Format("yyyy/MM/dd"), len(daily.Sessions))
		row := tbl.AddRow()

		// Solar date (Farsi, B Nazanin, RTL)
		dateCell := row.AddCell()
		para := dateCell.AddParagraph()
		para.Properties().SetAlignment(wml.ST_JcRight)
		paraProps := para.Properties().X()
		if paraProps != nil {
			paraProps.Bidi = wml.NewCT_OnOff()
		} else {
			log.Println("Warning: para.Properties().X() is nil for solar date")
		}
		run := para.AddRun()
		runProps := run.Properties().X()
		if runProps != nil {
			rFonts := wml.NewCT_Fonts()
			rFonts.AsciiAttr = &f
			rFonts.HAnsiAttr = &f
			rFonts.CsAttr = &f
			runProps.RFonts = rFonts
		} else {
			log.Println("Warning: run.Properties().X() is nil for solar date")
		}
		yearMonthDay := toPersianDigits(daily.SolarDate.Format("yyyy/MM/dd"))
		run.AddText(yearMonthDay)

		// Cumulative hours till 11th (Farsi, B Nazanin, RTL)
		year, month, _ := daily.SolarDate.Date()
		key := fmt.Sprintf("%d-%02d", year, month)
		cumulative[key] += daily.TotalHours
		cumulativeText := ""
		if daily.SolarDate.Day() == 11 {
			hours := toPersianDigits(fmt.Sprintf("%.1f", cumulative[key]))
			cumulativeText = fmt.Sprintf("%s ساعت", hours)
		}
		cell := row.AddCell()
		para = cell.AddParagraph()
		para.Properties().SetAlignment(wml.ST_JcRight)
		paraProps = para.Properties().X()
		if paraProps != nil {
			paraProps.Bidi = wml.NewCT_OnOff()
		} else {
			log.Println("Warning: para.Properties().X() is nil for cumulative hours")
		}
		run = para.AddRun()
		runProps = run.Properties().X()
		if runProps != nil {
			rFonts := wml.NewCT_Fonts()
			rFonts.AsciiAttr = &f
			rFonts.HAnsiAttr = &f
			rFonts.CsAttr = &f
			runProps.RFonts = rFonts
		} else {
			log.Println("Warning: run.Properties().X() is nil for cumulative hours")
		}
		run.AddText(cumulativeText)

		// Activities (Farsi: B Nazanin, Latin: Times New Roman, RTL)
		activities := strings.Builder{}
		for j, session := range daily.Sessions {
			if session.Subtasks == nil {
				log.Printf("Warning: session %d has nil Subtasks", j)
				continue
			}
			original := strings.Join(unique(session.Subtasks), "\n")
			translated := safeTranslate(original)
			activities.WriteString(translated)
			duration := toPersianDigits(fmt.Sprintf("%.1f", session.Duration.Hours()))
			activities.WriteString(fmt.Sprintf(" (%s ساعت)\n", duration))
		}
		cell = row.AddCell()
		para = cell.AddParagraph()
		para.Properties().SetAlignment(wml.ST_JcRight)
		paraProps = para.Properties().X()
		if paraProps != nil {
			paraProps.Bidi = wml.NewCT_OnOff()
		} else {
			log.Println("Warning: para.Properties().X() is nil for activities")
		}
		lines := strings.Split(activities.String(), "\n")
		for i, line := range lines {
			if line == "" {
				continue
			}
			run = para.AddRun()
			runProps = run.Properties().X()
			isLatin := false
			for _, session := range daily.Sessions {
				if session.Subtasks == nil {
					continue
				}
				original := strings.Join(unique(session.Subtasks), "\n")
				if strings.Contains(line, original) && safeTranslate(original) == original {
					isLatin = true
					break
				}
			}
			if runProps != nil {
				rFonts := wml.NewCT_Fonts()
				if isLatin {
					rFonts.AsciiAttr = &ft
					rFonts.HAnsiAttr = &ft
					rFonts.CsAttr = &ft
				} else {
					rFonts.AsciiAttr = &f
					rFonts.HAnsiAttr = &f
					rFonts.CsAttr = &f
				}
				runProps.RFonts = rFonts
			} else {
				log.Println("Warning: run.Properties().X() is nil for activity line")
			}
			run.AddText(line)
			if i < len(lines)-1 {
				run.AddBreak()
			}
		}

		// Total duration (Farsi, B Nazanin, RTL)
		cell = row.AddCell()
		para = cell.AddParagraph()
		para.Properties().SetAlignment(wml.ST_JcRight)
		paraProps = para.Properties().X()
		if paraProps != nil {
			paraProps.Bidi = wml.NewCT_OnOff()
		} else {
			log.Println("Warning: para.Properties().X() is nil for total duration")
		}
		run = para.AddRun()
		runProps = run.Properties().X()
		if runProps != nil {
			rFonts := wml.NewCT_Fonts()
			rFonts.AsciiAttr = &f
			rFonts.HAnsiAttr = &f
			rFonts.CsAttr = &f
			runProps.RFonts = rFonts
		} else {
			log.Println("Warning: run.Properties().X() is nil for total duration")
		}
		totalHours := toPersianDigits(fmt.Sprintf("%.1f", daily.TotalHours))
		run.AddText(fmt.Sprintf("%s ساعت", totalHours))

		// Results (empty)
		row.AddCell().AddParagraph().AddRun().AddText("")
	}

	return doc
}

func unique(items []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

func safeTranslate(text string) string {
	translated, err := translateToFarsi(text)
	if err != nil {
		log.Printf("Translation failed: %v", err)
		return text
	}
	return translated
}

func translateToFarsi(text string) (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("https://translate.googleapis.com/translate_a/single?client=gtx&sl=en&tl=fa&dt=t&q=%s",
		url.QueryEscape(text))

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result []any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result) < 1 {
		return text, nil
	}

	var translated strings.Builder
	for _, segment := range result[0].([]any) {
		if pair, ok := segment.([]any); ok && len(pair) > 0 {
			if text, ok := pair[0].(string); ok {
				translated.WriteString(text)
			}
		}
	}

	return translated.String(), nil
}

func saveDocument(doc *document.Document, path string) error {
	tmpPath := path + ".tmp"
	if err := doc.SaveToFile(tmpPath); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func toPersianDigits(s string) string {
	persianDigits := []rune{'۰', '۱', '۲', '۳', '۴', '۵', '۶', '۷', '۸', '۹'}
	var result strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			result.WriteRune(persianDigits[r-'0'])
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
