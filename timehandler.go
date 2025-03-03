package main

import (
	"time"
)

func updateTimeDisplay() {
	for {
		if state.running && !state.paused {
			elapsed := time.Since(state.startTime)
			timeLabel.SetText(formatDuration(elapsed))
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func startTimer() {
	if !state.running {
		state.startTime = time.Now()
		state.running = true
		state.paused = false
		logEvent("START")
	}
}

func togglePause() {
	if state.running {
		if state.paused {
			state.startTime = time.Now().Add(-state.pausedTime)
			state.paused = false
			logEvent("RESUME")
		} else {
			state.pausedTime = time.Since(state.startTime)
			state.paused = true
			logEvent("PAUSE")
		}
	}
}

func toggleStop() {
	if state.running {
		state.running = false
		state.paused = false
		logEvent("STOP")
		timeLabel.SetText("00:00:00")
		state.pausedTime = 0
	}
}

func logEvent(eventType string) {
	entry := TimerEntry{
		Timestamp: time.Now(),
		Event:     eventType,
		Name:      nameEntry.Text,
	}

	if eventType == "STOP" {
		entry.Duration = time.Since(state.startTime) + state.pausedTime
	} else if eventType == "EXIT" { // New event type for app exit
		entry.Duration = time.Since(state.startTime) + state.pausedTime
	}

	state.entries = append(state.entries, entry)
	saveToExcel(entry)
}
