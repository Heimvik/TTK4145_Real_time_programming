package node

import (
	"the-elevator/node/elevator"
	"time"
)

var nodeOperations = T_NodeOperations{
	c_getNodeInfo:          make(chan chan T_NodeInfo),
	c_setNodeInfo:          make(chan T_NodeInfo),
	c_getSetNodeInfo:       make(chan chan T_NodeInfo),
	c_getGlobalQueue:       make(chan chan []T_GlobalQueueEntry),
	c_setGlobalQueue:       make(chan []T_GlobalQueueEntry),
	c_getSetGlobalQueue:    make(chan chan []T_GlobalQueueEntry),
	c_getConnectedNodes:    make(chan chan []T_NodeInfo),
	c_setConnectedNodes:    make(chan []T_NodeInfo),
	c_getSetConnectedNodes: make(chan chan []T_NodeInfo),
}
var elevatorOperations = elevator.T_ElevatorOperations{
	C_getElevator:    make(chan chan elevator.T_Elevator),
	C_setElevator:    make(chan elevator.T_Elevator),
	C_getSetElevator: make(chan chan elevator.T_Elevator),
}

func f_NodeOperationManager(node *T_Node) {
	for {
		select {
		case responseChan := <-nodeOperations.c_getNodeInfo:
			responseChan <- node.NodeInfo
		case newNodeInfo := <-nodeOperations.c_setNodeInfo:
			node.NodeInfo = newNodeInfo
			node.Elevator.P_info = &node.NodeInfo.ElevatorInfo
		case responseChan := <-nodeOperations.c_getSetNodeInfo:
			responseChan <- node.NodeInfo
			node.NodeInfo = <-responseChan
			node.Elevator.P_info = &node.NodeInfo.ElevatorInfo

		case responseChan := <-nodeOperations.c_getGlobalQueue:
			responseChan <- node.GlobalQueue
		case newGlobalQueue := <-nodeOperations.c_setGlobalQueue:
			node.GlobalQueue = newGlobalQueue
		case responseChan := <-nodeOperations.c_getSetGlobalQueue:
			responseChan <- node.GlobalQueue
			node.GlobalQueue = <-responseChan

		case responseChan := <-nodeOperations.c_getConnectedNodes:
			responseChan <- node.ConnectedNodes
		case newConnectedNodes := <-nodeOperations.c_setConnectedNodes:
			node.ConnectedNodes = newConnectedNodes
		case responseChan := <-nodeOperations.c_getSetConnectedNodes:
			responseChan <- node.ConnectedNodes
			node.ConnectedNodes = <-responseChan

		case responseChan := <-elevatorOperations.C_getElevator:
			responseChan <- node.Elevator
		case newElevator := <-elevatorOperations.C_setElevator:
			node.Elevator = newElevator
		case responseChan := <-elevatorOperations.C_getSetElevator:
			responseChan <- node.Elevator
			node.Elevator = <-responseChan

		default:
		}
	}
}

func f_GetNodeInfo() T_NodeInfo {
	c_responseChan := make(chan T_NodeInfo)
	nodeOperations.c_getNodeInfo <- c_responseChan 
	nodeInfo := <-c_responseChan 
	return nodeInfo
}
func f_SetNodeInfo(nodeInfo T_NodeInfo) {
	nodeOperations.c_setNodeInfo <- nodeInfo // Send the nodeInfo directly to be written
}

func f_GetSetNodeInfo(c_getSetNodeInfoInterface chan T_GetSetNodeInfoInterface) {
	for {
	WAITFORINTERFACE:
		select {
		case nodeInfoInterface := <-c_getSetNodeInfoInterface:
			c_responsChan := make(chan T_NodeInfo)
			nodeOperations.c_getSetNodeInfo <- c_responsChan
			getSetTimer := time.NewTicker(time.Duration(GETSETPERIOD) * time.Second)
			for {
				select {
				case oldNodeInfo := <-c_responsChan:
					nodeInfoInterface.c_get <- oldNodeInfo
				case newNodeInfo := <-nodeInfoInterface.c_set:
					c_responsChan <- newNodeInfo
					break WAITFORINTERFACE
				case <-getSetTimer.C:
					F_WriteLog("Ended GetSet goroutine of NI because of deadlock")
					break WAITFORINTERFACE
				}
			}
		}
	}
}

func f_GetGlobalQueue() []T_GlobalQueueEntry {
	c_responseChan := make(chan []T_GlobalQueueEntry)
	nodeOperations.c_getGlobalQueue <- c_responseChan // Send the response channel to the NodeOperationManager
	globalQueue := <-c_responseChan                   // Receive the global queue from the response channel
	return globalQueue
}
func f_SetGlobalQueue(globalQueue []T_GlobalQueueEntry) {
	nodeOperations.c_setGlobalQueue <- globalQueue // Send the globalQueue directly to be written
}

func f_GetSetGlobalQueue(c_getSetGlobalQueueInterface chan T_GetSetGlobalQueueInterface) {
	for {
	WAITFORINTERFACE:
		select {
		case globalQueueInterface := <-c_getSetGlobalQueueInterface:
			c_responsChan := make(chan []T_GlobalQueueEntry)
			nodeOperations.c_getSetGlobalQueue <- c_responsChan
			getSetTimer := time.NewTicker(time.Duration(GETSETPERIOD) * time.Second)
			for {
				select {
				case oldGlobalQueue := <-c_responsChan:
					globalQueueInterface.c_get <- oldGlobalQueue
				case newGlobalQueue := <-globalQueueInterface.c_set:
					c_responsChan <- newGlobalQueue
					break WAITFORINTERFACE
				case <-getSetTimer.C:
					F_WriteLog("Ended GetSet goroutine of GQ because of deadlock")
					break WAITFORINTERFACE
				}
			}
		}
	}
}

func f_GetConnectedNodes() []T_NodeInfo {
	c_responseChan := make(chan []T_NodeInfo)
	nodeOperations.c_getConnectedNodes <- c_responseChan // Send the response channel to the NodeOperationManager
	connectedNodes := <-c_responseChan                   // Receive the connected nodes from the response channel
	return connectedNodes
}
func f_SetConnectedNodes(connectedNodes []T_NodeInfo) {
	nodeOperations.c_setConnectedNodes <- connectedNodes // Send the connectedNodes directly to be written
}
func f_GetSetConnectedNodes(c_getSetConnectedNodesInterface chan T_GetSetConnectedNodesInterface) {
	for {
	WAITFORINTERFACE:
		select {
		case globalQueueInterface := <-c_getSetConnectedNodesInterface:
			c_responsChan := make(chan []T_NodeInfo)
			nodeOperations.c_getSetConnectedNodes <- c_responsChan
			getSetTimer := time.NewTicker(time.Duration(GETSETPERIOD) * time.Second)
			for {
				select {
				case oldConnectedNodes := <-c_responsChan:
					globalQueueInterface.c_get <- oldConnectedNodes
				case newConnectedNodes := <-globalQueueInterface.c_set:
					c_responsChan <- newConnectedNodes
					break WAITFORINTERFACE
				case <-getSetTimer.C:
					F_WriteLog("Ended GetSet goroutine of GQ because of deadlock")
					break WAITFORINTERFACE
				}
			}
		}
	}
}
