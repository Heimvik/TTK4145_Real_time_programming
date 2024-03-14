package node

import (
	//"fmt"

	"the-elevator/network/network_libraries/bcast"
	"time"
)

var MESSAGES_TO_SEND int = 30

/*
Broadcasts the current global queue and node information as a master message to connected nodes, ensuring network-wide state consistency.

Prerequisites: An initialized global queue and node information.

Returns: Nothing, but triggers the broadcast of the master node's state to connected nodes.
*/
func f_TransmitMasterInfo(c_transmitMasterMessage chan T_MasterMessage) {
	globalQueue := f_GetGlobalQueue()
	masterMessage := T_MasterMessage{
		Transmitter: f_GetNodeInfo(),
		GlobalQueue: f_CopyGlobalQueue(globalQueue),
	}
	c_transmitMasterMessage <- masterMessage
}

/*
Sends out a slave message with the transmitter node's information, allowing the master node to update its view of the network.

Prerequisites: Node information must be initialized.

Returns: Nothing, but sends the slave node's current state to the master node for network synchronization.
*/
func f_TransmitSlaveInfo(c_transmitSlaveMessage chan T_SlaveMessage) {
	transmitter := f_GetNodeInfo()
	slaveMessage := T_SlaveMessage{
		Transmitter: transmitter,
		Entry:       T_GlobalQueueEntry{},
	}
	c_transmitSlaveMessage <- slaveMessage
}

/*
Listens for slave messages on its channel and broadcasts them, ensuring the master node and other nodes are updated. Only transmits when new messages are received, maintaining network synchronization.

Prerequisites: An initialized channel for receiving slave messages to broadcast and a configured network port for broadcasting.

Returns: Nothing, but broadcasts slave messages upon receiving new messages through its channel.
*/
func f_TransmitSlaveMessage(c_transmitSlaveMessage chan T_SlaveMessage, port int) {
	c_slaveMessageNodes := make(chan T_SlaveMessage)
	c_slaveMessageLocal := make(chan T_SlaveMessage)
	go bcast.Transmitter("255.255.255.255", port, c_slaveMessageNodes)
	go bcast.Transmitter("localhost", port, c_slaveMessageLocal)
	//Source: https://github.com/TTK4145/Network-go
	for {
		select {
		case transmitSlaveMessage := <-c_transmitSlaveMessage:
			for i := 0; i < MESSAGES_TO_SEND; i++ {
				c_slaveMessageNodes <- transmitSlaveMessage
				time.Sleep(time.Duration(1) * time.Millisecond)
				c_slaveMessageNodes <- transmitSlaveMessage
				time.Sleep(time.Duration(2) * time.Millisecond)
			}
		}
	}
}

/*
Listens for master messages on its channel and broadcasts them to connected nodes, only transmitting when new messages are received to ensure network-wide state consistency.

Prerequisites: An initialized channel for receiving master messages to broadcast and a configured network port for broadcasting.

Returns: Nothing, but broadcasts master messages to slave nodes upon receiving new messages through its channel.
*/
func f_TransmitMasterMessage(c_transmitMasterMessage chan T_MasterMessage, port int) {
	c_masterMessageNodes := make(chan T_MasterMessage)
	c_masterMessageLocal := make(chan T_MasterMessage)
	go bcast.Transmitter("255.255.255.255", port, c_masterMessageNodes)
	go bcast.Transmitter("localhost", port, c_masterMessageLocal)
	//Source: https://github.com/TTK4145/Network-go
	for {
		select {
		case transmitMasterMessage := <-c_transmitMasterMessage:
			for i := 0; i < MESSAGES_TO_SEND; i++ {
				c_masterMessageNodes <- transmitMasterMessage
				time.Sleep(time.Duration(1) * time.Millisecond)
				c_masterMessageLocal <- transmitMasterMessage
				time.Sleep(time.Duration(2) * time.Millisecond)
			}
		}
	}
}

/*
Initiates a background process to listen for slave messages on a specified port, directly forwarding them to a designated channel for processing.

Prerequisites: Network configured to listen on the specified port and a mechanism to gracefully handle termination signals.

Returns: Nothing, but initiates continuous listening for slave messages, which are then passed to the system for processing.
*/
func f_ReceiveSlaveMessage(c_receivedSlaveMessage chan T_SlaveMessage, port int, c_quit chan bool) {
	go bcast.Receiver(port, c_quit, c_receivedSlaveMessage)
	//Source: https://github.com/TTK4145/Network-go
}

/*
Initiates a background process to listen for master messages on a specified port, directly forwarding them to a designated channel for processing.

Prerequisites: Network configured to listen on the specified port and a mechanism to gracefully handle termination signals.

Returns: Nothing, but initiates continuous listening for master messages, which are then passed to the system for processing.
*/
func f_ReceiveMasterMessage(c_receivedMasterMessage chan T_MasterMessage, port int, c_quit chan bool) {
	go bcast.Receiver(port, c_quit, c_receivedMasterMessage)
	//Source: https://github.com/TTK4145/Network-go
}
