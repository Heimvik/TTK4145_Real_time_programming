package elevator

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var getSetElevatorInterface T_GetSetElevatorInterface        //Delete when no testing is needed, define only in run elevator
var c_getSetElevatorInterface chan T_GetSetElevatorInterface //Delete when no testing is needed, define only in run elevator

func F_SimulateRequest(c_requestFromElevator chan T_Request, c_requestToElevator chan T_Request) {
	getSetElevatorInterface = T_GetSetElevatorInterface{ //make local when node Dont need acces with f_SimulateRequest() no more
		C_get: make(chan T_Elevator),
		C_set: make(chan T_Elevator),
	}
	c_getSetElevatorInterface = make(chan T_GetSetElevatorInterface) //make local when node Dont need acces with f_SimulateRequest() no more

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
- Brief heimvik på initialisering av heis (fjerna channels, la til ID, obstructed og stop variabler, start på floor -1)
- Fikse alt av lampegreier (ÆSJ!!!)
- Finn ut av hva som skal skje med ID
- Fjerne unødvendige variabler og funksjoner (ongoing)
- Rydde opp i griseriet. Fjerne unødvendige kommentarer og kode. (going pretty good)
- Legge til elevatormusic.
*/

var DOOROPENTIME int = 3                                     //kan kanskje flyttes men foreløpig kan den bli

func F_RunElevator(ops T_ElevatorOperations, c_requestOut chan T_Request, c_requestIn chan T_Request, elevatorport int) {
	getSetElevatorInterface = T_GetSetElevatorInterface{ //make local when node Dont need acces with f_SimulateRequest() no more
		C_get: make(chan T_Elevator),
		C_set: make(chan T_Elevator),
	}

	c_getSetElevatorInterface = make(chan T_GetSetElevatorInterface) //make local when node Dont need acces with f_SimulateRequest() no more

	F_InitDriver(fmt.Sprintf("localhost:%d", elevatorport))

	F_SetMotorDirection(DOWN)

	//channels
	//c_timerStop := make(chan bool)
	//c_timerTimeout := make(chan bool) //COMMENT: Navn
	c_buttons := make(chan T_ButtonEvent)
	c_floors := make(chan int)
	c_obstr := make(chan bool)
	c_stop := make(chan bool)

	go F_PollButtons(c_buttons)
	go F_PollFloorSensor(c_floors)
	go F_PollObstructionSwitch(c_obstr)
	go F_PollStopButton(c_stop)

	/*
			Kommentarer kodekvalitet:
			- La alle ganger du skriver til c_out og c_in være lesbare her, og ikke pakk det inn i funksjon (ta ut receiveRequest og sendRequest)
			- Lag en sentral FSM, ikke fordelt på mange funksjoner, som switcher på elevator.state, hvor alt som skal
				skje i IDLE, skjer i IDLE casen, alt som skal skje i MOVING skjer i moving casen osv. SÅ heller sende den
				til og fra forskjellige states her ute
			- Prøv å generaliser (krymp) if-statements, evt lag en funksjon med conditions hvis det er nødt til å være veldig mye
			- f_StorbokstavStorbokstav i funksjoner
			- andre navn en "a" i caser

		for {
			select {
			case a := <-c_buttons:
				C_getSetElevatorInterface <- GetSetElevatorInterface
				oldElevator := <-GetSetElevatorInterface.C_get
				oldElevator.CurrentID++
				GetSetElevatorInterface.C_set <- oldElevator
				F_sendRequest(a, c_requestOut, oldElevator) //COMMENT: legg ut

			case a := <-c_floors: //COMMENT: a
				go F_GetAndSetElevator(ops, c_readElevator, c_writeElevator, c_quitGetSetElevator)
				oldElevator := <-c_readElevator
				newElevator := F_fsmFloorArrival(int8(a), oldElevator)

				c_writeElevator <- newElevator
				c_quitGetSetElevator <- true
				if newElevator.P_info.State == DOOROPEN {
					go F_Timer(c_timerStop, c_timerTimeout) //COMMENT: Hva slags timer? Hva timer den?
				}
				if newElevator.P_info.Direction == NONE && !newElevator.StopButton && oldElevator.P_serveRequest != nil {
					fmt.Println(strconv.Itoa(int(oldElevator.P_serveRequest.Floor)) + " | " + strconv.Itoa(int(a)))
					oldElevator.P_serveRequest.State = DONE
					c_requestOut <- *oldElevator.P_serveRequest
				}

			case a := <-c_obstr: //COMMENT: a
				go F_GetAndSetElevator(ops, c_readElevator, c_writeElevator, c_quitGetSetElevator)
				oldElevator := <-c_readElevator
				oldElevator.Obstructed = a
				c_writeElevator <- oldElevator
				c_quitGetSetElevator <- true

			case a := <-c_stop: //COMMENT: a
				go F_GetAndSetElevator(ops, c_readElevator, c_writeElevator, c_quitGetSetElevator)
				oldElevator := <-c_readElevator
				oldElevator.StopButton = a
				newElevator := F_SetElevatorDirection(oldElevator)
				c_writeElevator <- newElevator
				c_quitGetSetElevator <- true
				if newElevator.P_info.State == DOOROPEN {
					go F_Timer(c_timerStop, c_timerTimeout) //COMMENT: samme her
				}
				if newElevator.P_info.Direction == NONE && !newElevator.StopButton && oldElevator.P_serveRequest != nil {
					oldElevator.P_serveRequest.State = DONE
					c_requestOut <- *oldElevator.P_serveRequest
				}

			case <-c_timerTimeout: //COMMENT: timerTimeout? hva er timeren på?
				go F_GetAndSetElevator(ops, c_readElevator, c_writeElevator, c_quitGetSetElevator)
				oldElevator := <-c_readElevator
				newElevator, newReq := F_fsmDoorTimeout(oldElevator, c_requestOut) //COMMENT: Tydeliggjør hva den skal returnere gjennom funksjonsnavnet gjør selve endringen på elevator på utsiden?
				c_writeElevator <- newElevator
				c_quitGetSetElevator <- true
				if newReq.State == UNASSIGNED && newElevator.P_serveRequest != nil {
					c_requestOut <- newReq
				} else if newElevator.P_info.State == IDLE {
					c_timerStop <- true
				}

			case a := <-c_requestIn:
				go F_GetAndSetElevator(ops, c_readElevator, c_writeElevator, c_quitGetSetElevator)
				oldElevator := <-c_readElevator
				newElevator := F_ReceiveRequest(a, oldElevator, c_requestOut) //COMMENT:ReceiveRequest og returnerer en elevator, og tar inn requestOut?
				c_writeElevator <- newElevator
				c_quitGetSetElevator <- true
				if newElevator.P_info.State == DOOROPEN {
					go F_Timer(c_timerStop, c_timerTimeout) //COMMENT: Timer
				} else if newElevator.P_info.State == MOVING {
					a.State = ACTIVE
					c_requestOut <- a
				}
			}
		}*/
}
