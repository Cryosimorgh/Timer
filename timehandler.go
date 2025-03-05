package main

import (
	"time"
)

func updateTimeDisplay() {
	for {
		if state.running && !state.paused {
			Elapsed += WaitDuration
			timeLabel.SetText(formatDuration(Elapsed))
		}
		time.Sleep(WaitDuration)
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
			state.paused = false
			logEvent("RESUME")
		} else {
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
		Elapsed = 0
	}
}

func logEvent(eventType string) {
	entry := TimerEntry{
		Timestamp: time.Now(),
		Event:     eventType,
		Name:      nameEntry.Text,
	}

	if eventType == "STOP" {
		entry.Duration = Elapsed
	} else if eventType == "EXIT" {
		entry.Duration = Elapsed
	}

	state.entries = append(state.entries, entry)
	saveToExcel(entry)
}
