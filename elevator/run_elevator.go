package elevator

import "fmt"

var Elevator T_Elevator

func F_runElevator(requestIn chan T_Request, requestOut chan T_Request){

    Init("localhost:15657")
	Elevator = Init_Elevator()

    SetMotorDirection(MD_Up)
    
    drv_buttons := make(chan ButtonEvent)
    drv_floors  := make(chan int)
    drv_obstr   := make(chan bool)
    drv_stop    := make(chan bool)    
    //drv_req := make(chan T_request)

    go PollButtons(drv_buttons)
    go PollFloorSensor(drv_floors)
    go PollObstructionSwitch(drv_obstr)
    go PollStopButton(drv_stop)
    //go PollNewRequest(requestOut)
    
    for {
        select {
        case a := <- drv_buttons: 
            // generate a new request based on buttonpress, and send to master
            fmt.Printf("%v+\n",a.Floor)

        case a := <- drv_floors:
            F_fsmFloorArrival(a)
			
        case a := <- drv_obstr:
            fmt.Printf("%v+\n",a)
            
        case a := <- drv_stop:
            fmt.Printf("%v+\n",a)

		// case a := <- requestOut:
			// when receiving new request, add to Elevator.Request array, and call F_chooseDirection(Elevator)
		}
    }    
}