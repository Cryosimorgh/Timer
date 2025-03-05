package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func windowMaker(myApp fyne.App) fyne.Window {

	window := myApp.NewWindow("Time Tracker")
	window.Resize(fyne.NewSize(400, 300))
	window.SetFixedSize(true)

	// Create draggable header
	//header := newDraggableHeader(window)

	// Create close button
	closeButton := widget.NewButton("X", func() {
		myApp.Quit()
	})
	closeButton.Importance = widget.HighImportance

	content := container.NewBorder(
		nil, nil, nil, closeButton,
		widget.NewLabel("Time Tracker"),
	)

	window.SetContent(content)
	return window
}

// func makeDraggable(window fyne.Window) fyne.CanvasObject {
// 	return newDraggableHeader(window)
// }
