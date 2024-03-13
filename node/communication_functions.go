package node

import (
	//"fmt"

	"the-elevator/network/network_libraries/bcast"
	"time"
)

var MESSAGES_TO_SEND int = 20

func f_TransmitMasterInfo(c_transmitMasterMessage chan T_MasterMessage) {
	globalQueue := f_GetGlobalQueue()
	masterMessage := T_MasterMessage{
		Transmitter: f_GetNodeInfo(),
		GlobalQueue: f_CopyGlobalQueue(globalQueue),
	}
	c_transmitMasterMessage <- masterMessage
}

func f_TransmitSlaveInfo(c_transmitSlaveMessage chan T_SlaveMessage) {
	transmitter := f_GetNodeInfo()
	slaveMessage := T_SlaveMessage{
		Transmitter: transmitter,
		Entry:       T_GlobalQueueEntry{},
	}
	c_transmitSlaveMessage <- slaveMessage
}

func f_TransmitSlaveMessage(c_transmitSlaveMessage chan T_SlaveMessage, port int) {
	c_slaveMessageOut := make(chan T_SlaveMessage)
	//Source:
	go bcast.Transmitter(port, c_slaveMessageOut)
	for {
		select {
		case transmitSlaveMessage := <-c_transmitSlaveMessage:
			for i := 0; i < MESSAGES_TO_SEND; i++ {
				c_slaveMessageOut <- transmitSlaveMessage
				time.Sleep(time.Duration(3) * time.Millisecond)
			}
		}
	}
}
func f_TransmitMasterMessage(c_transmitMasterMessage chan T_MasterMessage, port int) {
	c_masterMessageOut := make(chan T_MasterMessage)
	//Source: 
	go bcast.Transmitter(port, c_masterMessageOut)
	for {
		select {
		case transmitMasterMessage := <-c_transmitMasterMessage:
			for i := 0; i < MESSAGES_TO_SEND; i++ {
				c_masterMessageOut <- transmitMasterMessage
				time.Sleep(time.Duration(3) * time.Millisecond)
			}
		}
	}
}

func f_ReceiveSlaveMessage(c_verifiedMessage chan T_SlaveMessage, port int, c_quit chan bool) {
	c_receive := make(chan T_SlaveMessage)
	go bcast.Receiver(port, c_quit, c_receive)
	for {
		select {
		case receivedMessage := <-c_receive:
			c_verifiedMessage <- receivedMessage
		}
	}
}

func f_ReceiveMasterMessage(c_verifiedMessage chan T_MasterMessage, port int, c_quit chan bool) {
	c_receive := make(chan T_MasterMessage)
	go bcast.Receiver(port, c_quit, c_receive)
	for {
		select {
		case receivedMessage := <-c_receive:
			c_verifiedMessage <- receivedMessage
		}
	}
}
