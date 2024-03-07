package node

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"the-elevator/network/network_libraries/bcast"
	"the-elevator/node/elevator"
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

func f_GlobalQueueAreEqual(q1 []T_GlobalQueueEntry, q2 []T_GlobalQueueEntry) bool {
	if len(q1) != len(q2) {
		return false
	}

	for i := 0; i < len(q1); i++ {
		if q1[i] != q2[i] {
			return false
		}
	}
	return true
}

func f_MasterMessagesAreEqual(singleMessage T_MasterMessage, messageList []T_MasterMessage) bool {
	for _, message := range messageList {
		if !f_GlobalQueueAreEqual(singleMessage.GlobalQueue, message.GlobalQueue) {
			return false
		}
	}
	return true
}

func F_TransmitSlaveMessage(c_transmitSlaveMessage chan T_SlaveMessage, port int) {
	c_slaveMessageWithCS := make(chan T_SlaveMessage)
	go bcast.Transmitter(port, c_slaveMessageWithCS)

	for {
		select {
		case transmitSlaveMessage := <-c_transmitSlaveMessage:
			transmitSlaveMessage.Checksum = f_convertSlaveMessageToCS(transmitSlaveMessage)
			for i := 0; i < 10; i++ {
				c_slaveMessageWithCS <- transmitSlaveMessage
				time.Sleep(time.Duration(100) * time.Millisecond)
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
			for i := 0; i < 50; i++ {
				c_masterMessageWithCS <- transmitMasterMessage
				time.Sleep(time.Duration(1) * time.Millisecond)
			}
		}
	}
}

func F_ReceiveSlaveMessage(c_verifiedMessage chan T_SlaveMessage, port int) {
	c_receive := make(chan T_SlaveMessage)

	go bcast.Receiver(port, c_receive)
	var currentCSmap map[uint32][]T_SlaveMessage
	for {
		select {
		case receivedMessage := <-c_receive:
			if len(currentCSmap[receivedMessage.Checksum]) < 50 {
				currentCSmap[receivedMessage.Checksum] = append(currentCSmap[receivedMessage.Checksum], receivedMessage)
			}
		default:
			var verifiedCS uint32
			shouldResetMap := false

			for currentCS, equalMMs := range currentCSmap {
				if len(equalMMs) == 10 {
					fmt.Printf("10 messages with equal checksum found: Verified\n", len(equalMMs))
					c_verifiedMessage <- equalMMs[0]

					for len(equalMMs) < 50 {
						equalMMs = append(equalMMs, equalMMs[0])
					}

					currentCSmap[currentCS] = equalMMs
					verifiedCS = currentCS
					shouldResetMap = true
					break
				}
			}

			if shouldResetMap {
				newCSmap := make(map[uint32][]T_SlaveMessage)
				newCSmap[verifiedCS] = currentCSmap[verifiedCS]
				currentCSmap = newCSmap
			}
		}
	}
}

func (m T_MasterMessage) String() string {
	// Example implementation. Adjust fields as needed.
	return fmt.Sprintf("Transmitter: %+v, Checksum: %d, GlobalQueue Length: %d",
		m.Transmitter, m.Checksum, len(m.GlobalQueue))
}
func logCurrentCSmap(currentCSmap map[uint32][]T_MasterMessage) {
	for cs, messages := range currentCSmap {
		fmt.Printf("Checksum: %d, Messages Count: %d\n", cs, len(messages))
		for i, message := range messages {
			fmt.Printf("\tMessage %d: %s\n", i+1, message.String())
		}
	}
}
func F_ReceiveMasterMessage(c_verifiedMessage chan T_MasterMessage, port int) {
	c_receive := make(chan T_MasterMessage)

	go bcast.Receiver(port, c_receive)
	currentCSmap := make(map[uint32][]T_MasterMessage)
	for {
		select {
		case receivedMessage := <-c_receive:
			if len(currentCSmap[receivedMessage.Checksum]) < 50 {
				currentCSmap[receivedMessage.Checksum] = append(currentCSmap[receivedMessage.Checksum], receivedMessage)
				logCurrentCSmap(currentCSmap)
			}
		default:
			var verifiedCS uint32
			shouldResetMap := false

			for currentCS, equalMMs := range currentCSmap {
				if len(equalMMs) == 10 && f_MasterMessagesAreEqual(equalMMs[0], equalMMs) {
					fmt.Printf("10 messages with equal checksum found: Verified\n", len(equalMMs))
					c_verifiedMessage <- equalMMs[0]

					for len(equalMMs) < 50 {
						equalMMs = append(equalMMs, equalMMs[0])
					}
					logCurrentCSmap(currentCSmap)

					currentCSmap[currentCS] = equalMMs
					verifiedCS = currentCS
					shouldResetMap = true
					break
				}
			}

			if shouldResetMap {
				newCSmap := make(map[uint32][]T_MasterMessage)
				newCSmap[verifiedCS] = currentCSmap[verifiedCS]
				currentCSmap = newCSmap
			}
		}
	}
}

func F_TestCommunication() {
	c_receiveSlaveMessage := make(chan T_SlaveMessage)
	c_receiveMasterMessage := make(chan T_MasterMessage)
	c_transmitMasterMessage := make(chan T_MasterMessage)
	c_transmitSlaveMessage := make(chan T_SlaveMessage)

	go F_ReceiveSlaveMessage(c_receiveSlaveMessage, SLAVEPORT)
	go F_ReceiveMasterMessage(c_receiveMasterMessage, MASTERPORT)
	go F_TransmitSlaveMessage(c_transmitSlaveMessage, SLAVEPORT)
	go F_TransmitMasterMessage(c_transmitMasterMessage, MASTERPORT)

	go func() {
		for i := 1; i < 10; i++ {
			var masterMessage T_MasterMessage
			masterMessage.Transmitter = T_NodeInfo{}
			masterMessage.GlobalQueue = make([]T_GlobalQueueEntry, 0)

			entry := T_GlobalQueueEntry{
				Request: elevator.T_Request{
					Id:        uint16(i),
					State:     elevator.UNASSIGNED,
					Calltype:  5,
					Floor:     4,
					Direction: 2,
				},
				RequestedNode:     2,
				AssignedNode:      0,
				TimeUntilReassign: 15,
			}
			masterMessage.GlobalQueue = append(masterMessage.GlobalQueue, entry)

			c_transmitMasterMessage <- masterMessage

			time.Sleep(time.Duration(15) * time.Second)

		}
	}()
	for {
		select {
		case masterMessage := <-c_receiveMasterMessage:
			fmt.Printf("%+v\n", masterMessage)
		}
	}
}
