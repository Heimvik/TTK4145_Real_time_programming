package node

import (
	"fmt"
	"hash/crc32"
	"encoding/json"
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
 

func f_calculateChecksum(data []byte) uint32 {
    crc32Table := crc32.MakeTable(crc32.Castagnoli)
    checksum := crc32.Checksum(data, crc32Table)
    return checksum
}

func f_convertSlaveMessageToCS(messageSlaveMessage T_SlaveMessage) uint32 {
	messageSlaveMessage.Checksum = 0
	messageString, err := json.Marshal(messageSlaveMessage)
	if err != nil{
		fmt.Println("Not able to Marshal SlaveMessage")
		return 0
	}
	messageByte := []byte(messageString)
	checksum := f_calculateChecksum(messageByte)
	return checksum
}

func f_convertMasterMessageToCS(messageMasterMessage T_MasterMessage) uint32  { 
	messageMasterMessage.Checksum = 0
	messageString, err := json.Marshal(messageMasterMessage)
	if err != nil{
		fmt.Println("Not able to Marshal MasterMessage")
		return 0
	}

	messageByte := []byte(messageString)
	checksum := f_calculateChecksum(messageByte)
	return checksum
	
}

func f_AcceptancetestSM(receivedSlaveMessage T_SlaveMessage) bool {
	calculatedCS := f_convertSlaveMessageToCS(receivedSlaveMessage)
	receivedCS := receivedSlaveMessage.Checksum
	return calculatedCS == receivedCS
}

func f_AcceptancetestMM(receivedMasterMessage T_MasterMessage) bool {
	calculatedCS := f_convertMasterMessageToCS(receivedMasterMessage)
	receivedCS := receivedMasterMessage.Checksum
	return calculatedCS == receivedCS
}

func f_compareSlaveMessages(messages []T_SlaveMessage) int {
	if messages[0] == messages[1] || messages[0] == messages[2]{
		return 0
	}else if messages[1] == messages[2]{
		return 1
	}
	return 2
}

func f_compareMasterMessages(messages []T_MasterMessage) int {
	if messages[0].Transmitter == messages[1].Transmitter || messages[0].Transmitter == messages[2].Transmitter{
		if f_compareGlobalQueue(messages[0].GlobalQueue, messages[1].GlobalQueue) || f_compareGlobalQueue(messages[0].GlobalQueue, messages[2].GlobalQueue){
			return 0
		}
	}else if messages[1].Transmitter == messages[2].Transmitter{
		if f_compareGlobalQueue(messages[1].GlobalQueue, messages[2].GlobalQueue){
			return 1
		}
	}
	return 2
}

func f_compareGlobalQueue(q1 []T_GlobalQueueEntry, q2 []T_GlobalQueueEntry) bool{
	if len(q1) != len(q2){
		return false
	}

	for i := 0 ; i < len(q1) ; i++{
		if q1[i] != q2[i]{
			return false
		}
	}
	return true
}

func F_TransmitSlaveMessage(c_transmitSlaveMessage chan T_SlaveMessage, port int) {
	var c_slaveMessageWithCS chan T_SlaveMessage
	go bcast.Transmitter(port, c_slaveMessageWithCS)

	for{
		select{
		case transmitSlaveMessage := <- c_transmitSlaveMessage:
			transmitSlaveMessage.Checksum = f_convertSlaveMessageToCS(transmitSlaveMessage)
			for i := 0; i < 10; i++ {
				c_slaveMessageWithCS <- transmitSlaveMessage
			}
		}
	}
}

func F_TransmitMasterMessage(c_transmitMasterMessage chan T_MasterMessage, port int) {
	var c_masterMessageWithCS chan T_MasterMessage
	go bcast.Transmitter(port, c_masterMessageWithCS) 

	for{
		select{
		case transmitMasterMessage := <- c_transmitMasterMessage:
			transmitMasterMessage.Checksum = f_convertMasterMessageToCS(transmitMasterMessage)
			for i := 0; i < 10; i++ {
				c_masterMessageWithCS <- transmitMasterMessage
			}
		}
	}
}

func F_ReceiveSlaveMessage(c_verifiedMessage chan T_SlaveMessage, ops T_NodeOperations, port int) {
	c_receive := make(chan T_SlaveMessage)
	go bcast.Receiver(port, c_receive)

	approvedMessages := []T_SlaveMessage{}

	for i := 0 ; i < 11 ; i++ {
		select {
		case receivedMessage := <-c_receive:
			if f_AcceptancetestSM(receivedMessage) && len(approvedMessages) < 3{ //FUTURE ACCEPTANCETEST
				approvedMessages = append(approvedMessages, receivedMessage)
			}else if len(approvedMessages) == 3{
				c_verifiedMessage <- approvedMessages[f_compareSlaveMessages(approvedMessages)]
				approvedMessages = []T_SlaveMessage{}
			}
		default:
			i = 0
		}
	}
}

func F_ReceiveMasterMessage(c_verifiedMessage chan T_MasterMessage, ops T_NodeOperations, port int) {
	c_receive := make(chan T_MasterMessage)
	go bcast.Receiver(port, c_receive)
	
	approvedMessages := []T_MasterMessage{}

	for {
		select {
		case receivedMessage := <-c_receive:
			if f_AcceptancetestMM(receivedMessage) { //FUTURE ACCEPTANCETEST
				c_verifiedMessage <- receivedMessage
			}else if len(approvedMessages) == 3{
				c_verifiedMessage <- approvedMessages[f_compareMasterMessages(approvedMessages)]
				approvedMessages = []T_MasterMessage{}
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
