package node

var FLOORS int = 4

type Node struct {
	Priority       int
	Role           NodeRole
	ConnectedNodes []Node
	NodeElevator   Elevator
}

type NodeRole struct {
	Master MasterNode
	Slave  SlaveNode
}

type MasterNode struct {
	GlobalQueue GlobalQueue
}

type SlaveNode struct {
	LocalQueue LocalQueue //NB, takes only one order
}

type GlobalQueue struct {
	Request Request
}

type LocalQueue struct {
}

type Elevator struct {
	RequestsToDistribution chan Request
	RequestsToService      chan Request
	Floor                  int
	Direction              Direction
	Avalibale              bool //Thoggled whenever disconnected/unavalebale/door sensor
}

type Request struct {
	Calltype  Call
	Elevator  int
	Floor     int
	Direction Direction
}

type Call int
type Direction int

const (
	down Direction = 0
	up   Direction = 1
)

const (
	cab  Call = 0
	hall Call = 1
)
