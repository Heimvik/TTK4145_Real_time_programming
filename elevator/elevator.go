package elevator

type t_ElevatorState int

const (
	EB_Idle t_ElevatorState = iota
	EB_DoorOpen t_ElevatorState = iota
	EB_Moving t_ElevatorState = iota
)

//Keeping this in case of future improvements regarding secondary requirements,
//see single-elevator/elevator.go for inspiration

type T_Elevator struct {
    Floor     		int
    MotorDirection	MotorDirection
    Request			T_Request
	// Requests        [NUMFLOORS][NUMBUTTONS]int
    State 		t_ElevatorState
}

//func for adding request, or create chan which sends to 

