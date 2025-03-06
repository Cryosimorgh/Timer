package main

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

var Elapsed time.Duration = 0
var WaitDuration = 200 * time.Millisecond

type TimerState struct {
	startTime  time.Time
	pausedTime time.Duration
	running    bool
	paused     bool
	entries    []TimerEntry
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
	App      fyne.App
	Wg       sync.WaitGroup
	LogEntry *widget.Label
)
