package elevator

type T_ElevatorState int
type T_MotorDirection int

const (
	EB_Idle T_ElevatorState = iota
	EB_DoorOpen T_ElevatorState = iota
	EB_Moving T_ElevatorState = iota
)

//Keeping this in case of future improvements regarding secondary requirements,
//see single-elevator/elevator.go for inspiration

type T_Elevator struct {
    Floor     		int
    MotorDirection	MotorDirection
    Request			T_Request
	// Requests        [NUMFLOORS][NUMBUTTONS]int
    State 		T_ElevatorState
}

//func for adding request, or create chan which sends to 

