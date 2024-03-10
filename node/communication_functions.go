package node

import (
	//"fmt"
	"encoding/json"
	"fmt"
	"hash/crc32"
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

/*
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

func F_ReceiveSlaveMessage(c_verifiedMessage chan T_SlaveMessage, port int, c_quit chan bool) {
	c_receive := make(chan T_SlaveMessage)

	go bcast.Receiver(port, c_quit, c_receive)
	for {
		select {
		case <-c_quit:
			F_WriteLog("Closed f_ReceiveSlaveMessage")
			return
		case receivedMessage := <-c_receive:
			if f_AcceptancetestSM() { //FUTURE ACCEPTANCETEST
				c_verifiedMessage <- receivedMessage
			}
		}
	}
}

func F_ReceiveMasterMessage(c_verifiedMessage chan T_MasterMessage, port int, c_quit chan bool) {
	c_receive := make(chan T_MasterMessage)

	go bcast.Receiver(port, c_quit, c_receive)
	for {
		select {
		case <-c_quit:
			F_WriteLog("Closed f_ReceiveMasterMessage")
			return
		case receivedMessage := <-c_receive:
			if f_AcceptancetestMM() { //FUTURE ACCEPTANCETEST
				c_verifiedMessage <- receivedMessage
			}
		}
	}
}
*/

var messagesToSend int = 3

func f_calculateChecksum(data []byte) uint32 {
	crc32Table := crc32.MakeTable(crc32.Castagnoli)
	checksum := crc32.Checksum(data, crc32Table)
	return checksum
}

func f_convertSlaveMessageToCS(messageSlaveMessage T_SlaveMessage) uint32 {
	messageSlaveMessage.Checksum = 0
	messageString, err := json.Marshal(messageSlaveMessage)
	if err != nil {
		fmt.Println("Not able to Marshal SlaveMessage")
		return 0
	}
	messageByte := []byte(messageString)
	checksum := f_calculateChecksum(messageByte)
	return checksum
}

func f_convertMasterMessageToCS(messageMasterMessage T_MasterMessage) uint32 {
	messageMasterMessage.Checksum = 0
	messageString, err := json.Marshal(messageMasterMessage)
	if err != nil {
		fmt.Println("Not able to Marshal MasterMessage")
		return 0
	}

	messageByte := []byte(messageString)
	checksum := f_calculateChecksum(messageByte)
	return checksum
}

// the AcceptanceTest also includes being able to demarshal the message, but having come here is it already demarshaled
func f_PassedAcceptancetestMM(receivedMessage T_MasterMessage) bool {
	return true //receivedMessage.Checksum == f_convertMasterMessageToCS(receivedMessage)
}

func f_PassedAcceptancetestSM(receivedMessage T_SlaveMessage) bool {
	return true //receivedMessage.Checksum == f_convertSlaveMessageToCS(receivedMessage)
}

func F_TransmitSlaveMessage(c_transmitSlaveMessage chan T_SlaveMessage, port int) {
	c_slaveMessageWithCS := make(chan T_SlaveMessage)
	go bcast.Transmitter(port, c_slaveMessageWithCS)
	for {
		select {
		case transmitSlaveMessage := <-c_transmitSlaveMessage:
			transmitSlaveMessage.Checksum = f_convertSlaveMessageToCS(transmitSlaveMessage)
			for i := 0; i < messagesToSend; i++ {
				c_slaveMessageWithCS <- transmitSlaveMessage
				time.Sleep(time.Duration(10) * time.Millisecond)
			}
		}
	}
}
func F_TransmitMasterMessage(c_transmitMasterMessage chan T_MasterMessage, port int) {
	c_masterMessageWithCS := make(chan T_MasterMessage)
	go bcast.Transmitter(port, c_masterMessageWithCS)
	for {
		select {
		case transmitMasterMessage := <-c_transmitMasterMessage:
			transmitMasterMessage.Checksum = f_convertMasterMessageToCS(transmitMasterMessage)
			for i := 0; i < messagesToSend; i++ {
				c_masterMessageWithCS <- transmitMasterMessage
				time.Sleep(time.Duration(10) * time.Millisecond)
			}
		}
	}
}

func F_ReceiveSlaveMessage(c_verifiedMessage chan T_SlaveMessage, port int, c_quit chan bool) {
	c_receive := make(chan T_SlaveMessage)
	go bcast.Receiver(port, c_quit, c_receive)
	for {
		select {
		case receivedMessage := <-c_receive:
			if f_PassedAcceptancetestSM(receivedMessage) {
				c_verifiedMessage <- receivedMessage
			} else {
				//fmt.Println("Slavemessage not verified: %v", receivedMessage)
			}
		}
	}
}

func F_ReceiveMasterMessage(c_verifiedMessage chan T_MasterMessage, port int, c_quit chan bool) {
	c_receive := make(chan T_MasterMessage)
	go bcast.Receiver(port, c_quit, c_receive)
	for {
		select {
		case receivedMessage := <-c_receive:
			if f_PassedAcceptancetestMM(receivedMessage) {
				c_verifiedMessage <- receivedMessage
			} else {
				//fmt.Println("Mastermessage not verified: %v", receivedMessage)
			}
		}
	}
}
