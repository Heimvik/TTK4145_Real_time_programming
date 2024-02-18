package elevator

type T_ElevatorState int
type T_ElevatorDirection int

const (
	Idle     T_ElevatorState = iota
	DoorOpen T_ElevatorState = iota
	Moving   T_ElevatorState = iota
)

const (
	Up   T_ElevatorDirection = iota
	Down T_ElevatorDirection = iota
	None T_ElevatorDirection = iota
)

//Keeping this in case of future improvements regarding secondary requirements,
//see single-elevator/elevator.go for inspiration

type T_Elevator struct {
	P_info              *T_ElevatorInfo     //Poitner to the info of this elevator
	P_serveRequest      *T_Request          //Pointer to the current request you are serviceing
	C_ReceiveRequest    chan T_Request      //Request to put in ServeRequest and do, you will get this from node
	C_DistributeRequest chan T_Request      //Requests to distriburte to node, you shoud provide this to node
	C_DistributeInfo    chan T_ElevatorInfo //Info on elevator whereabouts, you should provide this to node
	//three last ones has to be channels as elevator and node has seperate goroutines, has to be like this
	//on comilation of one request: redistribute it on D
}
type T_ElevatorInfo struct {
	Direction T_Direction
	Floor     int
	State     T_ElevatorState
}

//func for adding request, or create chan which sends to
