package node

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"the-elevator/network/network_libraries/bcast"
	"the-elevator/node/elevator"
	"time"
	"math/rand"
)

// KILDE:

//This should:
//Take in:
// - channel to put the message in
// - port
//Give out:
// - Object of T_Message
// - Array of connected nodes (any unconnected nodes)

type OrderedMasterMap struct { //Used to keep track of the last 100 messages
	keys   []uint32
	values []T_MasterMessage
}

type OrderedSlaveMap struct { //Used to keep track of the last 100 messages
	keys   []uint32
	values []T_SlaveMessage
}

func (m *OrderedMasterMap) Add(key uint32, value T_MasterMessage) {
    for _, existingKey := range m.keys {
        if existingKey == key {
            return
        }
    }

    if len(m.keys) == 100 {
        m.keys = m.keys[1:]
        m.values = m.values[1:]
    }

    m.keys = append(m.keys, key)
    m.values = append(m.values, value)
}

func (m *OrderedMasterMap) Exists(key uint32) bool {
    for _, existingKey := range m.keys {
        if existingKey == key {
            return true
        }
    }
    return false
}

func (m *OrderedMasterMap) Print() {
    for i, key := range m.keys {
        fmt.Printf("Key: %v, Value: %v\n", key, m.values[i])
    }
}

func (m *OrderedSlaveMap) Add(key uint32, value T_SlaveMessage) {
    for _, existingKey := range m.keys {
        if existingKey == key {
            return
        }
    }

    if len(m.keys) == 100 {
        m.keys = m.keys[1:]
        m.values = m.values[1:]
    }

    m.keys = append(m.keys, key)
    m.values = append(m.values, value)
}

func (m *OrderedSlaveMap) Exists(key uint32) bool {
    for _, existingKey := range m.keys {
        if existingKey == key {
            return true
        }
    }
    return false
}

func (m *OrderedSlaveMap) Print() {
    for i, key := range m.keys {
        fmt.Printf("Key: %v, Value: %v\n", key, m.values[i])
    }
}

var transmittedMasterMessages = OrderedMasterMap{
	keys:   make([]uint32, 0),
	values: []T_MasterMessage{},
}

var receivedMasterMessages = OrderedMasterMap{
	keys:   make([]uint32, 0),
	values: []T_MasterMessage{},
}

var verifiedMasterMessages = OrderedMasterMap{
	keys:   make([]uint32, 0),
	values: []T_MasterMessage{},
}

var transmittedSlaveMessages = OrderedSlaveMap{
	keys:   make([]uint32, 0),
	values: []T_SlaveMessage{},
}

var receivedSlaveMessages = OrderedSlaveMap{
	keys:   make([]uint32, 0),
	values: []T_SlaveMessage{},
}

var verifiedSlaveMessages = OrderedSlaveMap{
	keys:   make([]uint32, 0),
	values: []T_SlaveMessage{},
}

func (m T_MasterMessage) String() string {
	// Example implementation. Adjust fields as needed.
	return fmt.Sprintf("Transmitter: %+v, Checksum: %d, GlobalQueue Length: %d",
		m.Transmitter, m.Checksum, len(m.GlobalQueue))
}
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

func f_ArboAmountOfEqualMasterMessages(messageList []T_MasterMessage) (int, int) { //Returns the amount of equal messages and the index of the message with the highest amount of equal messages
	var counter int
	highestCounter := 0
	highestCounterIndex := 0

	for i := 0; i < len(messageList); i++ {
		counter = 0
		for j := 0; j < len(messageList); j++ {
			if messageList[i].Transmitter == messageList[j].Transmitter && f_GlobalQueueAreEqual(messageList[i].GlobalQueue, messageList[j].GlobalQueue) {
				counter++
			}
		}
		if counter > highestCounter {
			highestCounter = counter
			highestCounterIndex = i
		}
	}
	return highestCounter, highestCounterIndex
}

func f_ArboAmountOfEqualSlaveMessages(messageList []T_SlaveMessage) (int, int) { //Returns the amount of equal messages and the index of the message with the highest amount of equal messages
	var counter int
	highestCounter := 0
	highestCounterIndex := 0

	for i := 0; i < len(messageList); i++ {
		counter = 0
		for j := 0; j < len(messageList); j++ {
			if messageList[i].Transmitter == messageList[j].Transmitter && messageList[i].Entry == messageList[j].Entry {
				counter++
			}
		}
		if counter > highestCounter {
			highestCounter = counter
			highestCounterIndex = i
		}
	}
	return highestCounter, highestCounterIndex
}

func logCurrentCSmap(currentCSmap map[uint32][]T_MasterMessage) {
	for cs, messages := range currentCSmap {
		fmt.Printf("Checksum: %d, Messages Count: %d\n", cs, len(messages))
		for i, message := range messages {
			fmt.Printf("\tMessage %d: %s\n", i+1, message.String())
		}
	}
}

func f_printOrderedMasterMaps() {
	fmt.Printf("\nTransmitted Master Messages: %v in total\n", len(transmittedMasterMessages.keys))
	transmittedMasterMessages.Print()

	fmt.Printf("\nReceived Master Messages: %v in total\n", len(receivedMasterMessages.keys))
	receivedMasterMessages.Print()

	fmt.Printf("\nVerified Master Messages: %v in total\n", len(verifiedMasterMessages.keys))
	verifiedMasterMessages.Print()
}

func f_printOrderedSlaveMaps() {
	fmt.Printf("\nTransmitted Slave Messages: %v in total\n", len(transmittedSlaveMessages.keys))
	transmittedSlaveMessages.Print()

	fmt.Printf("\nReceived Slave Messages: %v in total\n", len(receivedSlaveMessages.keys))
	receivedSlaveMessages.Print()

	fmt.Printf("\nVerified Slave Messages: %v in total\n", len(verifiedSlaveMessages.keys))
	verifiedSlaveMessages.Print()
}	

func F_TransmitSlaveMessage(c_transmitSlaveMessage chan T_SlaveMessage, port int, c_quit chan bool) {
	c_slaveMessageWithCS := make(chan T_SlaveMessage)
	go bcast.Transmitter(port, c_slaveMessageWithCS)

	for {
		select {
		case transmitSlaveMessage := <-c_transmitSlaveMessage:
			transmitSlaveMessage.Checksum = f_convertSlaveMessageToCS(transmitSlaveMessage)
			for i := 0; i < 20; i++ {
				//fmt.Printf("Transmitting slaveMessage: %d\n", transmitSlaveMessage)
				c_slaveMessageWithCS <- transmitSlaveMessage
				transmittedSlaveMessages.Add(transmitSlaveMessage.Checksum, transmitSlaveMessage)
				time.Sleep(100 * time.Millisecond)
			}
		
		case <-c_quit:
			fmt.Println("SLAVE TRANSMITTER:\t Quitting")
			time.Sleep(500 * time.Millisecond)
			return
		}
	}
}

func F_TransmitMasterMessage(c_transmitMasterMessage chan T_MasterMessage, port int, c_quit chan bool) {
	c_masterMessageWithCS := make(chan T_MasterMessage)
	go bcast.Transmitter(port, c_masterMessageWithCS)

	for {
		select {
		case transmitMasterMessage := <-c_transmitMasterMessage:
			transmitMasterMessage.Checksum = f_convertMasterMessageToCS(transmitMasterMessage)
			for i := 0; i < 20; i++ {
				//fmt.Printf("Transmitting masterMessage: %d\n", transmitMasterMessage)
				c_masterMessageWithCS <- transmitMasterMessage
				transmittedMasterMessages.Add(transmitMasterMessage.Checksum, transmitMasterMessage)
				time.Sleep(1 * time.Millisecond)
			}
		case <-c_quit:
			fmt.Println("MASTER TRANSMITTER:\t Quitting")
			time.Sleep(500 * time.Millisecond)
			return
		}
	}
}


func F_ReceiveSlaveMessage(c_verifiedMessage chan T_SlaveMessage, port int, c_quit chan bool) {
	c_receive := make(chan T_SlaveMessage)
	go bcast.Receiver(port, c_receive)
	currentCSmap := make(map[uint32][]T_SlaveMessage)

	messageAlreadySentmap := OrderedSlaveMap{
		keys:   make([]uint32, 0),
		values: []T_SlaveMessage{},
	}

	ticker := time.NewTicker(time.Millisecond * 10) // check for new messages every 500 milliseconds
	defer ticker.Stop()

	for {
		select {
		case receivedMessage := <-c_receive:
			if !messageAlreadySentmap.Exists(receivedMessage.Checksum) {
				fmt.Printf("Received SlaveMessage: %d\n", receivedMessage)
				currentCSmap[receivedMessage.Checksum] = append(currentCSmap[receivedMessage.Checksum], receivedMessage)
				receivedSlaveMessages.Add(receivedMessage.Checksum, receivedMessage)
			}

		case <-ticker.C: // check for new messages every 500 milliseconds
			for checksum, messages := range currentCSmap {
				counter, index := f_ArboAmountOfEqualSlaveMessages(messages)
				if counter >= 3 {
					fmt.Printf("%d identical SlaveMessages found: Verified\n", counter)
					verifiedSlaveMessages.Add(checksum, messages[index])

					c_verifiedMessage <- messages[index]
					messageAlreadySentmap.Add(checksum, messages[index])
					delete(currentCSmap, checksum)
				}
			}

		case <-c_quit:
			fmt.Println("SLAVE RECEIVER:\t Quitting")
			time.Sleep(500 * time.Millisecond)
			return
		}
	}
}

func F_ReceiveMasterMessage(c_verifiedMessage chan T_MasterMessage, port int, c_quit chan bool) {
	c_receive := make(chan T_MasterMessage)
	go bcast.Receiver(port, c_receive)
	currentCSmap := make(map[uint32][]T_MasterMessage)

	messageAlreadySentmap := OrderedMasterMap{
		keys:   make([]uint32, 0),
		values: []T_MasterMessage{},
	}

	ticker := time.NewTicker(time.Millisecond * 10) // check for new messages every 500 milliseconds
	defer ticker.Stop()

	for {
		select {
		case receivedMessage := <-c_receive:
			if !messageAlreadySentmap.Exists(receivedMessage.Checksum) {
				fmt.Printf("Received MasterMessage: %d\n", receivedMessage)
				currentCSmap[receivedMessage.Checksum] = append(currentCSmap[receivedMessage.Checksum], receivedMessage)
				receivedMasterMessages.Add(receivedMessage.Checksum, receivedMessage)
			}

		case <-ticker.C: // check for new messages every 500 milliseconds
			for checksum, messages := range currentCSmap {
				counter, index := f_ArboAmountOfEqualMasterMessages(messages)
				if counter >= 3 {
					fmt.Printf("%d identical MasterMessages found: Verified\n", counter)
					verifiedMasterMessages.Add(checksum, messages[index])

					c_verifiedMessage <- messages[index]
					messageAlreadySentmap.Add(checksum, messages[index])
					delete(currentCSmap, checksum)
	
				}
			}
		case <-c_quit:
			fmt.Println("MASTER RECEIVER:\t Quitting")
			time.Sleep(10 * time.Second)
			return
		}
	}
}

func F_ArboTestCommunication() {
	c_verifiedMasterMessage := make(chan T_MasterMessage)
	c_transmitMasterMessage := make(chan T_MasterMessage)
	c_verifiedSlaveMessage := make(chan T_SlaveMessage)
	c_transmitSlaveMessage := make(chan T_SlaveMessage)
	c_quit := make(chan bool)
	
	go F_ReceiveMasterMessage(c_verifiedMasterMessage, MASTERPORT, c_quit)
	go F_TransmitMasterMessage(c_transmitMasterMessage, MASTERPORT, c_quit)
	go F_ReceiveSlaveMessage(c_verifiedSlaveMessage, SLAVEPORT, c_quit)
	go F_TransmitSlaveMessage(c_transmitSlaveMessage, SLAVEPORT, c_quit)

	quitTimer := time.NewTicker(7 * time.Second)

	globalQ1 := []T_GlobalQueueEntry{}
	//globalQ2 := []T_GlobalQueueEntry{}

	sentToMasterTransmit := []T_MasterMessage{}
	sentToSlaveTransmit := []T_SlaveMessage{}

	rand.Seed(time.Now().UnixNano())

	go func() {

		for i := 1; i < 5; i++ {
			var masterMessage1 T_MasterMessage
			var slaveMessage1 T_SlaveMessage

			masterMessage1.Transmitter = T_NodeInfo{PRIORITY: uint8(rand.Intn(10)), 
													Role: T_NodeRole(uint8(rand.Intn(10))), 
													TimeUntilDisconnect: rand.Intn(10), 
													ElevatorInfo: elevator.T_ElevatorInfo{
														Floor: int8(rand.Intn(5)), 
														Direction: elevator.T_ElevatorDirection(rand.Intn(2)), 
														State: elevator.T_ElevatorState(uint8(rand.Intn(3))),
													}}
			slaveMessage1.Transmitter = T_NodeInfo{	PRIORITY: uint8(rand.Intn(10)), 
													Role: T_NodeRole(uint8(rand.Intn(10))), 
													TimeUntilDisconnect: rand.Intn(10), 
													ElevatorInfo: elevator.T_ElevatorInfo{
														Floor: int8(rand.Intn(5)), 
														Direction: elevator.T_ElevatorDirection(rand.Intn(2)), 
														State: elevator.T_ElevatorState(uint8(rand.Intn(3))),
													}}

			entry := T_GlobalQueueEntry{
				Request: elevator.T_Request{
					Id:        uint16(i),
					State:     elevator.UNASSIGNED,
					Calltype:  elevator.T_Call(rand.Intn(10)),
					Floor:     int8(rand.Intn(4)),
					Direction: elevator.T_ElevatorDirection(rand.Intn(2)),
				},
				RequestedNode:     uint8(rand.Intn(3)),
				AssignedNode:      uint8(rand.Intn(3)),
				TimeUntilReassign: uint8(rand.Intn(10)),
			}

			globalQ1 = append(globalQ1, entry)
			masterMessage1.GlobalQueue = globalQ1
			slaveMessage1.Entry = entry

			c_transmitMasterMessage <- masterMessage1
			sentToMasterTransmit = append(sentToMasterTransmit, masterMessage1)

			c_transmitSlaveMessage <- slaveMessage1
			sentToSlaveTransmit = append(sentToSlaveTransmit, slaveMessage1)
			
			time.Sleep(100 * time.Millisecond)

		}
	}()

	// go func() {

	// 	for i := 1; i < 5; i++ {
	// 		var masterMessage2 T_MasterMessage
	// 		var slaveMessage2 T_SlaveMessage
	// 		masterMessage2.Transmitter = T_NodeInfo{}
	// 		slaveMessage2.Transmitter = T_NodeInfo{}

	// 		entry := T_GlobalQueueEntry{
	// 			Request: elevator.T_Request{
	// 				Id:        uint16(i),
	// 				State:     elevator.UNASSIGNED,
	// 				Calltype:  6,
	// 				Floor:     5,
	// 				Direction: 4,
	// 			},
	// 			RequestedNode:     3,
	// 			AssignedNode:      2,
	// 			TimeUntilReassign: 1,
	// 		}

	// 		globalQ2 = append(globalQ2, entry)
	// 		masterMessage2.GlobalQueue = globalQ2
	// 		slaveMessage2.Entry = entry
	// 		c_transmitMasterMessage <- masterMessage2
	// 		c_transmitSlaveMessage <- slaveMessage2
	// 		time.Sleep(100 * time.Millisecond)

	// 	}
	// }()
	
	go func () {
		for {
			select{
			case verified := <- c_verifiedMasterMessage:
				fmt.Printf("\nMASTER VERIFIED Message In: %d | Message Out: %d\n\n", sentToMasterTransmit[0], verified)
				copy(sentToMasterTransmit, sentToMasterTransmit[1:])
				sentToMasterTransmit = sentToMasterTransmit[:len(sentToMasterTransmit)-1]
			
			case verified := <- c_verifiedSlaveMessage:
				fmt.Printf("\nSLAVE VERIFIED Message In %d | Message Out %d\n\n", sentToSlaveTransmit[0], verified)
				copy(sentToSlaveTransmit, sentToSlaveTransmit[1:])
				sentToSlaveTransmit = sentToSlaveTransmit[:len(sentToSlaveTransmit)-1]

			case <-quitTimer.C:
				fmt.Printf("TEST:\t\t Quitting\n")
				f_printOrderedMasterMaps()
				f_printOrderedSlaveMaps()
				c_quit <- true
				time.Sleep(1 * time.Second)
				return	
			}
		}
	}()

	<-c_quit
	time.Sleep(5 * time.Second)
	
}