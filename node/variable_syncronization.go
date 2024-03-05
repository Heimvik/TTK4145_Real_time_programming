package node

import (
	"the-elevator/node/elevator"
	"time"
)

func f_NodeOperationManager(node *T_Node, nodeOps T_NodeOperations, elevatorOps elevator.T_ElevatorOperations) {
	for {
		select {
		case responseChan := <-nodeOps.c_readNodeInfo:
			responseChan <- node.NodeInfo
		case newNodeInfo := <-nodeOps.c_writeNodeInfo:
			node.NodeInfo = newNodeInfo
			node.Elevator.P_info = &node.NodeInfo.ElevatorInfo
		case responseChan := <-nodeOps.c_readAndWriteNodeInfo:
			responseChan <- node.NodeInfo
			node.NodeInfo = <-responseChan
			node.Elevator.P_info = &node.NodeInfo.ElevatorInfo

		case responseChan := <-nodeOps.c_readGlobalQueue:
			responseChan <- node.GlobalQueue
		case newGlobalQueue := <-nodeOps.c_writeGlobalQueue:
			node.GlobalQueue = newGlobalQueue
		case responseChan := <-nodeOps.c_readAndWriteGlobalQueue:
			responseChan <- node.GlobalQueue
			node.GlobalQueue = <-responseChan

		case responseChan := <-nodeOps.c_readConnectedNodes:
			responseChan <- node.ConnectedNodes
		case newConnectedNodes := <-nodeOps.c_writeConnectedNodes:
			node.ConnectedNodes = newConnectedNodes
		case responseChan := <-nodeOps.c_readAndWriteConnectedNodes:
			responseChan <- node.ConnectedNodes
			node.ConnectedNodes = <-responseChan

		case responseChan := <-elevatorOps.C_readElevator:
			responseChan <- node.Elevator
		case newElevator := <-elevatorOps.C_writeElevator:
			node.Elevator = newElevator
		case responseChan := <-elevatorOps.C_readAndWriteElevator:
			responseChan <- node.Elevator
			node.Elevator = <-responseChan
		default:
			//No sleep, this has to be fastest of them all
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
			F_WriteLog("Ended GetSet goroutine of NI because of deadlock")
		}
		//No sleep
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
			F_WriteLog("Ended GetSet goroutine of GQ because of deadlock")
		}
		//No sleep
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
		//No sleep
	}
}
