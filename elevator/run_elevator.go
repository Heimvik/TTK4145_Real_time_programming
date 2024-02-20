package elevator

import (
	"fmt"
	"the-elevator/node"
)

var Elevator T_Elevator
var C_stop bool
var C_obstruction bool

func F_RunElevator(ops node.T_NodeOperation, c_requestIn chan T_Request, c_requestOut chan T_Request) {

	Init("localhost:15657") //henter port fra config elno, må smelle på localhost sjæl tror jeg
	Elevator = Init_Elevator(c_requestIn, c_requestOut)

	SetMotorDirection(MD_Down)

    C_stop = false
    C_obstruction = false

	drv_buttons := make(chan ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	//drv_req := make(chan T_request)

	go PollButtons(drv_buttons)
	go PollFloorSensor(drv_floors)
	go PollObstructionSwitch(drv_obstr)
	go PollStopButton(drv_stop)
	//go PollNewRequest(c_requestOut)

	for {
		select {
		case a := <-drv_buttons:
			F_sendRequest(a, Elevator.C_distributeRequest)

		case a := <-drv_floors:
			F_FloorArrival(a)

		case a := <-drv_obstr: //tipper dette er nok til å kun teste funksjonalitet
			C_obstruction = a
            //legg til F_fsmObstructionSwitch(a)
            //den skal ikke gjøre noe hvis heisen ikke er i DOOROPEN
            //hvis heisen er i DOOROPEN, skal den sette heisen i IDLE og lukke døra hvis C_obstruction er false, hvis ikke skal den ikke gjøre noe

		case a := <-drv_stop:
			C_stop = a //legg til i F_chooseDirection at hvis C_stop er true, så skal den stoppe
            F_chooseDirection(Elevator)

		case a := <-Elevator.C_receiveRequest:
			fmt.Printf("Elevator: Mottok request fra node\n")
			Elevator.P_serveRequest = &a
			F_chooseDirection(Elevator)
			// SetMotorDirection(MotorDirection(Elevator.P_info.Direction))

		}
	}
}
