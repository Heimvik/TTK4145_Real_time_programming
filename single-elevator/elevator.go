package singleelevator


// import "fmt"

const NUMFLOORS int = 4
const NUMBUTTONS int = 3

type t_ElevatorBehaviour int

const (
	EB_Idle t_ElevatorBehaviour = iota
	EB_DoorOpen t_ElevatorBehaviour = iota
	EB_Moving t_ElevatorBehaviour = iota
)

type t_ClearRequestVariant int

const (
	CV_All t_ClearRequestVariant = iota
	CV_InDirn t_ClearRequestVariant = iota
)

type Elevator struct {
    Floor     		int
    MotorDirection	MotorDirection
    Requests        [NUMFLOORS][NUMBUTTONS]int
    Behaviour 		t_ElevatorBehaviour
    Config          struct {
        clearRequestVariant t_ClearRequestVariant
        doorOpenDuration_s  float64
    }
}