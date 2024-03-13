package elevator

import (
	"time"
)

const DOOROPEN_TIME = 3000 //ms

/*
Manages the door timer, allowing for starting, stopping, and timing out door open events.

Prerequisites: None

Returns: Nothing, but sends a timeout signal to the c_timerTimeout channel when the door timer times out.
*/
func F_DoorTimer(chans T_ElevatorChannels) {
	timer := time.NewTimer(0)
	<-timer.C
	for {
		select {
		case <-chans.C_timerStart:
			timer.Reset(time.Duration(DOOROPEN_TIME) * time.Millisecond)
		case <-timer.C:
			chans.C_timerTimeout <- true
		case <-chans.C_timerStop:
			timer.Stop()
		}
	}
}
