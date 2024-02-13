package node

import (
	"the-elevator/network/network_libraries/bcast"
)

// KILDE:
func f_AcceptancetestReceive(message T_Message) bool {
	//make approperiate Acceptance test, test all values for decent values
	//test checksum

	return true
}

//This should:
//Take in:
// - channel to put the message in
// - port
//Give out:
// - Object of T_Message
// - Array of connected nodes (any unconnected nodes)

func f_VerifyReceive(c_received chan T_Message, c_verifiedReceived chan T_Message, c_currentConnectedNodes chan int) {
	receivedMessage := <-c_received
	if f_AcceptancetestReceive(receivedMessage) && true {
		c_verifiedReceived <- receivedMessage
		c_currentConnectedNodes <- receivedMessage.Transmitter
	} else {
		return
	}
}

func f_RemoveNode(nodes []int, nodeToRemove int) []int {
	for i, p_node := range nodes {
		if p_node == nodeToRemove {
			return append(nodes[:i], nodes[i+1:]...)
		}
	}
	return nodes
}

func f_AppendNode(nodes []int, nodeToRemove int) []int {
	return append(nodes, nodeToRemove)
}

func F_TransmitMessages(c_transmitMessage chan T_Message, port int){
	go bcast.Transmitter(port, c_transmitMessage)
}

func f_updateConnectedNodes(c_connectedNodes chan []int, c_currentConnectedNodes chan int){
	
	for {
		currentConnectedNode := <-c_currentConnectedNodes
		connectedNodes := <-c_connectedNodes
		foundNode := true
		for _, p_oldConnectedNode := range connectedNodes {
			if currentConnectedNode != p_oldConnectedNode {
				foundNode = false
			} else {
				foundNode = true
				break
			}
		}
		if foundNode {
			connectedNodes = f_AppendNode(connectedNodes, currentConnectedNode)
			c_connectedNodes <- connectedNodes
		} else {
			connectedNodes = f_RemoveNode(connectedNodes, currentConnectedNode)
			c_connectedNodes <- connectedNodes
		}
	}
}

// requires that it receives its own messages
func F_ReceiveMessages(c_verifiedMessage chan T_Message, c_connectedNodes chan []int ,port int) {
	c_currentConnectedNodes := make(chan int)
	c_receive := make(chan T_Message)

	go bcast.Receiver(port, c_receive)
	go f_VerifyReceive(c_receive, c_verifiedMessage, c_currentConnectedNodes)
	go f_updateConnectedNodes(c_connectedNodes, c_currentConnectedNodes)

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
