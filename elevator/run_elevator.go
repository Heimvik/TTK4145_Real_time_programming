package elevator

import "fmt"

var Elevator T_Elevator

func F_RunElevator(requestIn chan T_Request, requestOut chan T_Request) {

	Init("localhost:15657")
	Elevator = Init_Elevator(requestIn, requestOut)

	SetMotorDirection(MD_Up)

	drv_buttons := make(chan ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	//drv_req := make(chan T_request)

	go PollButtons(drv_buttons)
	go PollFloorSensor(drv_floors)
	go PollObstructionSwitch(drv_obstr)
	go PollStopButton(drv_stop)
	//go PollNewRequest(requestOut)

	for {
		select {
		case a := <-drv_buttons:
			F_sendRequest(a, Elevator.C_distributeRequest)

		case a := <-drv_floors:
			F_FloorArrival(a)

		case a := <-drv_obstr: //tipper dette er nok til å kun teste funksjonalitet
			if a {
				SetMotorDirection(MD_Stop)
			} else {
				F_chooseDirection(Elevator)
				SetMotorDirection(MotorDirection(Elevator.P_info.Direction))
			}

		case a := <-drv_stop: //vet egt ikke hva denne gjør, men lar den stå for nå
			fmt.Printf("%v+\n", a)

		case a := <-Elevator.C_receiveRequest:
			Elevator.P_serveRequest = &a
			F_chooseDirection(Elevator)
			SetMotorDirection(MotorDirection(Elevator.P_info.Direction))

		}
	}
}
