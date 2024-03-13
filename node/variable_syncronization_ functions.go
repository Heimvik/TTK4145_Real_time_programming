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

/*
Serializes access to node variables, ensuring thread-safe operations on the node's state, including elevator information and global queue changes.

Prerequisites: Initialized channels for node and elevator operation communication.

Returns: Nothing, but ensures thread-safe updates to node and elevator states through serialized access.
*/
func F_NodeOperationManager(node *T_Node) {
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

/*
Retrieves the current state information of the node, including priority and elevator status, through synchronized access to ensure data integrity.

Prerequisites: None.

Returns: The current state information of the node as a T_NodeInfo structure.
*/
func f_GetNodeInfo() T_NodeInfo {
	c_responseChan := make(chan T_NodeInfo)
	nodeOperations.c_getNodeInfo <- c_responseChan
	nodeInfo := <-c_responseChan
	return nodeInfo
}

/*
Updates the node's state information with synchronized access.

Prerequisites: None.

Returns: Nothing, but updates the node's state.
*/
func f_SetNodeInfo(nodeInfo T_NodeInfo) {
	nodeOperations.c_setNodeInfo <- nodeInfo
}

/*
Provides synchronized access for retrieving and updating the node's state information, ensuring thread-safe operations.

Prerequisites: None.

Returns: Nothing, but allows retrieval and update of node's state.
*/
func f_GetSetNodeInfo(c_getSetNodeInfoInterface chan T_GetSetNodeInfoInterface) {
	for {
	WAITFORINTERFACE:
		select {
		case nodeInfoInterface := <-c_getSetNodeInfoInterface:
			c_responsChan := make(chan T_NodeInfo)
			nodeOperations.c_getSetNodeInfo <- c_responsChan
			getSetTimer := time.NewTicker(time.Duration(GETSET_PERIOD) * time.Second)
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

/*
Retrieves the current global queue of elevator requests through synchronized access.

Prerequisites: None.

Returns: The current global queue.
*/
func f_GetGlobalQueue() []T_GlobalQueueEntry {
	c_responseChan := make(chan []T_GlobalQueueEntry)
	nodeOperations.c_getGlobalQueue <- c_responseChan 
	globalQueue := <-c_responseChan                   
	return globalQueue
}

/*
Updates the global queue of elevator requests using synchronized access to ensure consistency.

Prerequisites: None.

Returns: Nothing, but updates the global queue.
*/
func f_SetGlobalQueue(globalQueue []T_GlobalQueueEntry) {
	nodeOperations.c_setGlobalQueue <- globalQueue // Send the globalQueue directly to be written
}

/*
Enables synchronized retrieval and updating of the global queue, ensuring data consistency.

Prerequisites: None.

Returns: Nothing, but facilitates access and modification of the global queue.
*/
func f_GetSetGlobalQueue(c_getSetGlobalQueueInterface chan T_GetSetGlobalQueueInterface) {
	for {
	WAITFORINTERFACE:
		select {
		case globalQueueInterface := <-c_getSetGlobalQueueInterface:
			c_responsChan := make(chan []T_GlobalQueueEntry)
			nodeOperations.c_getSetGlobalQueue <- c_responsChan
			getSetTimer := time.NewTicker(time.Duration(GETSET_PERIOD) * time.Second)
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

/*
Fetches the current list of connected nodes in the system through synchronized access to maintain up-to-date network information.

Prerequisites: None.

Returns: The current list of connected nodes.
*/
func f_GetConnectedNodes() []T_NodeInfo {
	c_responseChan := make(chan []T_NodeInfo)
	nodeOperations.c_getConnectedNodes <- c_responseChan 
	connectedNodes := <-c_responseChan                  
	return connectedNodes
}

/*
Updates the list of connected nodes in the system using synchronized access to ensure network information remains accurate.

Prerequisites: None.

Returns: Nothing, but updates the list of connected nodes.
*/
func f_SetConnectedNodes(connectedNodes []T_NodeInfo) {
	nodeOperations.c_setConnectedNodes <- connectedNodes // Send the connectedNodes directly to be written
}

/*
Allows for synchronized retrieval and updating of the connected nodes list, ensuring the network's integrity and accuracy.

Prerequisites: None.

Returns: Nothing, but provides access and modification capabilities for the list of connected nodes.
*/
func f_GetSetConnectedNodes(c_getSetConnectedNodesInterface chan T_GetSetConnectedNodesInterface) {
	for {
	WAITFORINTERFACE:
		select {
		case globalQueueInterface := <-c_getSetConnectedNodesInterface:
			c_responsChan := make(chan []T_NodeInfo)
			nodeOperations.c_getSetConnectedNodes <- c_responsChan
			getSetTimer := time.NewTicker(time.Duration(GETSET_PERIOD) * time.Second)
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
