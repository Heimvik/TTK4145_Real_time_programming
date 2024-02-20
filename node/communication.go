package node

import (
	//"fmt"

	"the-elevator/network/network_libraries/bcast"
)

// KILDE:

//This should:
//Take in:
// - channel to put the message in
// - port
//Give out:
// - Object of T_Message
// - Array of connected nodes (any unconnected nodes)

func f_AcceptancetestSM() bool {
	return true
}
func f_AcceptancetestMM() bool {
	return true
}

func F_TransmitSlaveMessage(c_transmitMessage chan T_SlaveMessage, port int) {
	go bcast.Transmitter(port, c_transmitMessage)
}
func F_TransmitMasterMessage(c_transmitMessage chan T_MasterMessage, port int) {
	go bcast.Transmitter(port, c_transmitMessage)
}

func F_ReceiveSlaveMessage(c_verifiedMessage chan T_SlaveMessage, ops T_NodeOperations, c_newConnectedNodes chan []T_NodeInfo, port int) {
	c_receive := make(chan T_SlaveMessage)

	go bcast.Receiver(port, c_receive)
	for {
		select {
		case receivedMessage := <-c_receive:
			if f_AcceptancetestSM() { //FUTURE ACCEPTANCETEST
				c_verifiedMessage <- receivedMessage
			}
		}
	}
}

func F_ReceiveMasterMessage(c_verifiedMessage chan T_MasterMessage, ops T_NodeOperations, c_newConnectedNodes chan []T_NodeInfo, port int) {
	c_receive := make(chan T_MasterMessage)

	go bcast.Receiver(port, c_receive)
	for {
		select {
		case receivedMessage := <-c_receive:
			if f_AcceptancetestMM() { //FUTURE ACCEPTANCETEST
				c_verifiedMessage <- receivedMessage
			}
		}
	}
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
