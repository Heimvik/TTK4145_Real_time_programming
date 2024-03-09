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

type OrderedMap struct { //Used to keep track of the last 100 messages
	keys   []uint32
	values map[uint32]T_MasterMessage
}

func (om *OrderedMap) Add(key uint32, value T_MasterMessage) { //Adds a message to the map, deletes the oldest message if the map is full
	if _, exists := om.values[key]; !exists {
		om.keys = append(om.keys, key)
	}
	om.values[key] = value

	if len(om.keys) > 100 {
		delete(om.values, om.keys[0])
		om.keys = om.keys[1:]
	}
}

func (m *OrderedMap) Exists(key uint32) bool {
    _, exists := m.values[key]
    return exists
}

func (m *OrderedMap) Print() {
    if len(m.keys) == 0 {
        fmt.Println("The map is empty.", )
        return
    }

    for _, key := range m.keys {
        value := m.values[key]
        fmt.Printf("Key: %v, Value: %+v\n", key, value)
    }
}

var transmittedMasterMessages = OrderedMap{
	keys:   make([]uint32, 0),
	values: make(map[uint32]T_MasterMessage),
}

var receivedMasterMessages = OrderedMap{
	keys:   make([]uint32, 0),
	values: make(map[uint32]T_MasterMessage),
}

var verifiedMasterMessages = OrderedMap{
	keys:   make([]uint32, 0),
	values: make(map[uint32]T_MasterMessage),
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

func F_TransmitSlaveMessage(c_transmitSlaveMessage chan T_SlaveMessage, port int) {
	c_slaveMessageWithCS := make(chan T_SlaveMessage)
	go bcast.Transmitter(port, c_slaveMessageWithCS)

	for {
		select {
		case transmitSlaveMessage := <-c_transmitSlaveMessage:
			transmitSlaveMessage.Checksum = f_convertSlaveMessageToCS(transmitSlaveMessage)
			for i := 0; i < 50; i++ {
				c_slaveMessageWithCS <- transmitSlaveMessage
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
					fmt.Printf("%d messages with equal checksum found: Verified\n", len(equalMMs))
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

func F_ReceiveMasterMessage(c_verifiedMessage chan T_MasterMessage, port int) {
	c_receive := make(chan T_MasterMessage)

	go bcast.Receiver(port, c_receive)
	currentCSmap := make(map[uint32][]T_MasterMessage)

	for {
		select {
		case receivedMessage := <-c_receive:
			if len(currentCSmap[receivedMessage.Checksum]) < 50 {
				currentCSmap[receivedMessage.Checksum] = append(currentCSmap[receivedMessage.Checksum], receivedMessage)
				//logCurrentCSmap(currentCSmap)
			}
		default:
			var verifiedCS uint32
			shouldResetMap := false

			for currentCS, equalMMs := range currentCSmap {
				if len(equalMMs) >= 10 && len(equalMMs) < 50 && f_MasterMessagesAreEqual(equalMMs[0], equalMMs) {
					fmt.Printf("%d messages with equal checksum found: Verified\n", len(equalMMs))
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
				logCurrentCSmap(currentCSmap)
				newCSmap := make(map[uint32][]T_MasterMessage)
				newCSmap[verifiedCS] = currentCSmap[verifiedCS]
				currentCSmap = newCSmap
			}
		}
	}
}

//Kommentarer til f_MasterMessagesAreEqual:
//Sammenligner kun første indexen i arrayet med resten av arrayet
//Dvs. hvis første index er feil, returnerer den alltid false
//Burde heller lage en funksjon som finner de meldingene som er like, og returnerer true hvis det er flere enn 10 og evt. indexnummeret.
//Sjekker kun GlobalQueue, ikke Transmitter

//Kommentarer til F_ReceiveMasterMessage:
//Hva skjedde med å calculate checksum og sammenligne med vedlagt checksum?


func f_ArboAmountOfEqualMessages(messageList []T_MasterMessage) (int, int) { //Returns the amount of equal messages and the index of the message with the highest amount of equal messages

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
func F_TransmitMasterMessage(c_transmitMasterMessage chan T_MasterMessage, port int) {
	c_masterMessageWithCS := make(chan T_MasterMessage)
	go bcast.Transmitter(port, c_masterMessageWithCS)
	for {
		select {
		case transmitMasterMessage := <-c_transmitMasterMessage:
			transmitMasterMessage.Checksum = f_convertMasterMessageToCS(transmitMasterMessage)
			for i := 0; i < 10; i++ {
				c_masterMessageWithCS <- transmitMasterMessage
				transmittedMasterMessages.Add(transmitMasterMessage.Checksum, transmitMasterMessage)
				time.Sleep(time.Duration(1) * time.Millisecond)
			}
		}
	}
}

func F_ArboReceiveMasterMessage(c_verifiedMessage chan T_MasterMessage, port int) {
	c_receive := make(chan T_MasterMessage)
	go bcast.Receiver(port, c_receive)
	currentCSmap := make(map[uint32][]T_MasterMessage)

	messageAlreadySentmap := OrderedMap{
		keys:   make([]uint32, 0),
		values: make(map[uint32]T_MasterMessage),
	}

	ticker := time.NewTicker(time.Millisecond * 100) // check for new messages every 500 milliseconds
	defer ticker.Stop()

	for {
		select {
		case receivedMessage := <-c_receive:
			fmt.Printf("Received message: %v\n", receivedMessage.Checksum)
			if !messageAlreadySentmap.Exists(receivedMessage.Checksum) {
				currentCSmap[receivedMessage.Checksum] = append(currentCSmap[receivedMessage.Checksum], receivedMessage)
				receivedMasterMessages.Add(receivedMessage.Checksum, receivedMessage)
			}
		case <-ticker.C: // check for new messages every 500 milliseconds
			for checksum, messages := range currentCSmap {
				counter, index := f_ArboAmountOfEqualMessages(messages)
				if counter >= 10 {
					fmt.Printf("%d identical messages found: Verified\n", counter)
					verifiedMasterMessages.Add(checksum, messages[index])

					c_verifiedMessage <- messages[index]
					messageAlreadySentmap.Add(checksum, messages[index])
					delete(currentCSmap, checksum)
	
				}
			}
		}
	}
}

func F_ArboTestCommunication() {
	c_receiveMasterMessage := make(chan T_MasterMessage)
	c_transmitMasterMessage := make(chan T_MasterMessage)

	go F_ArboReceiveMasterMessage(c_receiveMasterMessage, MASTERPORT)
	go F_TransmitMasterMessage(c_transmitMasterMessage, MASTERPORT)

	quitTimer := time.NewTicker(time.Duration(15) * time.Second)

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
			time.Sleep(time.Duration(1) * time.Second)

		}
	}()

	for {
		select {
		case <-quitTimer.C:
			f_printOrderedMaps()
			return
		}
	}

}

func logCurrentCSmap(currentCSmap map[uint32][]T_MasterMessage) {
	for cs, messages := range currentCSmap {
		fmt.Printf("Checksum: %d, Messages Count: %d\n", cs, len(messages))
		for i, message := range messages {
			fmt.Printf("\tMessage %d: %s\n", i+1, message.String())
		}
	}
}

func printChecksumT_MasterMessageMap(checksumT_MasterMessageMap map[uint32]T_MasterMessage) {
	for checksum, message := range checksumT_MasterMessageMap {
		fmt.Printf("Checksum: %d, Message: %s\n", checksum, message.String())
	}
}

func f_printOrderedMaps() {
	fmt.Println("Transmitted Master Messages:")
	transmittedMasterMessages.Print()

	fmt.Println("\nReceived Master Messages:")
	receivedMasterMessages.Print()

	fmt.Println("\nVerified Master Messages:")
	verifiedMasterMessages.Print()
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
			fmt.Println("Message sent: %v", masterMessage)

			time.Sleep(time.Duration(5) * time.Second)

		}
	}()

	for {
		select {
		case masterMessage := <-c_receiveMasterMessage:
			fmt.Println("Message received: %v", masterMessage)

		}
	}
}


