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

// func F_Timer(ops T_ElevatorOperations, timerStart chan bool, c_readElevator chan T_Elevator, c_writeElevator chan T_Elevator, c_quitGetSetElevator chan bool) {
// 	for {
// 		<-timerStart
// 		timer := time.NewTicker(time.Duration(DOOROPENTIME) * time.Second)
// 		for range timer.C {
// 			go F_GetAndSetElevator(ops, c_readElevator, c_writeElevator, c_quitGetSetElevator)
// 			oldElevator := <-c_readElevator
// 			if oldElevator.P_info.State == DOOROPEN && !C_obstruction { //hvis heisen ikke er obstructed skal den gå til IDLE 
// 				oldElevator.P_info.State = IDLE
// 				c_writeElevator <- oldElevator
// 				c_quitGetSetElevator <- true
// 				timer.Stop()
// 				break
// 			} else if oldElevator.P_info.State == DOOROPEN && C_obstruction { //hvis heisen er obstructed skal den fortsette å være DOOROPEN
// 				if oldElevator.P_serveRequest != nil { //hvis den i tillegg har en request den skal serve må denne resendes til node
// 					resendReq := *oldElevator.P_serveRequest
// 					resendReq.State = UNASSIGNED
// 					oldElevator.C_distributeRequest <- resendReq
// 				}
// 				c_writeElevator <- oldElevator
// 				c_quitGetSetElevator <- true
// 			}
// 		}
// 	}
// }