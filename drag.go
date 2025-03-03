package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type draggableHeader struct {
	*widget.Label
	window fyne.Window
	start  fyne.Position
}

func newDraggableHeader(w fyne.Window) *draggableHeader {
	header := &draggableHeader{
		Label:  widget.NewLabel("Drag here to move"),
		window: w,
	}
	header.Alignment = fyne.TextAlignCenter
	header.ExtendBaseWidget(header)
	return header
}

func (h *draggableHeader) MouseDown(e *desktop.MouseEvent) {
	h.start = e.AbsolutePosition
	log.Printf("MouseDown")
}

func (h *draggableHeader) MouseUp(*desktop.MouseEvent) {
	h.start = fyne.NewPos(0, 0)
	log.Printf("MouseUp")
}

func (h *draggableHeader) Dragged(e *desktop.MouseEvent) {
	log.Printf("Dragged")
	if h.start.IsZero() {
		return
	}

	delta := e.AbsolutePosition.Subtract(h.start)
	currentPos := h.window.Canvas().Content().Position()
	h.window.Content().Move(currentPos.Add(delta))
	h.start = e.AbsolutePosition
}
