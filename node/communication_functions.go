package node

import (
	//"fmt"

	"the-elevator/network/network_libraries/bcast"
	"time"
)

var messagesToSend int = 20

func F_TransmitSlaveMessage(c_transmitSlaveMessage chan T_SlaveMessage, port int) {
	c_slaveMessageOut := make(chan T_SlaveMessage)
	go bcast.Transmitter(port, c_slaveMessageOut)
	for {
		select {
		case transmitSlaveMessage := <-c_transmitSlaveMessage:
			for i := 0; i < messagesToSend; i++ {
				c_slaveMessageOut <- transmitSlaveMessage
				time.Sleep(time.Duration(2) * time.Millisecond)
			}
		}
	}
}
func F_TransmitMasterMessage(c_transmitMasterMessage chan T_MasterMessage, port int) {
	c_masterMessageOut := make(chan T_MasterMessage)
	go bcast.Transmitter(port, c_masterMessageOut)
	for {
		select {
		case transmitMasterMessage := <-c_transmitMasterMessage:
			for i := 0; i < messagesToSend; i++ {
				c_masterMessageOut <- transmitMasterMessage
				time.Sleep(time.Duration(2) * time.Millisecond)
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
			c_verifiedMessage <- receivedMessage
		}
	}
}

func F_ReceiveMasterMessage(c_verifiedMessage chan T_MasterMessage, port int, c_quit chan bool) {
	c_receive := make(chan T_MasterMessage)
	go bcast.Receiver(port, c_quit, c_receive)
	for {
		select {
		case receivedMessage := <-c_receive:
			c_verifiedMessage <- receivedMessage
		}
	}
}
