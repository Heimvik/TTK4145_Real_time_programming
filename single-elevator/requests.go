package singleelevator

type DirnBehaviourPair struct {
	Dirn      MotorDirection
	Behaviour t_ElevatorBehaviour
}

func Requests_chooseDirection(e Elevator) DirnBehaviourPair {
	switch e.MotorDirection {
	case MD_Up:
		if requests_above(e) {
			return DirnBehaviourPair{MD_Up, EB_Moving}
		} else if requests_here(e) {
			return DirnBehaviourPair{MD_Down, EB_DoorOpen}
		} else if requests_below(e) {
			return DirnBehaviourPair{MD_Down, EB_Moving}
		} else {
			return DirnBehaviourPair{MD_Stop, EB_Idle}
		}

	case MD_Down:
		if requests_below(e) {
			return DirnBehaviourPair{MD_Down, EB_Moving}
		} else if requests_here(e) {
			return DirnBehaviourPair{MD_Up, EB_DoorOpen}
		} else if requests_above(e) {
			return DirnBehaviourPair{MD_Up, EB_Moving}
		} else {
			return DirnBehaviourPair{MD_Stop, EB_Idle}
		}

	case MD_Stop:
		if requests_here(e) {
			return DirnBehaviourPair{MD_Stop, EB_DoorOpen}
		} else if requests_above(e) {
			return DirnBehaviourPair{MD_Up, EB_Moving}
		} else if requests_below(e) {
			return DirnBehaviourPair{MD_Down, EB_Moving}
		} else {
			return DirnBehaviourPair{MD_Stop, EB_Idle}
		}
	default:
		return DirnBehaviourPair{MD_Stop, EB_Idle}
	}
}

func requests_above(e Elevator) bool {
	for f := e.Floor + 1; f < NUMFLOORS; f++ {
		for b := 0; b < NUMBUTTONS; b++ {
			if e.Requests[f][b] == 1 {
				return true

			}
		}
	}
	return false
}

func requests_below(e Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for b := 0; b < NUMBUTTONS; b++ {
			if e.Requests[f][b] == 1 {
				return true

			}
		}
	}
	return false
}

func requests_here(e Elevator) bool {
	for b := 0; b < NUMBUTTONS; b++ {
		if e.Requests[e.Floor][b] == 1 {
			return true
		}
	}
	return false
}
