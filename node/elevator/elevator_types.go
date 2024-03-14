package elevator

type T_ElevatorState uint8

const (
	ELEVATORSTATE_IDLE     T_ElevatorState = 0
	ELEVATORSTATE_DOOROPEN T_ElevatorState = 1
	ELEVATORSTATE_MOVING   T_ElevatorState = 2
)

type T_ElevatorDirection int8

const (
	ELEVATORDIRECTION_UP   T_ElevatorDirection = 1
	ELEVATORDIRECTION_DOWN T_ElevatorDirection = -1
	ELEVATORDIRECTION_NONE T_ElevatorDirection = 0
)

type T_RequestState uint8

const (
	REQUESTSTATE_UNASSIGNED T_RequestState = 0
	REQUESTSTATE_ASSIGNED   T_RequestState = 1
	REQUESTSTATE_ACTIVE     T_RequestState = 2
	REQUESTSTATE_DONE       T_RequestState = 3
)

type T_CallType uint8

const (
	CALLTYPE_NONECALL T_CallType = 0
	CALLTYPE_CAB      T_CallType = 1
	CALLTYPE_HALL     T_CallType = 2
)

type T_ButtonType int

const (
	BUTTONTYPE_HALLUP   T_ButtonType = 0
	BUTTONTYPE_HALLDOWN T_ButtonType = 1
	BUTTONTYPE_CAB      T_ButtonType = 2
)

type T_ButtonEvent struct {
	Floor  int
	Button T_ButtonType
}

type T_Request struct {
	Id        uint16
	State     T_RequestState
	Calltype  T_CallType
	Floor     int8
	Direction T_ElevatorDirection
}

type T_Elevator struct {
	CurrentID      int
	StopButton     bool
	P_info         *T_ElevatorInfo //MUST be pointer to info (points to info stored in ThisNode.NodeInfo.ElevatorInfo)
	ServeRequest T_Request      //Pointer to the current request you are serviceing
}

type T_ElevatorInfo struct {
	Direction  T_ElevatorDirection
	Floor      int8 //ranges from 0-3
	State      T_ElevatorState
	Obstructed bool
}

type T_ElevatorOperations struct {
	C_getElevator    chan chan T_Elevator
	C_setElevator    chan T_Elevator
	C_getSetElevator chan chan T_Elevator
}

type T_GetSetElevatorInterface struct {
	C_get chan T_Elevator
	C_set chan T_Elevator
}

type T_ElevatorChannels struct {
	getSetElevatorInterface T_GetSetElevatorInterface
	C_timerStart            chan bool
	C_timerStop             chan bool
	C_timerTimeout          chan bool
	C_buttons               chan T_ButtonEvent
	C_floors                chan int
	C_obstr                 chan bool
	C_stop                  chan bool
	C_requestIn             chan T_Request
	C_requestOut            chan T_Request
}
