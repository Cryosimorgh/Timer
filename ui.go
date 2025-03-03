package main

import "fyne.io/fyne/v2/widget"

func button(name string, action func()) *widget.Button {
	return widget.NewButton(name, action)
}
