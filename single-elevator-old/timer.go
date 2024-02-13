package singleelevator

import "time"

var (
	TimerEndTime time.Time
	TimerActive  bool
)

func TimerStart(duration float64) {
	TimerEndTime = time.Now().Add(time.Duration(duration * float64(time.Second)))
	TimerActive = true
}

func TimerStop() {
	TimerActive = false
}