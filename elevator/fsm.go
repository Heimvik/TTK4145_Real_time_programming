package elevator

import (
	"time"
)
func F_FloorArrival(newFloor int) {

	Elevator.P_info.Floor = newFloor
	// SetFloorIndicator(newFloor)

	switch Elevator.P_info.State {
	case MOVING:
		if F_shouldStop(Elevator){
			SetMotorDirection(MD_Stop)
			Elevator.P_info.State = DOOROPEN
			//make timer logic so door stays open for as long as it should
			time.Sleep(3 * time.Second) //placeholder
			F_clearRequest(Elevator)
		}
	case DOOROPEN:
		//make timer logic so door stays open for as long as it should

	}

}


func F_sendRequest(button ButtonEvent, requestOut chan T_Request) {
	if button.Button == BT_Cab {
		requestOut <- T_Request{Calltype: CAB, Floor: button.Floor}
		return
	} else {
		requestOut <- T_Request{Calltype: HALL, Floor: button.Floor}
		return
	}
}
