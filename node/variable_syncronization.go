package node

import (
	"the-elevator/elevator"
	"time"
)

func f_NodeOperationManager(node *T_Node, ops T_NodeOperations) {
	for {
		select {
		case responseChan := <-ops.c_readNodeInfo:
			responseChan <- node.Info
		case newNodeInfo := <-ops.c_writeNodeInfo:
			node.Info = newNodeInfo
		case responseChan := <-ops.c_readAndWriteNodeInfo:
			responseChan <- node.Info
			node.Info = <-responseChan

		case responseChan := <-ops.c_readGlobalQueue:
			responseChan <- node.GlobalQueue
		case newGlobalQueue := <-ops.c_writeGlobalQueue:
			node.GlobalQueue = newGlobalQueue
		case responseChan := <-ops.c_readAndWriteGlobalQueue:
			responseChan <- node.GlobalQueue
			node.GlobalQueue = <-responseChan

		case responseChan := <-ops.c_readConnectedNodes:
			responseChan <- node.ConnectedNodes
		case newConnectedNodes := <-ops.c_writeConnectedNodes:
			node.ConnectedNodes = newConnectedNodes
		case responseChan := <-ops.c_readAndWriteConnectedNodes:
			responseChan <- node.ConnectedNodes
			node.ConnectedNodes = <-responseChan

		case responseChan := <-ops.c_readElevator:
			responseChan <- *node.P_ELEVATOR
		case newElevator := <-ops.c_writeElevator:
			*node.P_ELEVATOR = newElevator
		case responseChan := <-ops.c_readAndWriteElevator:
			responseChan <- *node.P_ELEVATOR
			*node.P_ELEVATOR = <-responseChan

		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}
func f_GetNodeInfo(ops T_NodeOperations) T_NodeInfo {
	c_responseChan := make(chan T_NodeInfo)
	ops.c_readNodeInfo <- c_responseChan // Send the response channel to the NodeOperationManager
	nodeInfo := <-c_responseChan         // Receive the node info from the response channel
	return nodeInfo
}
func f_SetNodeInfo(ops T_NodeOperations, nodeInfo T_NodeInfo) {
	ops.c_writeNodeInfo <- nodeInfo // Send the nodeInfo directly to be written
}
func f_GetAndSetNodeInfo(ops T_NodeOperations, c_readConnectedNodes chan T_NodeInfo, c_writeConnectedNodes chan T_NodeInfo, c_quit chan bool) { //let run in a sepreate goroutine
	getSetTimer := time.NewTicker(time.Duration(GETSETPERIOD) * time.Second)
	c_responsChan := make(chan T_NodeInfo)

	ops.c_readAndWriteNodeInfo <- c_responsChan
	for {
		select {
		case oldNodeInfo := <-c_responsChan:
			c_readConnectedNodes <- oldNodeInfo
		case newNodeInfo := <-c_writeConnectedNodes:
			c_responsChan <- newNodeInfo
		case <-c_quit:
			return
		case <-getSetTimer.C:
			F_WriteLog("Ended GetSet goroutine of CN because of deadlock")
		}
	}
}
func f_GetGlobalQueue(ops T_NodeOperations) []T_GlobalQueueEntry {
	c_responseChan := make(chan []T_GlobalQueueEntry)
	ops.c_readGlobalQueue <- c_responseChan // Send the response channel to the NodeOperationManager
	globalQueue := <-c_responseChan         // Receive the global queue from the response channel
	return globalQueue
}
func f_SetGlobalQueue(ops T_NodeOperations, globalQueue []T_GlobalQueueEntry) {
	ops.c_writeGlobalQueue <- globalQueue // Send the globalQueue directly to be written
}
func f_GetAndSetGlobalQueue(ops T_NodeOperations, c_readGlobalQueue chan []T_GlobalQueueEntry, c_writeGlobalQueue chan []T_GlobalQueueEntry, c_quit chan bool) { //let run in a sepreate goroutine
	getSetTimer := time.NewTicker(time.Duration(GETSETPERIOD) * time.Second)
	c_responsChan := make(chan []T_GlobalQueueEntry)

	ops.c_readAndWriteGlobalQueue <- c_responsChan
	for {
		select {
		case oldGlobalQueue := <-c_responsChan:
			c_readGlobalQueue <- oldGlobalQueue
		case newGlobalQueue := <-c_writeGlobalQueue:
			c_responsChan <- newGlobalQueue
		case <-c_quit:
			return
		case <-getSetTimer.C:
			F_WriteLog("Ended GetSet goroutine of CN because of deadlock")
		}
	}
}
func f_GetConnectedNodes(ops T_NodeOperations) []T_NodeInfo {
	c_responseChan := make(chan []T_NodeInfo)
	ops.c_readConnectedNodes <- c_responseChan // Send the response channel to the NodeOperationManager
	connectedNodes := <-c_responseChan         // Receive the connected nodes from the response channel
	return connectedNodes
}
func f_SetConnectedNodes(ops T_NodeOperations, connectedNodes []T_NodeInfo) {
	ops.c_writeConnectedNodes <- connectedNodes // Send the connectedNodes directly to be written
}
func f_GetAndSetConnectedNodes(ops T_NodeOperations, c_readConnectedNodes chan []T_NodeInfo, c_writeConnectedNodes chan []T_NodeInfo, c_quit chan bool) { //let run in a sepreate goroutine
	getSetTimer := time.NewTicker(time.Duration(GETSETPERIOD) * time.Second)
	c_responsChan := make(chan []T_NodeInfo)

	ops.c_readAndWriteConnectedNodes <- c_responsChan
	for {
		select {
		case oldConnectedNodes := <-c_responsChan:
			c_readConnectedNodes <- oldConnectedNodes
		case newConnectedNodes := <-c_writeConnectedNodes:
			c_responsChan <- newConnectedNodes
		case <-c_quit:
			return
		case <-getSetTimer.C:
			F_WriteLog("Ended GetSet goroutine of CN because of deadlock")
		}
	}
}
func f_GetElevator(ops T_NodeOperations) elevator.T_Elevator {
	c_responseChan := make(chan elevator.T_Elevator)
	ops.c_readElevator <- c_responseChan // Send the response channel to the NodeOperationManager
	elevator := <-c_responseChan         // Receive the connected nodes from the response channel
	return elevator
}
func f_SetElevator(ops T_NodeOperations, elevator elevator.T_Elevator) {
	ops.c_writeElevator <- elevator // Send the connectedNodes directly to be written
}
func f_GetAndSetElevator(ops T_NodeOperations, c_readElevator chan elevator.T_Elevator, c_writeElevator chan elevator.T_Elevator, c_quit chan bool) { //let run in a sepreate goroutine
	getSetTimer := time.NewTicker(time.Duration(GETSETPERIOD) * time.Second)
	c_responsChan := make(chan elevator.T_Elevator)

	ops.c_readAndWriteElevator <- c_responsChan
	for {
		select {
		case oldElevator := <-c_responsChan:
			c_readElevator <- oldElevator
		case newElevator := <-c_writeElevator:
			c_responsChan <- newElevator
		case <-c_quit:
			return
		case <-getSetTimer.C:
			F_WriteLog("Ended GetSet goroutine of CN because of deadlock")
		}
	}
}
