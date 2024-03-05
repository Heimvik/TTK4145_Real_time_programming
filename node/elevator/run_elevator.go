package elevator

import (
	"fmt"
)

/*
05.03.2024
TODO:
- Brief heimvik på initialisering av heis (fjerna channels, la til ID, obstructed og stop variabler, start på floor -1)
- Fikse alt av lampegreier (ÆSJ!!!)
- Finn ut av hva som skal skje med ID
- Endre elevio til å bruke egendefinerte typer
- Fjerne unødvendige variabler og funksjoner (ongoing)
- Rydde opp i griseriet. Fjerne unødvendige kommentarer og kode.
- Legge til elevatormusic.
*/



var DOOROPENTIME int = 3 //kan kanskje flyttes men foreløpig kan den bli

func F_RunElevator(ops T_ElevatorOperations, c_requestOut chan T_Request, c_requestIn chan T_Request, elevatorport int) {
	Init(fmt.Sprintf("localhost:%d", elevatorport))

	SetMotorDirection(MD_Down)

	c_readElevator := make(chan T_Elevator)
	c_writeElevator := make(chan T_Elevator)
	c_quitGetSetElevator := make(chan bool)

	//might delete if elevator is initialized outside of this function *
	go F_GetAndSetElevator(ops, c_readElevator, c_writeElevator, c_quitGetSetElevator)
	//initialize elevator
	oldElevator := <-c_readElevator
	oldElevator = Init_Elevator() //mulig denne droppes
	oldElevator.P_info.State = IDLE
	c_writeElevator <- oldElevator
	c_quitGetSetElevator <- true
	// *

	c_timerStop := make(chan bool)
	c_timerTimeout := make(chan bool)
	drv_buttons := make(chan ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go PollButtons(drv_buttons)
	go PollFloorSensor(drv_floors)
	go PollObstructionSwitch(drv_obstr)
	go PollStopButton(drv_stop)

	for {
		select {
		case a := <-drv_buttons:
			go F_GetAndSetElevator(ops, c_readElevator, c_writeElevator, c_quitGetSetElevator)
			oldElevator := <-c_readElevator
			newElevator := F_sendRequest(a, c_requestOut, oldElevator)
			c_writeElevator <- newElevator
			c_quitGetSetElevator <- true

		case a := <-drv_floors:
			go F_GetAndSetElevator(ops, c_readElevator, c_writeElevator, c_quitGetSetElevator)
			oldElevator := <-c_readElevator
			newElevator := F_fsmFloorArrival(int8(a), oldElevator, c_requestOut)
			if newElevator.P_info.State == IDLE {
				go F_Timer(c_timerStop, c_timerTimeout)
			}
			c_writeElevator <- newElevator
			c_quitGetSetElevator <- true

		case a := <-drv_obstr:
			go F_GetAndSetElevator(ops, c_readElevator, c_writeElevator, c_quitGetSetElevator)
			oldElevator := <-c_readElevator
			oldElevator.Obstructed = a
			c_writeElevator <- oldElevator
			c_quitGetSetElevator <- true

		case a := <-drv_stop:
			go F_GetAndSetElevator(ops, c_readElevator, c_writeElevator, c_quitGetSetElevator)
			oldElevator := <-c_readElevator
			oldElevator.StopButton = a
			newElevator := F_chooseDirection(oldElevator, c_requestOut)
			c_writeElevator <- newElevator
			c_quitGetSetElevator <- true

		case <-c_timerTimeout:
			go F_GetAndSetElevator(ops, c_readElevator, c_writeElevator, c_quitGetSetElevator)
			oldElevator := <-c_readElevator
			newElevator := F_fsmDoorTimeout(oldElevator, c_requestOut)
			c_writeElevator <- newElevator
			c_quitGetSetElevator <- true

		case a := <-c_requestIn:
			go F_GetAndSetElevator(ops, c_readElevator, c_writeElevator, c_quitGetSetElevator)
			oldElevator := <-c_readElevator
			newElevator := F_ReceiveRequest(a, oldElevator, c_requestOut)
			newElevator = F_chooseDirection(newElevator, c_requestOut)
			c_writeElevator <- newElevator
			c_quitGetSetElevator <- true
		}
	}
}
