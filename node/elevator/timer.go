package elevator

import (
	"time"
)

const DOOROPENTIME = 3000 //ms

func F_DoorTimer(chans T_ElevatorChannels){
	timer := time.NewTimer(0) //time to open
	<-timer.C
	for {
		select {
		case <-chans.C_timerStart:
			timer.Reset(time.Duration(DOOROPENTIME) * time.Millisecond)
		case <-timer.C:
			chans.C_timerTimeout <- true
		case <-chans.C_timerStop:
			timer.Stop()	
		}
	}
}
