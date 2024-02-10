package node

type Node struct{
	priority int
	role NodeRole
	connectedNodes []Node
	nodeElevator Elevator
}

type NodeRole struct{
	master MasterNode
	slave SlaveNode
}

type MasterNode struct{
	globalQueue GlobalQueue
}

type SlaveNode struct{
	localQueue LocalQueue //NB, takes only one order
}

type Elevator struct{
	requestsToDistribution chan Request
	requestsToService chan Request
	floor int
	direction Direction
	avalibale bool //Thoggled whenever disconnected/unavalebale/door sensor
}

type Request struct{
	calltype Call
	elevator int
	floor int
	direction Direction
}

type Call int
type Direction int

const (
	up Direction = iota
	down Direction
)

const (
	hall Call = iota
	cab Call
)



