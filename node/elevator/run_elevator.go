package elevator



/*
29.02.2024
TODO:
- Legge til elevatormusic.
- Rydde opp i griseriet. Fjerne unødvendige kommentarer og kode.
- Fjerne unødvendige variabler og funksjoner.
- Fikse alt av lampegreier.
*/

var C_stop bool
var C_obstruction bool
var ID int //temp, spør arbo om flytting
var C_timerStart chan bool
var DOOROPENTIME int = 3

func F_RunElevator(ops T_ElevatorOperations, c_requestOut chan T_Request, c_requestIn chan T_Request) {

	Init("localhost:15657") //henter port fra config elno, må smelle på localhost sjæl tror jeg

	SetMotorDirection(MD_Down)

	C_stop = false
	C_obstruction = false

	c_readElevator := make(chan T_Elevator)
	c_writeElevator := make(chan T_Elevator)
	c_quitGetSetElevator := make(chan bool)

	go F_GetAndSetElevator(ops, c_readElevator, c_writeElevator, c_quitGetSetElevator)
	//initialize elevator
	oldElevator := <- c_readElevator
	oldElevator = Init_Elevator(c_requestIn, c_requestOut)
	oldElevator.P_info.State = IDLE
	c_writeElevator <- oldElevator
	c_quitGetSetElevator <- true

	C_timerStart = make(chan bool)
	drv_buttons := make(chan ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go PollButtons(drv_buttons)
	go PollFloorSensor(drv_floors)
	go PollObstructionSwitch(drv_obstr)
	go PollStopButton(drv_stop)
	go F_Timer(C_timerStart, c_readElevator, c_writeElevator, c_quitGetSetElevator)

	for {
		select {
		case a := <-drv_buttons:
			oldElevator := F_GetElevator(ops)
			F_sendRequest(a, oldElevator.C_distributeRequest, oldElevator)

		case a := <-drv_floors:
			oldElevator := <- c_readElevator
			newElevator := F_fsmFloorArrival(int8(a), oldElevator)
			c_writeElevator <- newElevator
			c_quitGetSetElevator <- true

		case a := <-drv_obstr:
			C_obstruction = a

		case a := <-drv_stop:
			C_stop = a
			oldElevator := <- c_readElevator
			newElevator := F_chooseDirection(oldElevator)
			c_writeElevator <- newElevator
			c_quitGetSetElevator <- true

		//dette er det jeg ringte deg om
		case a := <- c_requestIn:
		// case a := <-Elevator.C_receiveRequest:
			oldElevator := <- c_readElevator
			newElevator := F_ReceiveRequest(a, oldElevator)
			F_chooseDirection(newElevator)
			c_writeElevator <- newElevator
			c_quitGetSetElevator <- true
		}
	}
}