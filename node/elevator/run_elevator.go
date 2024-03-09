package elevator

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func F_SimulateRequest(elevatorOperations T_ElevatorOperations, c_requestFromElevator chan T_Request, c_requestToElevator chan T_Request) {
	c_getSetElevatorInterface := make(chan T_GetSetElevatorInterface)
	getSetElevatorInterface := T_GetSetElevatorInterface{
		C_get: make(chan T_Elevator),
		C_set: make(chan T_Elevator),
	}

	go F_GetAndSetElevator(elevatorOperations, c_getSetElevatorInterface)

	increment := uint16(0)
	go func() {
		for {
			select {
			case request := <-c_requestToElevator:
				c_getSetElevatorInterface <- getSetElevatorInterface
				currentElevator := <-getSetElevatorInterface.C_get
				(*currentElevator.P_info).State = MOVING
				getSetElevatorInterface.C_set <- currentElevator

				request.State = ACTIVE
				c_requestFromElevator <- request

				time.Sleep(10 * time.Second)

				c_getSetElevatorInterface <- getSetElevatorInterface
				newElevator := <-getSetElevatorInterface.C_get
				(*newElevator.P_info).State = IDLE
				getSetElevatorInterface.C_set <- newElevator

				request.State = DONE
				c_requestFromElevator <- request
			default:
				time.Sleep(time.Duration(5000) * time.Microsecond)
			}
		}
	}()

	for {
		var input string
		fmt.Println("Enter request (C/H-floor): ")
		fmt.Scanln(&input)
		delimiter := "-"
		parts := strings.Split(input, delimiter)
		partToConvert := parts[1]
		floor, _ := strconv.Atoi(partToConvert)
		var returnRequest T_Request
		if parts[0] == "C" {
			returnRequest = T_Request{
				Id:        increment,
				State:     UNASSIGNED,
				Calltype:  CAB,
				Floor:     int8(floor),
				Direction: UP,
			}
			increment += 1
			c_requestFromElevator <- returnRequest
		} else if parts[0] == "H" {
			returnRequest = T_Request{
				Id:        increment,
				State:     UNASSIGNED,
				Calltype:  HALL,
				Floor:     int8(floor),
				Direction: UP,
			}
			increment += 1
			c_requestFromElevator <- returnRequest
		}
		time.Sleep(time.Duration(5000) * time.Microsecond)
	}
}

//***END TEST FUNCTIONS***

/*
05.03.2024
TODO:
- Fikse alt av lampegreier (ÆSJ!!!)
- Finn ut av hva som skal skje med ID
- Fjerne unødvendige variabler og funksjoner (ongoing)
- Rydde opp i griseriet. Fjerne unødvendige kommentarer og kode. (going pretty good)
- Legge til elevatormusic.
*/

var DOOROPENTIME int = 3 //kan kanskje flyttes men foreløpig kan den bli

func F_RunElevator(elevatorOperations T_ElevatorOperations, c_getSetElevatorInterface chan T_GetSetElevatorInterface,c_requestOut chan T_Request, c_requestIn chan T_Request, elevatorport int) {

	F_InitDriver(fmt.Sprintf("localhost:%d", elevatorport))

	F_SetMotorDirection(DOWN)

	getSetElevatorInterface := T_GetSetElevatorInterface{
		C_get: make(chan T_Elevator),
		C_set: make(chan T_Elevator),
	}
	//channels
	c_timerStop := make(chan bool)
	c_timerTimeout := make(chan bool) //COMMENT: Navn
	c_buttons := make(chan T_ButtonEvent)
	c_floors := make(chan int)
	c_obstr := make(chan bool)
	c_stop := make(chan bool)

	go F_PollButtons(c_buttons)
	go F_PollFloorSensor(c_floors)
	go F_PollObstructionSwitch(c_obstr)
	go F_PollStopButton(c_stop)
	go F_GetAndSetElevator(elevatorOperations, c_getSetElevatorInterface)

	
			// Kommentarer kodekvalitet:
			// - La alle ganger du skriver til c_out og c_in være lesbare her, og ikke pakk det inn i funksjon (ta ut receiveRequest og sendRequest)
			// - Lag en sentral FSM, ikke fordelt på mange funksjoner, som switcher på elevator.state, hvor alt som skal
			// 	skje i IDLE, skjer i IDLE casen, alt som skal skje i MOVING skjer i moving casen osv. SÅ heller sende den
			// 	til og fra forskjellige states her ute
			// - Prøv å generaliser (krymp) if-statements, evt lag en funksjon med conditions hvis det er nødt til å være veldig mye
			// - f_StorbokstavStorbokstav i funksjoner
			// - andre navn en "a" i caser

	for {
		select {
		case button := <-c_buttons:
			c_getSetElevatorInterface <- getSetElevatorInterface
			oldElevator := <-getSetElevatorInterface.C_get
			oldElevator.CurrentID++
			getSetElevatorInterface.C_set <- oldElevator

			F_SendRequest(button, c_requestOut, oldElevator) //COMMENT: legg ut

		case newFloor := <-c_floors:
			c_getSetElevatorInterface <- getSetElevatorInterface
			oldElevator := <-getSetElevatorInterface.C_get
			newElevator := F_FloorArrival(int8(newFloor), oldElevator)
			getSetElevatorInterface.C_set <- newElevator
			
			if newElevator.P_info.State == DOOROPEN {
				go F_DoorTimer(c_timerStop, c_timerTimeout)
			}
			if newElevator.P_info.Direction == NONE && !newElevator.StopButton && oldElevator.P_serveRequest != nil { //kan nok forenkle logikken her
				// fmt.Println(strconv.Itoa(int(oldElevator.P_serveRequest.Floor)) + " | " + strconv.Itoa(int(newFloor)))
				oldElevator.P_serveRequest.State = DONE
				c_requestOut <- *oldElevator.P_serveRequest
			}

		case obstructed := <-c_obstr:
			c_getSetElevatorInterface <- getSetElevatorInterface
			oldElevator := <-getSetElevatorInterface.C_get

			oldElevator.Obstructed = obstructed

			getSetElevatorInterface.C_set <- oldElevator
			
		case stop := <-c_stop: //COMMENT: a
			c_getSetElevatorInterface <- getSetElevatorInterface
			oldElevator := <-getSetElevatorInterface.C_get

			oldElevator.StopButton = stop
			newElevator := F_SetElevatorDirection(oldElevator)

			getSetElevatorInterface.C_set <- newElevator

		case <-c_timerTimeout: //COMMENT: timerTimeout? hva er timeren på?
			c_getSetElevatorInterface <- getSetElevatorInterface
			oldElevator := <-getSetElevatorInterface.C_get
			newElevator, newReq := F_DoorTimeout(oldElevator, c_requestOut) //COMMENT: Tydeliggjør hva den skal returnere gjennom funksjonsnavnet gjør selve endringen på elevator på utsiden?
			getSetElevatorInterface.C_set <- newElevator

			if newReq.State == UNASSIGNED && newElevator.P_serveRequest != nil {
				c_requestOut <- newReq
			} else if newElevator.P_info.State == IDLE {
				c_timerStop <- true
			}

		case newRequest := <-c_requestIn:
			c_getSetElevatorInterface <- getSetElevatorInterface
			oldElevator := <-getSetElevatorInterface.C_get
			newElevator := F_ReceiveRequest(newRequest, oldElevator) //COMMENT:ReceiveRequest og returnerer en elevator, og tar inn requestOut?
			getSetElevatorInterface.C_set <- newElevator

			if newElevator.P_info.State == DOOROPEN {
				newRequest.State = ACTIVE
				c_requestOut <- newRequest
				newRequest.State = DONE
				c_requestOut <- newRequest
				go F_DoorTimer(c_timerStop, c_timerTimeout)
			} else if newElevator.P_info.State == MOVING {
				c_requestOut <- *newElevator.P_serveRequest
			}
		}
	}
}
