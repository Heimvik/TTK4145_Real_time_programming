package elevator

import (
	"time"
)

const DOOROPENTIME = 3000 //ms

func F_DoorTimer(chans T_ElevatorChannels){
	for {
		<-chans.C_timerStart
		timer := time.NewTicker(time.Duration(DOOROPENTIME/2) * time.Millisecond) //time to open
		TIMERLOOP:
		for {
			select {
			case <-timer.C:
				chans.C_timerTimeout <- true
			case <-chans.C_timerStop:
				
				timer.Stop()
				break TIMERLOOP
			}
		}
	}
}