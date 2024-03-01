package node

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"the-elevator/elevator"
	"time"
)

/*
func f_TestDistribution() {
	var elevators []*elevator.T_Elevator
	for i := 0; i <= 2; i++ {
		newElevator := elevator.T_Elevator{
			Floor:     4 - i,
			MotorDirection: elevator.Down,
			State: elevator.EB_Idle,
		}
		if i == 2 {
			newElevator.Avalibale = false
		}
		elevators = append(elevators, &newElevator)
	}

	request := T_Request{
		Calltype:   Hall,
		P_Elevator: elevators[0],
		Floor:      1, //elevators[0].Floor,
		Direction:  Up,
	}

	fmt.Println(F_AssignRequest(&request, elevators).Floor)
}
*/
/*
func f_TestCommunication(port int) {
	// Setup
	c_transmitMessage := make(chan T_Message)
	c_receivedMessage := make(chan T_Message)
	c_connectedNodes := make(chan []*T_NodeInfo)

	go F_TransmitMessages(c_transmitMessage, port)
	go F_ReceiveMessages(c_receivedMessage, thisNode.ConnectedNodes, c_connectedNodes, port)

	go func() {
		i := 0
		helloMsg := T_Message{*thisNode.Info, " says " + strconv.Itoa(i)}
		for {
			helloMsg.TestStr = " says " + strconv.Itoa(i)
			c_transmitMessage <- helloMsg
			time.Sleep(1 * time.Second)
			i++
		}
	}()

	for {
		select {
		case received := <-c_receivedMessage:
			// Log the received message immediately when it's available
			F_WriteLog("Received message:" + received.TestStr)
		case connectedNodes := <-c_connectedNodes:
			// Log the updated list of connected nodes immediately when it's available
			F_WriteLog("Connected nodes updated:")
			for _, node := range connectedNodes {
				F_WriteLog("Node: " + strconv.Itoa(node.PRIORITY) + " as " + strconv.Itoa(int(node.Role)))
			}
		}
	}
}
*/
// ***	END TEST FUNCTIONS	***//
// nano .gitconfig

func f_InitNode(config T_Config) T_Node {
	thisNodeInfo := T_NodeInfo{
		PRIORITY: config.Priority,
		Role:     MASTER,
	}
	p_thisElevatorInfo := &elevator.T_ElevatorInfo{
		Direction: elevator.NONE,
		State:     elevator.IDLE,
	}
	var c_receiveRequest chan elevator.T_Request
	var c_distributeRequest chan elevator.T_Request
	var c_distributeInfo chan elevator.T_ElevatorInfo

	p_thisElevator := &elevator.T_Elevator{
		P_info:              p_thisElevatorInfo,
		C_receiveRequest:    c_receiveRequest,
		C_distributeRequest: c_distributeRequest,
		C_distributeInfo:    c_distributeInfo,
	}
	thisNode := T_Node{
		Info:       thisNodeInfo,
		P_ELEVATOR: p_thisElevator,
	}
	return thisNode
}

func init() {
	logFile, _ := os.OpenFile("log/debug1.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	defer logFile.Close()

	configFile, _ := os.Open("config/default.json")
	defer configFile.Close()

	var config T_Config
	json.NewDecoder(configFile).Decode(&config)

	// Init thisNode
	ThisNode = f_InitNode(config)

	IP = config.Ip
	FLOORS = config.Floors
	REASSIGNTIME = config.ReassignTime
	CONNECTIONTIME = config.ConnectionTime
	SENDPERIOD = config.SendPeriod
	GETSETPERIOD = config.GetSetPeriod
	SLAVEPORT = config.SlavePort
	MASTERPORT = config.MasterPort
	ELEVATORPORT = config.ElevatorPort
}

func F_WriteLog(text string) bool {
	logFile, _ := os.OpenFile("log/debug1.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Print(text)
	return true
}
func f_WriteLogConnectedNodes(ops T_NodeOperations, connectedNodes []T_NodeInfo) {
	thisNode := f_GetNodeInfo(ops)
	logStr := "Node: | " + strconv.Itoa(thisNode.PRIORITY) + " | "
	if thisNode.Role == MASTER {
		logStr += "MASTER"
	} else {
		logStr += "SLAVE"
	}
	logStr += "| has connected nodes | "
	for _, nodeInfo := range f_GetConnectedNodes(ops) {
		logStr += strconv.Itoa(nodeInfo.PRIORITY) + " | "
	}
	F_WriteLog(logStr)
}
func F_WriteLogGlobalQueueEntry(ops T_NodeOperations, entry T_GlobalQueueEntry) {
	logStr := "Requested: | " + strconv.Itoa(entry.Request.Id) + " | " + strconv.Itoa(entry.RequestedNode.PRIORITY) + " | "
	if entry.RequestedNode.Role == MASTER {
		logStr += "MASTER |\t"
	} else {
		logStr += "SLAVE |\t"
	}
	logStr += "\tAssigned: " + strconv.Itoa(entry.Request.Id) + " | " + strconv.Itoa(entry.RequestedNode.PRIORITY) + " | "
	if entry.RequestedNode.Role == MASTER {
		logStr += "MASTER | "
	} else {
		logStr += "SLAVE | "
	}
	logStr += "Request: | "
	if entry.Request.Calltype == 0 {
		logStr += "CAB | "
	} else {
		logStr += "HALL | "
	}
	logStr += "\t" + strconv.Itoa(entry.Request.Floor) + " | "
	if entry.Request.Direction == 0 {
		logStr += "UP | "
	} else if entry.Request.Direction == 1 {
		logStr += "DOWN | "
	} else {
		logStr += "NONE | "
	}
	logStr += "\t Reassigned in: " + strconv.FormatFloat(float64(entry.TimeUntilReassign), 'f', 2, 32)
	F_WriteLog(logStr)
}
func f_WriteLogSlaveMessage(ops T_NodeOperations, slaveMessage T_SlaveMessage) {
	request := slaveMessage.Entry.Request
	thisNode := f_GetNodeInfo(ops)
	logStr := "Node: | " + strconv.Itoa(thisNode.PRIORITY) + " | "
	if slaveMessage.Transmitter.Role == MASTER {
		logStr += "MASTER"
	} else {
		logStr += "SLAVE"
	}
	logStr += " | received SM from | " + strconv.Itoa(slaveMessage.Transmitter.PRIORITY) + " | "
	if request.Calltype == 0 {
		logStr += "NONECALL | "
	} else if request.Calltype == 1 {
		logStr += "CAB | "
	} else if request.Calltype == 2 {
		logStr += "HALL | "
	}
	logStr += strconv.Itoa(request.Floor) + " | "
	if request.Direction == 1 {
		logStr += "UP | "
	} else if request.Direction == -1 {
		logStr += "DOWN | "
	} else if request.Direction == 0 {
		logStr += "NONE | "
	}
	F_WriteLog(logStr)
}
func f_WriteLogMasterMessage(ops T_NodeOperations, masterMessage T_MasterMessage) {
	thisNode := f_GetNodeInfo(ops)
	logStr := "Node: | " + strconv.Itoa(thisNode.PRIORITY) + " | "
	if masterMessage.Transmitter.Role == MASTER {
		logStr += "MASTER"
	} else {
		logStr += "SLAVE"
	}
	logStr += " | received MM from | " + strconv.Itoa(masterMessage.Transmitter.PRIORITY) + " | "
	if masterMessage.Transmitter.Role == MASTER {
		logStr += "MASTER | "
	} else {
		logStr += "SLAVE | "
	}
	//add more about GQ
	F_WriteLog(logStr)
}
func f_AssignNewRole(thisNodeInfo T_NodeInfo, connectedNodes []T_NodeInfo) T_NodeInfo {
	var returnRole T_NodeRole
	for _, remoteNodeInfo := range connectedNodes {
		if remoteNodeInfo.PRIORITY < thisNodeInfo.PRIORITY {
			returnRole = SLAVE
		} else {
			returnRole = MASTER
		}
	}
	newNodeInfo := T_NodeInfo{
		PRIORITY:            thisNodeInfo.PRIORITY,
		Role:                returnRole,
		ElevatorInfo:        thisNodeInfo.ElevatorInfo,
		TimeUntilDisconnect: thisNodeInfo.TimeUntilDisconnect,
	}
	return newNodeInfo
}
func f_RemoveNode(nodes []T_NodeInfo, nodeToRemove T_NodeInfo) []T_NodeInfo {
	for i, nodeInfo := range nodes {
		if nodeInfo.PRIORITY == nodeToRemove.PRIORITY {
			return append(nodes[:i], nodes[i+1:]...)
		}
	}
	return nodes
}
func f_GetAvalibaleNodes(ops T_NodeOperations) []T_NodeInfo {
	var avalibaleNodes []T_NodeInfo
	connectedNodes := f_GetConnectedNodes(ops)
	for _, nodeInfo := range connectedNodes {
		if nodeInfo.ElevatorInfo.State == elevator.IDLE {
			avalibaleNodes = append(avalibaleNodes, nodeInfo)
		}
	}
	return avalibaleNodes
}

// decrements all timers in GQ and checks for any that has run out, will always try to reassign/remove GQ elements/connectednodes
func f_TimeManager(ops T_NodeOperations, c_send chan bool) {
	for {
		fmt.Println("Here")
		c_readGlobalQueue := make(chan []T_GlobalQueueEntry)
		c_writeGlobalQueue := make(chan []T_GlobalQueueEntry)
		c_quit := make(chan bool)
		go f_GetAndSetGlobalQueue(ops, c_readGlobalQueue, c_writeGlobalQueue, c_quit)
		globalQueue := <-c_readGlobalQueue

		entryToReassign := T_GlobalQueueEntry{}
		for _, entry := range globalQueue {
			if entry.TimeUntilReassign > 0 {
				entry.TimeUntilReassign -= 1
			}
			if entry.TimeUntilReassign == 0 && entry.Request.State != elevator.DONE {
				entryToReassign = entry
				break
			}
		}
		if (entryToReassign != T_GlobalQueueEntry{}) {
			globalQueue = append(globalQueue, entryToReassign)
			c_writeGlobalQueue <- globalQueue
		} else {
			c_writeGlobalQueue <- globalQueue
		}

		c_readConnectedNodes := make(chan []T_NodeInfo)
		c_writeConnectedNodes := make(chan []T_NodeInfo)
		go f_GetAndSetConnectedNodes(ops, c_readConnectedNodes, c_writeConnectedNodes, c_quit)
		oldConnectedNodes := <-c_readConnectedNodes

		nodeToDisconnect := T_NodeInfo{}
		for _, nodeInfo := range oldConnectedNodes {
			if nodeInfo.TimeUntilDisconnect > 0 {
				nodeInfo.TimeUntilDisconnect -= 1
			}
			fmt.Println(strconv.Itoa(nodeInfo.PRIORITY) + " | " + strconv.FormatFloat(float64(nodeInfo.TimeUntilDisconnect), 'f', -1, 32))
			if nodeInfo.TimeUntilDisconnect == 0 {
				nodeToDisconnect = nodeInfo
				break
			}
		}
		if (nodeToDisconnect != T_NodeInfo{}) {
			newConnectedNodes := f_RemoveNode(oldConnectedNodes, nodeToDisconnect)
			F_WriteLog("Node " + strconv.Itoa(nodeToDisconnect.PRIORITY) + " disconnected")
			c_writeConnectedNodes <- newConnectedNodes
		} else {
			c_writeConnectedNodes <- oldConnectedNodes
		}
		close(c_quit)

		time.Sleep(1 * time.Second)
	}
}
func f_UpdateConnectedNodes(ops T_NodeOperations, currentNode T_NodeInfo) {
	c_readConnectedNodes := make(chan []T_NodeInfo)
	c_writeConnectedNodes := make(chan []T_NodeInfo)
	c_quit := make(chan bool)
	go f_GetAndSetConnectedNodes(ops, c_readConnectedNodes, c_writeConnectedNodes, c_quit)
	oldConnectedNodes := <-c_readConnectedNodes

	nodeIsUnique := true
	nodeIndex := 0
	for i, oldConnectedNode := range oldConnectedNodes {
		if currentNode.PRIORITY == oldConnectedNode.PRIORITY {
			nodeIsUnique = false
			nodeIndex = i
			break
		}
	}

	if nodeIsUnique {
		currentNode.TimeUntilDisconnect = CONNECTIONTIME
		connectedNodes := append(oldConnectedNodes, currentNode)
		c_writeConnectedNodes <- connectedNodes
	} else {
		currentNode.TimeUntilDisconnect = CONNECTIONTIME
		oldConnectedNodes[nodeIndex] = currentNode
		c_writeConnectedNodes <- oldConnectedNodes
	}
	c_quit <- true
}
func f_AddEntryGlobalQueue(ops T_NodeOperations, entryToAdd T_GlobalQueueEntry) {
	c_readGlobalQueue := make(chan []T_GlobalQueueEntry)
	c_writeGlobalQueue := make(chan []T_GlobalQueueEntry)
	c_quit := make(chan bool)
	go f_GetAndSetGlobalQueue(ops, c_readGlobalQueue, c_writeGlobalQueue, c_quit)
	oldGlobalQueue := <-c_readGlobalQueue

	entryIsUnique := true
	entryIndex := 0
	for i, entry := range oldGlobalQueue {
		if entryToAdd.Request.Id == entry.Request.Id && entryToAdd.RequestedNode == entry.RequestedNode { //random id generated to each entry
			entryIsUnique = false
			entryIndex = i
			break
		}
	}
	if entryIsUnique {
		entryToAdd.AssignedNode.PRIORITY = 0
		//entryToAdd.TimeUntilReassign = REASSIGNTIME //Should be set by slave
		oldGlobalQueue = append(oldGlobalQueue, entryToAdd)
		c_writeGlobalQueue <- oldGlobalQueue
	} else { //should update the existing entry
		oldGlobalQueue[entryIndex] = entryToAdd
		c_writeGlobalQueue <- oldGlobalQueue
	}
	c_quit <- true
}
func f_ElevatorManager(ops T_NodeOperations, c_entryFromElevator chan T_GlobalQueueEntry) {

	c_requestToElevator := make(chan elevator.T_Request)
	c_requestFromElevator := make(chan elevator.T_Request)
	//go elevator.F_RunElevator(c_requestToElevator, c_requestFromElevator)

	for {
		globalQueue := f_GetGlobalQueue(ops)
		thisNodeInfo := f_GetNodeInfo(ops)
		requestToElevator := F_FindAssignedRequest(globalQueue, thisNodeInfo) //request for this node to take
		if requestToElevator != (elevator.T_Request{}) {
			c_requestToElevator <- requestToElevator
		}
		select {
		case receivedRequest := <-c_requestFromElevator:
			//make a GlobalQueueEentry, and add to globalQueue
			entry := T_GlobalQueueEntry{
				Request:           receivedRequest,
				RequestedNode:     f_GetNodeInfo(ops),
				AssignedNode:      T_NodeInfo{},
				TimeUntilReassign: REASSIGNTIME,
			}
			c_entryFromElevator <- entry
		default:
			continue
		}
	}
}

//IMPORTANT:
//-global variables should ALWAYS be handled by server to operate onn good data
//-all receive from channles should be organized in for-select!!! -> walk trough code and do

// should contain the main master/slave fsm in Run() function, to be called from main
func F_RunNode() {
	C_nodeOpMsg = T_NodeOperations{ //Make global for jonas
		c_readNodeInfo:         make(chan chan T_NodeInfo),
		c_writeNodeInfo:        make(chan T_NodeInfo),
		c_readAndWriteNodeInfo: make(chan chan T_NodeInfo),

		c_readGlobalQueue:         make(chan chan []T_GlobalQueueEntry),
		c_writeGlobalQueue:        make(chan []T_GlobalQueueEntry),
		c_readAndWriteGlobalQueue: make(chan chan []T_GlobalQueueEntry),

		c_readConnectedNodes:         make(chan chan []T_NodeInfo),
		c_writeConnectedNodes:        make(chan []T_NodeInfo),
		c_readAndWriteConnectedNodes: make(chan chan []T_NodeInfo),

		c_readElevator:         make(chan chan elevator.T_Elevator),
		c_writeElevator:        make(chan elevator.T_Elevator),
		c_readAndWriteElevator: make(chan chan elevator.T_Elevator),
	}

	//to run the main FSM
	c_receiveSlaveMessage := make(chan T_SlaveMessage)
	c_receiveMasterMessage := make(chan T_MasterMessage)
	c_transmitMasterMessage := make(chan T_MasterMessage)
	c_transmitSlaveMessage := make(chan T_SlaveMessage)

	c_send := make(chan bool)
	c_entryFromElevator := make(chan T_GlobalQueueEntry)

	go f_TimeManager(c_nodeOpMsg, c_send)
	go f_NodeOperationManager(&ThisNode, c_nodeOpMsg) //SHOULD BE THE ONLY REFERENCE TO ThisNode!
	go f_ElevatorManager(c_nodeOpMsg, c_entryFromElevator)
	go F_ReceiveSlaveMessage(c_receiveSlaveMessage, c_nodeOpMsg, SLAVEPORT)
	go F_ReceiveMasterMessage(c_receiveMasterMessage, c_nodeOpMsg, MASTERPORT)
	go F_TransmitSlaveMessage(c_transmitSlaveMessage, SLAVEPORT)
	go F_TransmitMasterMessage(c_transmitMasterMessage, MASTERPORT)
	sendTimer := time.NewTicker(time.Duration(SENDPERIOD) * time.Millisecond)
	for {
		nodeRole := f_GetNodeInfo(c_nodeOpMsg).Role
		switch nodeRole {
		case MASTER:
			c_quitMasterRoutines := make(chan bool)
			select {
			case masterMessage := <-c_receiveMasterMessage:
				f_WriteLogMasterMessage(c_nodeOpMsg, masterMessage)
				f_UpdateConnectedNodes(c_nodeOpMsg, masterMessage.Transmitter)
				f_WriteLogConnectedNodes(c_nodeOpMsg, f_GetConnectedNodes(c_nodeOpMsg))
				for _, remoteEntry := range masterMessage.GlobalQueue {
					f_AddEntryGlobalQueue(c_nodeOpMsg, remoteEntry)
				}
				//IMPORTANT: cannot really propagate to slave until it knows that the other master has received its GQ

			case slaveMessage := <-c_receiveSlaveMessage:
				f_WriteLogSlaveMessage(c_nodeOpMsg, slaveMessage)
				f_UpdateConnectedNodes(c_nodeOpMsg, slaveMessage.Transmitter)
				f_WriteLogConnectedNodes(c_nodeOpMsg, f_GetConnectedNodes(c_nodeOpMsg))
				f_AddEntryGlobalQueue(c_nodeOpMsg, slaveMessage.Entry)

			case entryFromElevator := <-c_entryFromElevator:
				f_AddEntryGlobalQueue(c_nodeOpMsg, entryFromElevator)

				thisNode := f_GetNodeInfo(c_nodeOpMsg)
				F_WriteLog("Node: " + strconv.Itoa(thisNode.PRIORITY) + " as MASTER added GQ entry:\n")
				F_WriteLogGlobalQueueEntry(c_nodeOpMsg, entryFromElevator)
			case <-sendTimer.C:
				masterMessage := T_MasterMessage{
					Transmitter: f_GetNodeInfo(c_nodeOpMsg),
					GlobalQueue: f_GetGlobalQueue(c_nodeOpMsg),
				}
				c_transmitMasterMessage <- masterMessage
				//F_WriteLog("MasterMessage sent on port: " + strconv.Itoa(MASTERPORT))
				sendTimer.Reset(time.Duration(SENDPERIOD) * time.Millisecond)
			default:
				c_readNodeInfo := make(chan T_NodeInfo)
				c_writeNodeInfo := make(chan T_NodeInfo)
				c_quitGetSet := make(chan bool)

				go f_GetAndSetNodeInfo(c_nodeOpMsg, c_readNodeInfo, c_writeNodeInfo, c_quitGetSet)
				oldConnectedNodes := f_GetConnectedNodes(c_nodeOpMsg)

				oldNodeInfo := <-c_readNodeInfo
				newNodeInfo := f_AssignNewRole(oldNodeInfo, oldConnectedNodes)
				c_writeNodeInfo <- newNodeInfo
				c_quitGetSet <- true

				f_UpdateConnectedNodes(c_nodeOpMsg, newNodeInfo) //Update connected nodes with newnodeinfo

				c_readGlobalQueue := make(chan []T_GlobalQueueEntry)
				c_writeGlobalQueue := make(chan []T_GlobalQueueEntry)
				c_quit := make(chan bool)
				go f_GetAndSetGlobalQueue(c_nodeOpMsg, c_readGlobalQueue, c_writeGlobalQueue, c_quit)

				globalQueue := <-c_readGlobalQueue
				avalibaleNodes := f_GetAvalibaleNodes(c_nodeOpMsg)
				for i, entry := range globalQueue {
					if (entry.Request.State == elevator.UNASSIGNED || entry.AssignedNode.PRIORITY == 0) && len(avalibaleNodes) > 0 { //OR for redundnacy, both should not be different in theory
						assignedEntry := F_AssignEntry(entry, avalibaleNodes)
						globalQueue[i] = assignedEntry
						break
					}
				}
				c_writeGlobalQueue <- globalQueue

				if newNodeInfo.Role == SLAVE {
					close(c_quitMasterRoutines)
				}
			}

		case SLAVE:
			c_quitSlaveRoutines := make(chan bool)
			select {
			case masterMessage := <-c_receiveMasterMessage:
				f_WriteLogMasterMessage(c_nodeOpMsg, masterMessage)
				for _, remoteEntry := range masterMessage.GlobalQueue {
					f_AddEntryGlobalQueue(c_nodeOpMsg, remoteEntry)
				}
				f_UpdateConnectedNodes(c_nodeOpMsg, masterMessage.Transmitter)
			case slaveMessage := <-c_receiveSlaveMessage:
				f_WriteLogSlaveMessage(c_nodeOpMsg, slaveMessage)
				f_UpdateConnectedNodes(c_nodeOpMsg, slaveMessage.Transmitter)

			case entryFromElevator := <-c_entryFromElevator:
				infoMessage := T_SlaveMessage{
					Transmitter: f_GetNodeInfo(c_nodeOpMsg),
					Entry:       entryFromElevator,
				}
				c_transmitSlaveMessage <- infoMessage
			case <-sendTimer.C:
				aliveMessage := T_SlaveMessage{
					Transmitter: f_GetNodeInfo(c_nodeOpMsg),
					Entry:       T_GlobalQueueEntry{},
				}
				c_transmitSlaveMessage <- aliveMessage
				sendTimer.Reset(time.Duration(SENDPERIOD) * time.Millisecond)
			default:
				connectedNodes := f_GetConnectedNodes(c_nodeOpMsg)
				thisNodeInfo := f_GetNodeInfo(c_nodeOpMsg)
				thisNodeInfo.Role = f_AssignNewRole(thisNodeInfo, connectedNodes)

				f_SetNodeInfo(c_nodeOpMsg, thisNodeInfo)

				f_UpdateConnectedNodes(c_nodeOpMsg, f_GetNodeInfo(c_nodeOpMsg))

				if thisNodeInfo.Role == MASTER {
					close(c_quitSlaveRoutines)
				}
			}
		}
	}
}
