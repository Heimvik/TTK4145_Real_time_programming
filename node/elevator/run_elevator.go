package elevator

import (
	"fmt"
	"the-elevator/node"
)

/*
29.02.2024
TODO:
- Ferdigutvikle timer.go, legge til channels som sier når timer skal starte. Opprette global variabel som sier hvor lenge døra skal være åpen.
  På denne måten kan man i fsmObstructionSwitch se om tiden døra skal være åpen har gått ut, og lukke døra hvis det ikke er noen hindring. Istedetfor å lukke døra når det ikke lenger er en hindring.
- Endre F_sendRequest slik at den sender request av riktig type, nå inneholder requesten for lite informasjon.
- Legge til at hvis heisen er obstructed, og har mottatt ny request, så skal den sende sin nåværende request tilbake til noden, slik at en annen heis fullfører requesten.
  (Kommer egentlig heisen til å ha noen request hvis den er obstructed? Obstruction stopper ikke heisen fra å åpne døra, så den skal egnetlig kunne cleare sin nåværende request uansett. 
   Og den kommer ikke til å bli satt i IDLE hvis den er obstructed, så den kommer ikke til å motta nye requests. 
   Så dette punktet er kanskje ikke nødvendig. Snakk med Heimvik om dette.)
- Legge til elevatormusic
- SKjønne seg på ops greia til Heimert
- Rydde opp i griseriet. Fjerne unødvendige kommentarer og kode.
*/

var Elevator T_Elevator
var C_stop bool
var C_obstruction bool
var ID int //temp, spør arbo om flytting


func F_RunElevator(ops node.T_NodeOperations, c_requestIn chan T_Request, c_requestOut chan T_Request) {

	Init("localhost:15657") //henter port fra config elno, må smelle på localhost sjæl tror jeg
	Elevator = Init_Elevator(c_requestIn, c_requestOut)

	SetMotorDirection(MD_Down)



    C_stop = false
    C_obstruction = false

	c_readElevatorInfo := make(chan T_ElevatorInfo)
	c_writeElevatorInfo := make(chan T_ElevatorInfo)
	c_quitGetSet := make(chan bool)

	go node.f_Get

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
			F_fsmFloorArrival(a)

		case a := <-drv_obstr: //tipper dette er nok til å kun teste funksjonalitet
			C_obstruction = a
            F_fsmObstructionSwitch(a)
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
