package elevator

type t_ElevatorBehaviour int

const (
	EB_Idle t_ElevatorBehaviour = iota
	EB_DoorOpen t_ElevatorBehaviour = iota
	EB_Moving t_ElevatorBehaviour = iota
)

//Keeping this in case of future improvements regarding secondary requirements,
//see single-elevator/elevator.go for inspiration

type Elevator struct {
    Floor     		int
    MotorDirection	MotorDirection
    //Requests        [NUMFLOORS][NUMBUTTONS]int
    Behaviour 		t_ElevatorBehaviour
}