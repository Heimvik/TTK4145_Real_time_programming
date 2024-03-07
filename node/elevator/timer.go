package elevator

import (
	"time"
)

func F_DoorTimer(timerStop chan bool, timerTimeout chan bool){
	timer := time.NewTicker(time.Duration(DOOROPENTIME) * time.Second)
	for {
		select {
		case <-timer.C:
			timerTimeout <- true
		case <-timerStop:
			timer.Stop()
			return
		}
	}
}