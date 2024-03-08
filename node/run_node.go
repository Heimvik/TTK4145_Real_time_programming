package node

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"the-elevator/node/elevator"
	"time"
)

/*
Kjappe kommentarer fra sia:
- Flytte testfunksjoner til eget dokument
- Flytte alt som har med å skrive til log til et eget dokument
- Kanskje flytte alt som har med globalQueue til eget dokument
*/

// ***	END TEST FUNCTIONS	***//

func f_InitNode(config T_Config) T_Node {
	thisElevatorInfo := elevator.T_ElevatorInfo{
		Direction: elevator.NONE,
		Floor:     1, //-1, 1 for test purposes only!
		State:     elevator.IDLE,
	}
	thisNodeInfo := T_NodeInfo{
		PRIORITY:     config.Priority,
		ElevatorInfo: thisElevatorInfo,
		Role:         MASTER,
	}

	thisElevator := elevator.T_Elevator{
		P_info:         &thisElevatorInfo,
		P_serveRequest: nil,
		CurrentID:      0,
		Obstructed:     false,
		StopButton:     false,
	}
	thisNode := T_Node{
		NodeInfo: thisNodeInfo,
		Elevator: thisElevator,
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

	nodeOperations = T_NodeOperations{
		c_getNodeInfo:    make(chan chan T_NodeInfo),
		c_setNodeInfo:    make(chan T_NodeInfo),
		c_getSetNodeInfo: make(chan chan T_NodeInfo),

		c_getGlobalQueue:    make(chan chan []T_GlobalQueueEntry),
		c_setGlobalQueue:    make(chan []T_GlobalQueueEntry),
		c_getSetGlobalQueue: make(chan chan []T_GlobalQueueEntry),

		c_getConnectedNodes:    make(chan chan []T_NodeInfo),
		c_setConnectedNodes:    make(chan []T_NodeInfo),
		c_getSetConnectedNodes: make(chan chan []T_NodeInfo),
	}
	elevatorOperations = elevator.T_ElevatorOperations{
		C_getElevator:    make(chan chan elevator.T_Elevator),
		C_setElevator:    make(chan elevator.T_Elevator),
		C_getSetElevator: make(chan chan elevator.T_Elevator),
	}

	IP = config.Ip
	FLOORS = config.Floors
	REASSIGNTIME = config.ReassignTime
	CONNECTIONTIME = config.ConnectionTime
	SENDPERIOD = config.SendPeriod
	GETSETPERIOD = config.GetSetPeriod
	SLAVEPORT = config.SlavePort
	MASTERPORT = config.MasterPort
	ELEVATORPORT = config.ElevatorPort
	ASSIGNBREAKOUTPERIOD = config.AssignBreakoutPeriod
	MOSTRESPONSIVEPERIOD = config.MostResponsivePeriod
	MEDIUMRESPONSIVEPERIOD = config.MiddleResponsivePeriod
	LEASTRESPONSIVEPERIOD = config.LeastResponsivePeriod
}

func F_WriteLog(text string) bool {
	logFile, _ := os.OpenFile("log/debug1.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Print(text)
	return true
}
func f_NodeRoleToString(role T_NodeRole) string {
	switch role {
	case MASTER:
		return "MASTER"
	default:
		return "SLAVE"
	}
}
func f_WriteLogConnectedNodes(connectedNodes []T_NodeInfo) {
	thisNode := f_GetNodeInfo()
	logStr := fmt.Sprintf("Node: | %d | %s | has connected nodes | ", thisNode.PRIORITY, f_NodeRoleToString(thisNode.Role))
	for _, info := range connectedNodes {
		logStr += fmt.Sprintf("%d (Role: %s, ElevatorInfo: %+v, TimeUntilDisconnect: %d) | ",
			info.PRIORITY, f_NodeRoleToString(info.Role), info.ElevatorInfo, info.TimeUntilDisconnect)
	}
	F_WriteLog(logStr)
}
func F_WriteLogGlobalQueueEntry(entry T_GlobalQueueEntry) {
	logStr := fmt.Sprintf("Entry: | %d | State: | %s | Calltype: %s | Floor: %d | Direction: %s | Reassigned in: %.2f | ",
		entry.Request.Id, f_RequestStateToString(entry.Request.State), f_CallTypeToString(entry.Request.Calltype), entry.Request.Floor, f_DirectionToString(entry.Request.Direction), float64(entry.TimeUntilReassign))
	logStr += fmt.Sprintf("Requested node: | %d | ",
		entry.RequestedNode)
	logStr += fmt.Sprintf("Assigned node: | %d | ",
		entry.AssignedNode)
	F_WriteLog(logStr)
}
func f_CallTypeToString(callType elevator.T_Call) string {
	switch callType {
	case 0:
		return "NONE"
	case 1:
		return "CAB"
	case 2:
		return "HALL"
	default:
		return "UNKNOWN"
	}
}
func f_RequestStateToString(state elevator.T_RequestState) string {
	switch state {
	case elevator.UNASSIGNED:
		return "UNASSIGNED"
	case elevator.ASSIGNED:
		return "ASSIGNED"
	case elevator.ACTIVE:
		return "ACTIVE"
	case elevator.DONE:
		return "DONE"
	default:
		return "UNKNOWN"
	}
}

func f_DirectionToString(direction elevator.T_ElevatorDirection) string {
	switch direction {
	case 1:
		return "UP"
	case -1:
		return "DOWN"
	case 0:
		return "NONE"
	default:
		return "UNKNOWN"
	}
}
func f_WriteLogSlaveMessage(slaveMessage T_SlaveMessage) {
	request := slaveMessage.Entry.Request
	thisNode := f_GetNodeInfo()
	logStr := fmt.Sprintf("Node: | %d | %s | received SM from | %d | Request ID: | %d | State: | %s | Calltype: %s | Floor: %d | Direction: %s |",
		int(thisNode.PRIORITY), f_NodeRoleToString(thisNode.Role), int(slaveMessage.Transmitter.PRIORITY), int(request.Id), f_RequestStateToString(slaveMessage.Entry.Request.State), f_CallTypeToString(request.Calltype), int(request.Floor), f_DirectionToString(request.Direction))
	F_WriteLog(logStr)
}
func f_WriteLogMasterMessage(masterMessage T_MasterMessage) {
	thisNode := f_GetNodeInfo()
	roleStr := f_NodeRoleToString(thisNode.Role)
	transmitterRoleStr := f_NodeRoleToString(masterMessage.Transmitter.Role)

	logStr := fmt.Sprintf("Node: | %d | %s | received MM from | %d | %s | GlobalQueue: [",
		thisNode.PRIORITY, roleStr, masterMessage.Transmitter.PRIORITY, transmitterRoleStr)

	for i, entry := range masterMessage.GlobalQueue {
		entryStr := fmt.Sprintf("Request ID: | %d | State: | %s | Calltype: %s | Floor: %d | Direction: %s | Reassigned in: %d | Requested node: | %d | Assigned node: | %d |",
			entry.Request.Id, f_RequestStateToString(entry.Request.State), f_CallTypeToString(entry.Request.Calltype), int(entry.Request.Floor),
			f_DirectionToString(entry.Request.Direction), int(entry.TimeUntilReassign),
			int(entry.RequestedNode), int(entry.AssignedNode))

		logStr += entryStr
		if i < len(masterMessage.GlobalQueue)-1 {
			logStr += ", "
		}
	}
	logStr += "]"

	F_WriteLog(logStr)
}

func f_AssignNewRole(thisNodeInfo T_NodeInfo, connectedNodes []T_NodeInfo) T_NodeInfo {
	var returnRole T_NodeRole = MASTER
	for _, remoteNodeInfo := range connectedNodes {
		if remoteNodeInfo.PRIORITY < thisNodeInfo.PRIORITY {
			returnRole = SLAVE
		}
	}
	newNodeInfo := T_NodeInfo{
		PRIORITY:            thisNodeInfo.PRIORITY,
		Role:                returnRole,
		TimeUntilDisconnect: thisNodeInfo.TimeUntilDisconnect,
		ElevatorInfo:        thisNodeInfo.ElevatorInfo,
	}
	return newNodeInfo
}

// JONASCOMMENT: bytte ut 'node' med et mer besrkivende navn (info kanskje)? har også forslag til å gjøre denne mer simpel (og kanskje mer oversiktlig, du kan vurdere det sjæl, det som er kommentert vekk er det som var fra før)
func f_FindNodeInfo(node uint8, connectedNodes []T_NodeInfo) T_NodeInfo {
	// returnNode := T_NodeInfo{}
	for _, nodeInfo := range connectedNodes {
		if node == nodeInfo.PRIORITY {
			return nodeInfo
			// returnNode = nodeInfo
			// break
		}
	}
	return T_NodeInfo{} // return returnNode
}
func f_GetAvalibaleNodes(connectedNodes []T_NodeInfo) []T_NodeInfo {
	var avalibaleNodes []T_NodeInfo
	for i, nodeInfo := range connectedNodes {
		if (nodeInfo != T_NodeInfo{} && nodeInfo.ElevatorInfo.State == elevator.IDLE) {
			avalibaleNodes = append(avalibaleNodes, connectedNodes[i])
		}
	}
	return avalibaleNodes
}

// JONASCOMMENT: syns 'id' er nice navn, er litt det jeg tenkte når jeg kommente i FindNodeInfo. Kan forenkles på nesten samme måte som FindNodeInfo, hvis du syns det funker
func f_FindEntry(id uint16, requestedNode uint8, globalQueue []T_GlobalQueueEntry) T_GlobalQueueEntry {
	returnEntry := T_GlobalQueueEntry{}
	for _, entry := range globalQueue {
		if id == entry.Request.Id && entry.RequestedNode == requestedNode {
			returnEntry = entry
		}
	}
	return returnEntry
}

// JONASCOMMENT: den her er for lang, har ikke noe umiddelbart fiks på hvordan den kan deles opp men en funksjon på 103 linjer er litt mye, har hørt snakk om at de ideelt skal ligge på 20-30 linjer
// JONASCOMMENT: legger til no halvveis tips om hvor du kan dele opp, ta med en klype salt
// JONASCOMMENT: kanskje også flytte ut alt som har med globalQueue til request_distrubutor.go?
// JONASCOMMENT: hva menes egentlig med variable watchdog? er det en watchdog for variabler?
func f_MasterVariableWatchDog(c_getSetGlobalQueueInterface chan T_GetSetGlobalQueueInterface, getSetGlobalQueueInterface T_GetSetGlobalQueueInterface, c_lastAssignedEntry chan T_GlobalQueueEntry, c_assignmentSuccessfull chan bool, c_ackSentEntryToSlave chan T_AckObject, c_quit chan bool) {
	go func() { //JONASCOMMENT: dette kan feks være en egen funksjon
		for {
		PollLastAssigned:
			select {
			case lastAssignedEntry := <-c_lastAssignedEntry:
				F_WriteLog("Getting ack from last assinged...")
				assignBreakoutTimer := time.NewTicker(time.Duration(ASSIGNBREAKOUTPERIOD) * time.Second)
				for {
					select {
					case <-assignBreakoutTimer.C:
						c_assignmentSuccessfull <- false
						assignBreakoutTimer.Stop()
						break PollLastAssigned
					case <-c_quit:
						return
					default:
						connectedNodes := f_GetConnectedNodes()
						globalQueue := f_GetGlobalQueue()
						updatedEntry := f_FindEntry(lastAssignedEntry.Request.Id, lastAssignedEntry.RequestedNode, globalQueue)
						updatedAssignedNode := f_FindNodeInfo(lastAssignedEntry.AssignedNode, connectedNodes)
						if updatedAssignedNode.ElevatorInfo.State == elevator.MOVING || updatedEntry.Request.State == elevator.ACTIVE {
							c_assignmentSuccessfull <- true
							F_WriteLog("Found ack")
							assignBreakoutTimer.Stop()
							break PollLastAssigned
						}
						//No sleep
					}
				}
			}
			time.Sleep(time.Duration(MEDIUMRESPONSIVEPERIOD) * time.Microsecond)
		}
	}()

	for {
		select {
		case <-c_quit:
			return
		default: //JONASCOMMENT: mye her inne kan også være egne funksjoner
			thisNodeInfo := f_GetNodeInfo()
			c_getSetGlobalQueueInterface <- getSetGlobalQueueInterface
			globalQueue := <-getSetGlobalQueueInterface.c_get

			//JONASCOMMENT: dette kan nok også være en egen funksjon, altså herifra til....
			unServicedEntry, servicedEntry := T_GlobalQueueEntry{}, T_GlobalQueueEntry{}
			unServicedEntryIndex, servicedEntryIdex := 0, 0
			for i, entry := range globalQueue {
				if entry.TimeUntilReassign == 0 {
					if globalQueue[i].Request.State != elevator.DONE {
						unServicedEntry = globalQueue[i]
						unServicedEntryIndex = i
						break
					} else {
						servicedEntry = globalQueue[i]
						servicedEntryIdex = i
						break
					}
				}
			}
			//JONASCOMMENT: ...hit

			//JONASCOMMENT: dette tror jeg også passer bra til å være egen funksjon, så herifra til...
			if (unServicedEntry != T_GlobalQueueEntry{}) {
				unServicedEntry.Request.State = elevator.UNASSIGNED
				entryToReassign := T_GlobalQueueEntry{
					Request:           unServicedEntry.Request,
					RequestedNode:     unServicedEntry.RequestedNode,
					AssignedNode:      0,
					TimeUntilReassign: REASSIGNTIME,
				}
				F_WriteLog("Reassigned entry: | " + strconv.Itoa(int(unServicedEntry.Request.Id)) + " | " + strconv.Itoa(int(unServicedEntry.RequestedNode)) + " | in global queue")
				globalQueue[unServicedEntryIndex] = entryToReassign
			}
			//JONASCOMMENT: ...hit
			getSetGlobalQueueInterface.c_set <- globalQueue

			//JONASCOMMENT: tenker egt samme her, føler dette passer bra til å være egen funksjon, så herifra til...
			if (servicedEntry != T_GlobalQueueEntry{}) {
				c_sentDoneEntryToSlave := make(chan bool)
				ackSentEntryToSlave := T_AckObject{
					ObjectToAcknowledge:        globalQueue,
					ObjectToSupportAcknowledge: thisNodeInfo,
					C_Acknowledgement:          c_sentDoneEntryToSlave,
				}
				c_ackSentEntryToSlave <- ackSentEntryToSlave
				breakOutTimer := time.NewTicker(time.Duration(1000) * time.Millisecond)
				F_WriteLog("MASTER found done entry, waiting for sending to slave before removing")
				select {
				case <-c_sentDoneEntryToSlave:
					F_WriteLog("Removed entry: | " + strconv.Itoa(int(servicedEntry.Request.Id)) + " | " + strconv.Itoa(int(servicedEntry.RequestedNode)) + " | from global queue")
					globalQueue = append(globalQueue[:servicedEntryIdex], globalQueue[servicedEntryIdex+1:]...)
					f_SetGlobalQueue(globalQueue)
				case <-breakOutTimer.C:
				}
			}
			//JONASCOMMENT: ...hit
			time.Sleep(time.Duration(LEASTRESPONSIVEPERIOD) * time.Microsecond)
		}
	}
}

// JONASCOMMENT: denne kan også deles tror jeg
func f_SlaveVariableWatchDog(c_getSetConnectedNodesInterface chan T_GetSetConnectedNodesInterface, getSetConnectedNodesInterface T_GetSetConnectedNodesInterface, c_quit chan bool) {
	for {
		select {
		case <-c_quit:
			return
		default:
			c_getSetConnectedNodesInterface <- getSetConnectedNodesInterface
			connectedNodes := <-getSetConnectedNodesInterface.c_get

			nodeToDisconnect := T_NodeInfo{}
			nodeToDisconnectIndex := 0
			for i, nodeInfo := range connectedNodes {
				if nodeInfo.TimeUntilDisconnect == 0 {
					nodeToDisconnect = nodeInfo
					nodeToDisconnectIndex = i
					break
				}
			}
			if (nodeToDisconnect != T_NodeInfo{}) {
				connectedNodes = append(connectedNodes[:nodeToDisconnectIndex], connectedNodes[nodeToDisconnectIndex+1:]...)
				F_WriteLog("Node " + strconv.Itoa(int(nodeToDisconnect.PRIORITY)) + " disconnected")
			}
			getSetConnectedNodesInterface.c_set <- connectedNodes

			time.Sleep(time.Duration(LEASTRESPONSIVEPERIOD) * time.Microsecond)
		}
	}
}

// JONASCOMMENT: Generelt sett ganske nice, har en liten bit jeg ville gjort til funksjon bare
func f_MasterTimeManager(c_getSetGlobalQueueInterface chan T_GetSetGlobalQueueInterface, getSetGlobalQueueInterface T_GetSetGlobalQueueInterface, c_quit chan bool) {
	for {
		select {
		case <-c_quit:
			return
		default:
			c_getSetGlobalQueueInterface <- getSetGlobalQueueInterface
			globalQueue := <-getSetGlobalQueueInterface.c_get

			for i, entry := range globalQueue {
				if entry.TimeUntilReassign > 0 && entry.Request.State != elevator.UNASSIGNED {
					globalQueue[i].TimeUntilReassign -= 1
				}
			}
			getSetGlobalQueueInterface.c_set <- globalQueue

			time.Sleep(1 * time.Second)
		}
	}
}

// JONASCOMMENT: nesten samme som i mastertimemanager
func f_SlaveTimeManager(c_getSetConnectedNodesInterface chan T_GetSetConnectedNodesInterface, getSetConnectedNodesInterface T_GetSetConnectedNodesInterface, c_quit chan bool) {
	for {
		select {
		case <-c_quit:
			return
		default:
			c_getSetConnectedNodesInterface <- getSetConnectedNodesInterface
			connectedNodes := <-getSetConnectedNodesInterface.c_get
			for i, connectedNode := range connectedNodes {
				if connectedNode.TimeUntilDisconnect > 0 {
					connectedNodes[i].TimeUntilDisconnect -= 1
				}
			}
			getSetConnectedNodesInterface.c_set <- connectedNodes

			time.Sleep(1 * time.Second)
		}
	}
}
func f_UpdateConnectedNodes(c_getSetConnectedNodesInterface chan T_GetSetConnectedNodesInterface, getSetConnectedNodesInterface T_GetSetConnectedNodesInterface, currentNode T_NodeInfo) {
	c_getSetConnectedNodesInterface <- getSetConnectedNodesInterface
	oldConnectedNodes := <-getSetConnectedNodesInterface.c_get

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
		getSetConnectedNodesInterface.c_set <- connectedNodes
	} else {
		currentNode.TimeUntilDisconnect = CONNECTIONTIME
		oldConnectedNodes[nodeIndex] = currentNode
		getSetConnectedNodesInterface.c_set <- oldConnectedNodes
	}
}
func f_AddEntryGlobalQueue(c_getSetGlobalQueueInterface chan T_GetSetGlobalQueueInterface, getSetGlobalQueueInterface T_GetSetGlobalQueueInterface, entryToAdd T_GlobalQueueEntry) {
	c_getSetGlobalQueueInterface <- getSetGlobalQueueInterface
	globalQueue := <-getSetGlobalQueueInterface.c_get

	entryIsUnique := true
	entryIndex := 0
	for i, entry := range globalQueue {
		if entryToAdd.Request.Id == entry.Request.Id && entryToAdd.RequestedNode == entry.RequestedNode { //random id generated to each entry
			entryIsUnique = false
			entryIndex = i
			break
		}
	}
	if entryIsUnique {
		globalQueue = append(globalQueue, entryToAdd)
	} else { //should update the existing entry
		if entryToAdd.Request.State >= globalQueue[entryIndex].Request.State || entryToAdd.TimeUntilReassign <= globalQueue[entryIndex].TimeUntilReassign { //only allow forward entry states //>=?
			globalQueue[entryIndex] = entryToAdd
		} else { //you get a entry that has come longer somewhere else, use its values
			F_WriteLog("Disallowed backward information")
		}
	}
	getSetGlobalQueueInterface.c_set <- globalQueue
}
func f_UpdateGlobalQueueMaster(c_getSetGlobalQueueInterface chan T_GetSetGlobalQueueInterface, getSetGlobalQueueInterface T_GetSetGlobalQueueInterface, masterMessage T_MasterMessage) {
	for _, remoteEntry := range masterMessage.GlobalQueue {
		f_AddEntryGlobalQueue(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, remoteEntry)
	}
}
func f_UpdateGlobalQueueSlave(c_getSetGlobalQueueInterface chan T_GetSetGlobalQueueInterface, getSetGlobalQueueInterface T_GetSetGlobalQueueInterface, masterMessage T_MasterMessage) {
	entriesToRemove := []T_GlobalQueueEntry{}
	for _, remoteEntry := range masterMessage.GlobalQueue {
		f_AddEntryGlobalQueue(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, remoteEntry)
		if remoteEntry.Request.State == elevator.DONE {
			entriesToRemove = append(entriesToRemove, remoteEntry)
		}
	}
	//JONASCOMMENT: kan removeEntriesFromGlobalQueue være en egen funksjon?
	//JONASCOMMENT: så kan man legge til at hvis entriesToRemove er tom, så hopper vi over de neste stegene
	c_getSetGlobalQueueInterface <- getSetGlobalQueueInterface
	oldGlobalQueue := <-getSetGlobalQueueInterface.c_get

	newGlobalQueue := oldGlobalQueue
	for i, entry := range oldGlobalQueue {
		for _, entryToRemove := range entriesToRemove {
			if entry.Request.Id == entryToRemove.Request.Id && entry.RequestedNode == entryToRemove.RequestedNode {
				newGlobalQueue = append(oldGlobalQueue[:i], oldGlobalQueue[i+1:]...)
			}
		}
	}
	getSetGlobalQueueInterface.c_set <- newGlobalQueue

}

// JONASCOMMENT: igjen, denne kan nok kuttes ned, synes at 50 linjer er litt mye
func f_ElevatorManager(c_shouldCheckIfAssigned chan bool, c_entryFromElevator chan T_GlobalQueueEntry) {
	c_requestFromElevator := make(chan elevator.T_Request)
	c_requestToElevator := make(chan elevator.T_Request)
	shouldCheckIfAssigned := true

	//go elevator.F_RunElevator(elevatorOps, c_requestFromElevator, c_requestToElevator, ELEVATORPORT)
	go elevator.F_SimulateRequest(c_requestFromElevator, c_requestToElevator)

	thisNodeInfo := f_GetNodeInfo()
	globalQueue := f_GetGlobalQueue()
	assignedEntry, _ := F_FindAssignedEntry(globalQueue, thisNodeInfo)
	for {
		select {
		case receivedRequest := <-c_requestFromElevator:
			newEntry := T_GlobalQueueEntry{}
			if receivedRequest.State == elevator.DONE {
				F_WriteLog("Node: | " + strconv.Itoa(int(thisNodeInfo.PRIORITY)) + " | request resent DONE")
				newEntry = T_GlobalQueueEntry{
					Request:           receivedRequest,
					RequestedNode:     assignedEntry.RequestedNode,
					AssignedNode:      assignedEntry.AssignedNode,
					TimeUntilReassign: 0,
				}
			} else if receivedRequest.State == elevator.ACTIVE {
				F_WriteLog("Node: | " + strconv.Itoa(int(thisNodeInfo.PRIORITY)) + " | request resent ACTIVE")
				newEntry = T_GlobalQueueEntry{
					Request:           receivedRequest,
					RequestedNode:     assignedEntry.RequestedNode,
					AssignedNode:      assignedEntry.AssignedNode,
					TimeUntilReassign: REASSIGNTIME,
				}
			} else if receivedRequest.State == elevator.UNASSIGNED {
				newEntry = T_GlobalQueueEntry{
					Request:           receivedRequest,
					RequestedNode:     thisNodeInfo.PRIORITY,
					AssignedNode:      0,
					TimeUntilReassign: REASSIGNTIME,
				}
			} else {
				F_WriteLog("Error: Received Assigned request from elevator")
			}
			c_entryFromElevator <- newEntry
		case <-c_shouldCheckIfAssigned:
			shouldCheckIfAssigned = true
		default:
			if shouldCheckIfAssigned {
				thisNodeInfo = f_GetNodeInfo()
				globalQueue = f_GetGlobalQueue()
				assignedEntry, _ = F_FindAssignedEntry(globalQueue, thisNodeInfo)
				if (assignedEntry != T_GlobalQueueEntry{}) {
					F_WriteLog("Found assigned entry!")
					c_requestToElevator <- assignedEntry.Request //NB! Depending on that elevator is polling in IDLE, Breakout here?
					shouldCheckIfAssigned = false
				}
			}
			time.Sleep(time.Duration(MOSTRESPONSIVEPERIOD) * time.Microsecond)
		}
	}
}

//IMPORTANT:
//-global variables should ALWAYS be handled by server to operate onn good data
//-all receive from channles should be organized in for-select!!! -> walk trough code and do

// should contain the main master/slave fsm in Run() function, to be called from main
func F_RunNode() {

	getSetNodeInfoInterface := T_GetSetNodeInfoInterface{
		c_get: make(chan T_NodeInfo),
		c_set: make(chan T_NodeInfo),
	}
	getSetGlobalQueueInterface := T_GetSetGlobalQueueInterface{
		c_get: make(chan []T_GlobalQueueEntry),
		c_set: make(chan []T_GlobalQueueEntry),
	}
	getSetConnectedNodesInterface := T_GetSetConnectedNodesInterface{
		c_get: make(chan []T_NodeInfo),
		c_set: make(chan []T_NodeInfo),
	}

	c_getSetNodeInfoInterface := make(chan T_GetSetNodeInfoInterface)
	c_getSetGlobalQueueInterface := make(chan T_GetSetGlobalQueueInterface)
	c_getSetConnectedNodesInterface := make(chan T_GetSetConnectedNodesInterface)

	//to run the main FSM
	c_receiveSlaveMessage := make(chan T_SlaveMessage)
	c_receiveMasterMessage := make(chan T_MasterMessage)
	c_transmitMasterMessage := make(chan T_MasterMessage)
	c_transmitSlaveMessage := make(chan T_SlaveMessage)

	c_entryFromElevator := make(chan T_GlobalQueueEntry)
	c_lastAssignedEntry := make(chan T_GlobalQueueEntry)
	c_assignmentWasSucessFull := make(chan bool)
	c_shouldCheckIfAssigned := make(chan bool)
	c_nodeIsMaster := make(chan bool)
	c_nodeIsSlave := make(chan bool)
	c_ackSentGlobalQueueToSlave := make(chan T_AckObject)

	c_quitMasterVariableWatchdog := make(chan bool)
	c_quitMasterTimeManager := make(chan bool)
	c_quitSlaveVariableWatchdog := make(chan bool)
	c_quitSlaveTimeManager := make(chan bool)

	go func() {
		go f_NodeOperationManager(&ThisNode) //SHOULD BE THE ONLY REFERENCE TO ThisNode!

		go f_GetSetNodeInfo(c_getSetNodeInfoInterface)
		go f_GetSetGlobalQueue(c_getSetGlobalQueueInterface)
		go f_GetSetConnectedNodes(c_getSetConnectedNodesInterface)

		go f_ElevatorManager(c_shouldCheckIfAssigned, c_entryFromElevator)
		go F_ReceiveSlaveMessage(c_receiveSlaveMessage, SLAVEPORT)
		go F_ReceiveMasterMessage(c_receiveMasterMessage, MASTERPORT)
		go F_TransmitSlaveMessage(c_transmitSlaveMessage, SLAVEPORT)
		go F_TransmitMasterMessage(c_transmitMasterMessage, MASTERPORT)
		for {
			select {
			case <-c_nodeIsMaster:
				go f_MasterVariableWatchDog(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, c_lastAssignedEntry, c_assignmentWasSucessFull, c_ackSentGlobalQueueToSlave, c_quitMasterVariableWatchdog)
				go f_SlaveVariableWatchDog(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface, c_quitSlaveVariableWatchdog)
				go f_MasterTimeManager(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, c_quitMasterTimeManager)
				go f_SlaveTimeManager(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface, c_quitSlaveTimeManager)
			case <-c_nodeIsSlave:
				go f_SlaveVariableWatchDog(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface, c_quitSlaveVariableWatchdog)
				go f_SlaveTimeManager(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface, c_quitSlaveTimeManager)
			default:
				time.Sleep(time.Duration(LEASTRESPONSIVEPERIOD) * time.Microsecond)
			}
		}
	}()

	sendTimer := time.NewTicker(time.Duration(SENDPERIOD) * time.Millisecond)
	printGQTimer := time.NewTicker(time.Duration(2000) * time.Millisecond) //Test function
	assignState := ASSIGN

	nodeRole := f_GetNodeInfo().Role
	if nodeRole == MASTER {
		c_nodeIsMaster <- true
	} else {
		c_nodeIsSlave <- true
	}

	for {
		nodeRole = f_GetNodeInfo().Role
		switch nodeRole {
		case MASTER:
			select {
			case masterMessage := <-c_receiveMasterMessage:
				f_WriteLogMasterMessage(masterMessage)
				f_UpdateConnectedNodes(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface, masterMessage.Transmitter)
				f_WriteLogConnectedNodes(f_GetConnectedNodes())
				thisNode := f_GetNodeInfo()
				if masterMessage.Transmitter.PRIORITY != thisNode.PRIORITY {
					f_UpdateGlobalQueueMaster(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, masterMessage)
				}

			case slaveMessage := <-c_receiveSlaveMessage:

				f_WriteLogSlaveMessage(slaveMessage)
				f_UpdateConnectedNodes(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface, slaveMessage.Transmitter)
				f_WriteLogConnectedNodes(f_GetConnectedNodes())
				if slaveMessage.Entry.Request.Calltype != elevator.NONECALL {
					f_AddEntryGlobalQueue(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, slaveMessage.Entry)
				}

			case entryFromElevator := <-c_entryFromElevator:
				f_AddEntryGlobalQueue(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, entryFromElevator)
				if entryFromElevator.Request.State == elevator.DONE {
					c_shouldCheckIfAssigned <- true
				}

				thisNode := f_GetNodeInfo()
				F_WriteLog("Node: | " + strconv.Itoa(int(thisNode.PRIORITY)) + " | MASTER | updated GQ entry:\n")
				F_WriteLogGlobalQueueEntry(entryFromElevator)

			case ackSentGlobalQueueToSlave := <-c_ackSentGlobalQueueToSlave:
				transmitterNodeInfo, nodeInfoTypeOk := ackSentGlobalQueueToSlave.ObjectToSupportAcknowledge.(T_NodeInfo)
				globalQueue, globalQueueTypeOk := ackSentGlobalQueueToSlave.ObjectToAcknowledge.([]T_GlobalQueueEntry)
				if !globalQueueTypeOk || !nodeInfoTypeOk {
					F_WriteLog("Ack operation failed in sending to slave, wrong type")
				} else {
					masterMessage := T_MasterMessage{
						Transmitter: transmitterNodeInfo,
						GlobalQueue: globalQueue,
					}
					c_transmitMasterMessage <- masterMessage
					ackSentGlobalQueueToSlave.C_Acknowledgement <- true
				}

			case <-sendTimer.C:
				transmitterNodeInfo := f_GetNodeInfo()
				masterMessage := T_MasterMessage{
					Transmitter: transmitterNodeInfo,
					GlobalQueue: f_GetGlobalQueue(),
				}
				c_transmitMasterMessage <- masterMessage
				sendTimer.Reset(time.Duration(SENDPERIOD) * time.Millisecond)

			case <-printGQTimer.C:
				globalQueue := f_GetGlobalQueue()
				nodeInfo := f_GetNodeInfo()
				F_WriteLog("Assignstate: | " + strconv.Itoa(int(assignState)) + " |")
				F_WriteLog("Node: | " + strconv.Itoa(int(nodeInfo.PRIORITY)) + " | MASTER | has GQ:\n")
				for _, entry := range globalQueue {
					F_WriteLogGlobalQueueEntry(entry)
				}
				printGQTimer.Reset(time.Duration(2000) * time.Millisecond)

			default:
				connectedNodes := f_GetConnectedNodes()

				c_getSetNodeInfoInterface <- getSetNodeInfoInterface
				oldNodeInfo := <-getSetNodeInfoInterface.c_get
				newNodeInfo := f_AssignNewRole(oldNodeInfo, connectedNodes)
				getSetNodeInfoInterface.c_set <- newNodeInfo

				thisNodeInfo := newNodeInfo
				f_UpdateConnectedNodes(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface, thisNodeInfo) //Update connected nodes with newnodeinfo

				//Need to be in own FSM
				switch assignState {
				case ASSIGN:
					connectedNodes := f_GetConnectedNodes()
					avalibaleNodes := f_GetAvalibaleNodes(connectedNodes)
					c_getSetGlobalQueueInterface <- getSetGlobalQueueInterface
					globalQueue := <-getSetGlobalQueueInterface.c_get

					assignedEntry, assignedEntryIndex := F_AssignNewEntry(globalQueue, connectedNodes, avalibaleNodes)
					if (assignedEntry != T_GlobalQueueEntry{}) {
						globalQueue[assignedEntryIndex] = assignedEntry
						c_lastAssignedEntry <- assignedEntry
						F_WriteLog("Assigned request with ID: " + strconv.Itoa(int(assignedEntry.Request.Id)) + " assigned to node " + strconv.Itoa(int(assignedEntry.AssignedNode)))
						assignState = WAITFORACK
					}

					getSetGlobalQueueInterface.c_set <- globalQueue

				case WAITFORACK:
					select {
					case assigmentWasSucessFull := <-c_assignmentWasSucessFull:
						if assigmentWasSucessFull {
							assignState = ASSIGN
						} else {
							assignState = ASSIGN //Start reassigning after 10 secs anyways
						}
					default:
					}
				}

				if newNodeInfo.Role == SLAVE {
					c_quitMasterVariableWatchdog <- true
					c_quitMasterTimeManager <- true
					c_nodeIsSlave <- true
					assignState = ASSIGN
				}

			}

		case SLAVE:
			select {
			case masterMessage := <-c_receiveMasterMessage:
				f_WriteLogMasterMessage(masterMessage)
				f_UpdateGlobalQueueSlave(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, masterMessage)
				f_UpdateConnectedNodes(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface, masterMessage.Transmitter)

			case slaveMessage := <-c_receiveSlaveMessage:
				f_WriteLogSlaveMessage(slaveMessage)
				f_UpdateConnectedNodes(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface, slaveMessage.Transmitter)

			case entryFromElevator := <-c_entryFromElevator:
				f_AddEntryGlobalQueue(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, entryFromElevator)
				if entryFromElevator.Request.State == elevator.DONE {
					c_shouldCheckIfAssigned <- true
				}
				thisNode := f_GetNodeInfo()
				F_WriteLog("Node: | " + strconv.Itoa(int(thisNode.PRIORITY)) + " | SLAVE | updated GQ entry:\n")
				F_WriteLogGlobalQueueEntry(entryFromElevator)

				transmitter := f_GetNodeInfo()
				infoMessage := T_SlaveMessage{
					Transmitter: transmitter,
					Entry:       entryFromElevator,
				}
				c_transmitSlaveMessage <- infoMessage
			case <-sendTimer.C:
				transmitter := f_GetNodeInfo()
				aliveMessage := T_SlaveMessage{
					Transmitter: transmitter,
					Entry:       T_GlobalQueueEntry{},
				}
				c_transmitSlaveMessage <- aliveMessage
				sendTimer.Reset(time.Duration(SENDPERIOD) * time.Millisecond)
			case <-printGQTimer.C:
				globalQueue := f_GetGlobalQueue()
				nodeInfo := f_GetNodeInfo()
				F_WriteLog("Node: | " + strconv.Itoa(int(nodeInfo.PRIORITY)) + " | SLAVE | has GQ:\n")
				for _, entry := range globalQueue {
					F_WriteLogGlobalQueueEntry(entry)
				}
				printGQTimer.Reset(time.Duration(2000) * time.Millisecond)
			default:
				connectedNodes := f_GetConnectedNodes()

				c_getSetNodeInfoInterface <- getSetNodeInfoInterface
				oldNodeInfo := <-getSetNodeInfoInterface.c_get
				newNodeInfo := f_AssignNewRole(oldNodeInfo, connectedNodes)
				getSetNodeInfoInterface.c_set <- newNodeInfo

				thisNodeInfo := newNodeInfo
				f_UpdateConnectedNodes(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface, thisNodeInfo)

				if newNodeInfo.Role == MASTER {
					c_nodeIsMaster <- true
					c_quitSlaveVariableWatchdog <- true
					c_quitSlaveTimeManager <- true
				}
			}
		}
		time.Sleep(time.Duration(MOSTRESPONSIVEPERIOD) * time.Microsecond)
	}
}
