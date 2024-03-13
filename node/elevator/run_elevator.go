package elevator

import (
	"fmt"
)

//***END TEST FUNCTIONS***

/*
10.03.2024
TODO:
- Fikse alt av lampegreier (ÆSJ!!!)
- Fjerne unødvendige variabler og funksjoner (ongoing)
- Rydde opp i griseriet. Fjerne unødvendige kommentarer og kode. (going pretty good)
*/

//kan kanskje flyttes men foreløpig kan den bli
/*
func F_AcknowledgeRequests(elevatorOperations T_ElevatorOperations, chans T_ElevatorChannels) {
	previousElevator := T_Elevator{}
	for {
		currentElevator := F_GetElevator(elevatorOperations)
		if currentElevator.P_serveRequest != nil {
			if currentElevator.P_serveRequest.State == ACTIVE && previousElevator.P_serveRequest == nil {
				chans.C_requestOut <- *currentElevator.P_serveRequest
			} else if currentElevator.P_serveRequest.State == DONE && previousElevator.P_serveRequest.State == ACTIVE {
				chans.C_requestOut <- *currentElevator.P_serveRequest
			}
		}
		previousElevator = currentElevator
	}
}*/

func F_RunElevator(elevatorOperations T_ElevatorOperations, c_getSetElevatorInterface chan T_GetSetElevatorInterface, c_requestOut chan T_Request, c_requestIn chan T_Request, elevatorport int, c_elevatorWithoutErrors chan bool) {

	F_InitDriver(fmt.Sprintf("localhost:%d", elevatorport))

	F_SetMotorDirection(DOWN)

	var chans T_ElevatorChannels = F_InitChannes(c_requestIn, c_requestOut)
	//interface for getting and setting elevator
	go F_GetAndSetElevator(elevatorOperations, c_getSetElevatorInterface)
	//polling sensors
	go F_PollButtons(chans.C_buttons)
	go F_PollFloorSensor(chans.C_floors)
	go F_PollObstructionSwitch(chans.C_obstr)
	go F_PollStopButton(chans.C_stop)
	//doortimer
	go F_DoorTimer(chans)
	//FSM
	go F_FSM(c_getSetElevatorInterface, chans, c_elevatorWithoutErrors)

	//go F_AcknowledgeRequests(elevatorOperations, chans)
}

// Kommentarer kodekvalitet:
// - La alle ganger du skriver til c_out og c_in være lesbare her, og ikke pakk det inn i funksjon (ta ut receiveRequest og sendRequest)
// - Lag en sentral FSM, ikke fordelt på mange funksjoner, som switcher på elevator.state, hvor alt som skal
// 	skje i IDLE, skjer i IDLE casen, alt som skal skje i MOVING skjer i moving casen osv. SÅ heller sende den
// 	til og fra forskjellige states her ute
// - Prøv å generaliser (krymp) if-statements, evt lag en funksjon med conditions hvis det er nødt til å være veldig mye
// - f_StorbokstavStorbokstav i funksjoner
// - andre navn en "a" i caser
