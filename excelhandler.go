package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/xuri/excelize/v2"
)

func initExcelFile() {
	date := time.Now().Format("2006-01-02")
	sheetName := date

	if _, err := os.Stat(excelFileName); os.IsNotExist(err) {
		f := excelize.NewFile()
		f.NewSheet(sheetName)
		f.SetCellValue(sheetName, "A1", "Timestamp")
		f.SetCellValue(sheetName, "B1", "Event")
		f.SetCellValue(sheetName, "C1", "Activity Name")
		f.SetCellValue(sheetName, "D1", "Duration")
		f.DeleteSheet("Sheet1")
		if err := f.SaveAs(excelFileName); err != nil {
			t := fmt.Sprint("Error creating Excel file:", err)
			LogEntry.SetText(t)
		}
	}
}

func saveToExcel(entry TimerEntry) {
	f, err := excelize.OpenFile(excelFileName)
	if err != nil {
		fmt.Println("Error opening Excel file:", err)
		return
	}
	defer f.Close()

	date := entry.Timestamp.Format("2006-01-02")
	sheetName := date

	// Create a new sheet for a new day if it doesn't exist
	if sheetIndex, err := f.GetSheetIndex(sheetName); err != nil || sheetIndex == -1 {
		f.NewSheet(sheetName)
		f.SetCellValue(sheetName, "A1", "Timestamp")
		f.SetCellValue(sheetName, "B1", "Event")
		f.SetCellValue(sheetName, "C1", "Activity Name")
		f.SetCellValue(sheetName, "D1", "Duration")
		f.DeleteSheet("Sheet1")
	}

	// Get the last row index
	rows, err := f.GetRows(sheetName)
	if err != nil {
		t := fmt.Sprint("Error getting rows:", err)
		LogEntry.SetText(t)

		return
	}
	rowIndex := len(rows) + 1

	// Write data to the sheet
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIndex), entry.Timestamp)
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIndex), entry.Event)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowIndex), entry.Name)

	if entry.Duration > 0 {
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowIndex),
			fmt.Sprintf("%.1f minutes", entry.Duration.Minutes()))
	}

	// Save the file
	if err := f.Save(); err != nil {
		t := fmt.Sprint("Error saving Excel file:", err)
		handleExcelInUse(err, f)
		LogEntry.SetText(t)
	} else {
		LogEntry.SetText("Data saved successfully.")
		resetLogText()
	}
}

func choinceSwitch(choice int, f *excelize.File) {
	switch choice {
	case 1:
		uniqueFilename := getUniqueFilename()
		if err := f.SaveAs(uniqueFilename); err != nil {
			t := fmt.Sprint("Error saving Excel file:", err)
			LogEntry.SetText(t)
		} else {
			t := fmt.Sprint("Data saved to:", uniqueFilename)
			LogEntry.SetText(t)
			resetLogText()
		}
	case 2:
		if err := saveWithRetry(f, excelFileName, 3); err != nil {
			t := fmt.Sprint("Error saving Excel file:", err)
			LogEntry.SetText(t)
		}
	case 3:
		LogEntry.SetText("Exiting...")
	default:
		LogEntry.SetText("Invalid choice")
	}
}

func handleExcelInUse(err error, f *excelize.File) {
	if f == nil {
		t := "handleExcelInUse(): Excel file is nil"
		LogEntry.SetText(t)
		return
	}

	t := fmt.Sprint("Error saving Excel file:", err)
	t += "\n1. Save to a new file\n2. Retry\n3. Exit"

	w := windowMaker(App, "Command")

	inputField := widget.NewEntry()
	inputField.SetText("Enter your choice: ")
	var choice int

	submitButton := button("Submit", func() {
		if inputField == nil {
			t := "handleExcelInUse(): inputField is nil"
			LogEntry.SetText(t)
			return
		}

		choiceText := inputField.Text
		var err error
		choice, err = strconv.Atoi(choiceText)
		if err != nil {
			t := fmt.Sprint("Invalid choice:", err)
			LogEntry.SetText(t)
			return
		}

		choinceSwitch(choice, f)
		if w == nil {
			t := "handleExcelInUse(): w is nil"
			LogEntry.SetText(t)
			return
		}

		w.Close()
	})

	if w == nil {
		t := "handleExcelInUse(): w is nil"
		LogEntry.SetText(t)
		return
	}

	w.SetContent(container.NewVBox(inputField, submitButton))
	w.Show()

	fmt.Print("Enter your choice: ")
}

func getUniqueFilename() string {
	return fmt.Sprintf("report_%s.xlsx", time.Now().Format("20060102_150405"))
}

func saveWithRetry(f *excelize.File, filename string, maxRetries int) error {
	for range maxRetries {
		if err := f.SaveAs(filename); err == nil {
			return nil // Success
		}
		time.Sleep(1 * time.Second) // Wait before retrying
	}
	return fmt.Errorf("failed to save file after %d retries", maxRetries)
}

func resetLogText() {
	Wg.Add(1)
	go func() {
		defer Wg.Done()
		time.Sleep(6 * time.Second) // Let saveToExcel finish

		LogEntry.SetText("logs:...")
	}()
}
