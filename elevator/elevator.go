package elevator

type T_ElevatorState int

const (
	EB_Idle     T_ElevatorState = iota
	EB_DoorOpen T_ElevatorState = iota
	EB_Moving   T_ElevatorState = iota
)

//Keeping this in case of future improvements regarding secondary requirements,
//see single-elevator/elevator.go for inspiration

type T_Elevator struct {
	Floor          int
	MotorDirection MotorDirection
	Request        []T_Request
	// Requests        [NUMFLOORS][NUMBUTTONS]int
	State T_ElevatorState
}

func Init_Elevator() T_Elevator {
	return T_Elevator{-1, MD_Stop, make([]T_Request,0), EB_Idle}
}

//func for adding request, or create chan which sends to

// choose direction, based on current request of elevator
func F_chooseDirection(e T_Elevator) {
	switch e.MotorDirection {
	case MD_Up:
		if f_requestOver(e) {
			e.MotorDirection = MD_Up
			e.State = EB_Moving
		} else if f_requestHere(e) {
			e.MotorDirection = MD_Down
			e.State = EB_DoorOpen
		} else if f_requestUnder(e) {
			e.MotorDirection = MD_Down
			e.State = EB_Moving
		}
	case MD_Down:
		if f_requestUnder(e) {
			e.MotorDirection = MD_Down
			e.State = EB_Moving
		} else if f_requestHere(e) {
			e.MotorDirection = MD_Up
			e.State = EB_DoorOpen
		} else if f_requestOver(e) {
			e.MotorDirection = MD_Up
			e.State = EB_Moving
		}
	case MD_Stop:
		if f_requestHere(e) {
			e.MotorDirection = MD_Stop
			e.State = EB_DoorOpen
		} else if f_requestOver(e) {
			e.MotorDirection = MD_Up
			e.State = EB_Moving
		} else if f_requestUnder(e) {
			e.MotorDirection = MD_Down
			e.State = EB_Moving
		}
	}
}

//might need updating when improving functionality
func F_shouldStop(e T_Elevator) bool {
	if len(e.Request) > 0{
		if e.Floor == e.Request[0].Floor{
			return true
		} else {
			return false
		}
	}
	return true
}

func F_clearRequest(e T_Elevator) {
	e.Request = e.Request[1:]
}

func f_requestUnder(e T_Elevator) bool {
	return e.Request[0].Floor < e.Floor
}

func f_requestOver(e T_Elevator) bool {
	return e.Request[0].Floor > e.Floor
}

func f_requestHere(e T_Elevator) bool {
	return e.Request[0].Floor == e.Floor
}

