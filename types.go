package main

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type TimerState struct {
	startTime   time.Time
	pausedTime  time.Duration
	running     bool
	paused      bool
	entries     []TimerEntry
	currentName string
}

type TimerEntry struct {
	Timestamp time.Time
	Event     string
	Name      string
	Duration  time.Duration
}

var (
	state         TimerState
	timeLabel     *widget.Label
	nameEntry     *widget.Entry
	excelFileName = "time_tracker.xlsx"
)

var (
	App fyne.App
)
