package main

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	"fyne.io/fyne/v2/widget"
)

func main() {
	// Create app and window
	App = app.New()
	window := windowMaker(App)

	// Initialize UI components
	timeLabel = widget.NewLabel("00:00:00")
	nameEntry = widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter activity name")

	// Create buttons
	startButton := button("Start", startTimer)
	pauseButton := button("Pause", pauseTimer)
	stopButton := button("Stop", stopTimer)
	exitButton := button("Exit", exitApp)
	exitButton.Importance = widget.HighImportance

	// Make Draggable
	draggableHeader := makeDraggable(window)
	// Create button container
	buttonContainer := container.NewHBox(
		startButton,
		pauseButton,
		stopButton,
	)
	// Create layout
	content := container.NewVBox(
		draggableHeader,
		nameEntry,
		timeLabel,
		buttonContainer,
		exitButton,
	)

	// Initialize Excel file
	initExcelFile()

	// Start update loop
	go updateTimeDisplay()

	// Show and run app
	window.SetContent(content)
	window.ShowAndRun()
}

func formatDuration(d time.Duration) string {
	total := int(d.Seconds())
	hours := total / 3600
	minutes := (total % 3600) / 60
	seconds := total % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}
