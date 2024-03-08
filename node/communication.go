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

func F_ReceiveSlaveMessage(c_verifiedMessage chan T_SlaveMessage, port int) {
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

func F_ReceiveMasterMessage(c_verifiedMessage chan T_MasterMessage, port int) {
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
