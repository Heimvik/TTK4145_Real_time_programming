package node

import (
	//"fmt"

	"strconv"
	"the-elevator/network/network_libraries/bcast"
	"time"
)

// KILDE:

//This should:
//Take in:
// - channel to put the message in
// - port
//Give out:
// - Object of T_Message
// - Array of connected nodes (any unconnected nodes)

func f_VerifyMasterMessage(c_received chan T_MasterMessage, c_verifiedReceived chan T_MasterMessage, c_currentConnectedNode chan T_NodeInfo) {
	for {
		receivedMessage := <-c_received
		if true {
			c_verifiedReceived <- receivedMessage
			c_currentConnectedNode <- receivedMessage.Transmitter
		}
	}
}
func f_VerifySlaveMessage(c_received chan T_SlaveMessage, c_verifiedReceived chan T_SlaveMessage, c_currentConnectedNode chan T_NodeInfo) {
	for {
		receivedMessage := <-c_received
		if true && true {
			c_verifiedReceived <- receivedMessage
			c_currentConnectedNode <- receivedMessage.Transmitter
		}
	}
}

func f_RemoveNode(nodes []T_NodeInfo, nodeToRemove T_NodeInfo) []T_NodeInfo {
	for i, p_node := range nodes {
		if p_node == nodeToRemove {
			return append(nodes[:i], nodes[i+1:]...)
		}
	}
	return nodes
}

func f_AppendNode(nodes []T_NodeInfo, nodeToRemove T_NodeInfo) []T_NodeInfo {
	return append(nodes, nodeToRemove)
}

func f_UpdateNodes(c_currentNode chan T_NodeInfo, ops T_NodeOperations, c_newConnectedNodes chan []T_NodeInfo) {
	for {
		currentNode := <-c_currentNode
		oldConnectedNodes := f_GetConnectedNodes(ops)
		newNode := true
		for _, oldConnectedNode := range oldConnectedNodes {
			if currentNode.PRIORITY == oldConnectedNode.PRIORITY {
				newNode = false
				break
			}
		}
		if newNode {
			connectedNodes := f_AppendNode(oldConnectedNodes, currentNode)
			c_newConnectedNodes <- connectedNodes
		} else {
			//IMPORTANT: IMPLEMENTATION TO REMOVE UNCONNECTED NODE HERE!
			//connectedNodes := f_RemoveNode(oldConnectedNodes, currentNode)
			//c_newConnectedNodes <- connectedNodes
		}
	}
}

func F_TransmitSlaveMessage(c_transmitMessage chan T_SlaveMessage, port int) {
	go bcast.Transmitter(port, c_transmitMessage)
}
func F_TransmitMasterMessage(ops T_NodeOperations, port int) {
	c_masterMessage := make(chan T_MasterMessage)
	go bcast.Transmitter(port, c_masterMessage)
	for {
		c_nodeInfoResponse := make(chan T_NodeInfo)
		ops.c_readNodeInfo <- c_nodeInfoResponse // Send the response channel to the NodeOperationManager
		nodeInfo := <-c_nodeInfoResponse

		c_globalQueueResponse := make(chan []T_GlobalQueueEntry)
		ops.c_readGlobalQueue <- c_globalQueueResponse // Send the response channel to the NodeOperationManager
		globalQueue := <-c_globalQueueResponse

		masterMessage := T_MasterMessage{
			Transmitter: nodeInfo,
			GlobalQueue: globalQueue,
		}
		c_masterMessage <- masterMessage
		F_WriteLog("MasterMessage sent on port: " + strconv.Itoa(port))
		time.Sleep(time.Duration(MMMILLS) * time.Millisecond)
	}
}

func F_ReceiveSlaveMessage(c_verifiedMessage chan T_SlaveMessage, ops T_NodeOperations, c_newConnectedNodes chan []T_NodeInfo, port int) {
	c_currentNode := make(chan T_NodeInfo)
	c_receive := make(chan T_SlaveMessage)

	go bcast.Receiver(port, c_receive)
	go f_VerifySlaveMessage(c_receive, c_verifiedMessage, c_currentNode)
	go f_UpdateNodes(c_currentNode, ops, c_newConnectedNodes)
}

func F_ReceiveMasterMessage(c_verifiedMessage chan T_MasterMessage, ops T_NodeOperations, c_newConnectedNodes chan []T_NodeInfo, port int) {
	c_currentNode := make(chan T_NodeInfo)
	c_receivedMessage := make(chan T_MasterMessage)

	go bcast.Receiver(port, c_receivedMessage)
	go f_VerifyMasterMessage(c_receivedMessage, c_verifiedMessage, c_currentNode)
	go f_UpdateNodes(c_currentNode, ops, c_newConnectedNodes)
}

//

/*What do the studasser tenker om å ha en hel FMS som løsning?
func f_ReceiveFSM(c_receiveMessage chan T_Message, c_transmitMessage chan T_Message, connectedNodes []*Node){
	//receive fsm
	//should handle the whole UDP transaction


	var step int = 0
	switch step{
	case 0: //check for messages and acceptance test on them
		receivedMessage := <- c_receiveMessage
		if f_AcceptancetestReceive(receivedMessage){
			step +=1
		}else{
			//kys or remove node from connectednodes
		}
	case 1: //readback to the node that sent it
		readbackMessage := {

		}
		c_transmitMessage <- c_transmitMessage
	}
	//adding/removing from connected nodes, taking in message
}
*/
