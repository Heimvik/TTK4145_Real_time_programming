package elevator

import (
	"time"
)

func F_Timer(start chan bool, stop chan bool) {
	for {
		select {
		case <-start:
			time.Sleep(3 * time.Second)
			stop <- true
		}
	}
}

func F_StopListener(stop chan bool) {
	for {
		select {
		case <-stop:
			F_fsmDoorTimeout()
		}
	}
}