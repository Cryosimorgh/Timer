package main

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	"fyne.io/fyne/v2/widget"
)

func main() {
	Wg = sync.WaitGroup{}
	// Create app and window
	App = app.New()
	App.SetIcon(ResourceIconPng)
	window := windowMaker(App, "Time Tracker")
	window.SetMaster()

	// Initialize UI components
	timeLabel = widget.NewLabel("00:00:00")
	timeLabel.Alignment = fyne.TextAlignCenter

	nameEntry = widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter activity name")

	LogEntry = widget.NewLabel("Logs:...")
	LogEntry.Wrapping = fyne.TextWrapWord
	LogEntry.Alignment = fyne.TextAlignCenter
	LogEntry.TextStyle.Italic = true
	LogEntry.TextStyle.Bold = true

	// Create buttons
	startButton := button("Start", startTimer)
	pauseButton := button("Pause", pauseTimer)
	stopButton := button("Stop", stopTimer)
	exitButton := button("Exit", exitApp)
	exitButton.Importance = widget.HighImportance

	buttonContainer := container.NewCenter(
		container.NewHBox(
			startButton,
			pauseButton,
			stopButton,
		),
	)

	// Create layout
	content := container.NewVBox(
		//draggableHeader,
		nameEntry,
		timeLabel,
		buttonContainer,
		exitButton,
		LogEntry,
	)

	// Initialize Excel file
	initExcelFile()

	// Start update loop
	go updateTimeDisplay()

	// Show and run app
	window.SetContent(content)
	window.SetCloseIntercept(func() {
		logEvent("EXIT")
		Wg.Add(1)
		LogEntry.SetText("Saving to Excel...")
		// Force immediate Excel save before closing
		go func() {
			time.Sleep(100 * time.Millisecond) // Let saveToExcel finish
			
			App.Quit()
		}()
	})

	window.ShowAndRun()

}

func formatDuration(d time.Duration) string {
	total := int(d.Seconds())
	hours := total / 3600
	minutes := (total % 3600) / 60
	seconds := total % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}
