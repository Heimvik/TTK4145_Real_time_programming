package elevator

import "elevio"
import "fmt"

/*

*/


const ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen ElevatorBehaviour = iota
	EB_Moving ElevatorBehaviour = iota
)

const ClearRequestVariant int

const (
	CV_All ClearRequestVariant = iota
	CV_InDirn ClearRequestVariant = iota
)

type Elevator struct {
    floor     		int
    motorDirection	MotorDirection
    requests  		[_numFloors][_numButtons]int
    rehaviour 		ElevatorBehaviour
    config    struct {
        clearRequestVariant ClearRequestVariant
        doorOpenDuration_s  float64
    }
}


