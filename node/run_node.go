package node

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"the-elevator/node/elevator"
	"time"
)

func f_simulateRequest(nodeOps T_NodeOperations, elevatorOps elevator.T_ElevatorOperations, c_requestFromElevator chan elevator.T_Request, c_requestToElevator chan elevator.T_Request) {
	increment := uint16(0)
	go func() {
		for {
			select {
			case request := <-c_requestToElevator:
				c_readElevator0 := make(chan elevator.T_Elevator)
				c_writeElevator0 := make(chan elevator.T_Elevator)
				c_quitGetSetElevator0 := make(chan bool)
				go elevator.F_GetAndSetElevator(elevatorOps, c_readElevator0, c_writeElevator0, c_quitGetSetElevator0)
				currentElevator0 := <-c_readElevator0
				(*currentElevator0.P_info).State = elevator.MOVING
				c_writeElevator0 <- currentElevator0
				c_quitGetSetElevator0 <- true

				request.State = elevator.ACTIVE
				c_requestFromElevator <- request

				time.Sleep(10 * time.Second)

				c_readElevatorInfo := make(chan elevator.T_Elevator)
				c_writeElevatorInfo := make(chan elevator.T_Elevator)
				c_quitGetSetElevatorInfo := make(chan bool)
				go elevator.F_GetAndSetElevator(elevatorOps, c_readElevatorInfo, c_writeElevatorInfo, c_quitGetSetElevatorInfo)
				currentElevator := <-c_readElevatorInfo
				(*currentElevator0.P_info).State = elevator.IDLE
				c_writeElevatorInfo <- currentElevator
				c_quitGetSetElevatorInfo <- true

				request.State = elevator.DONE
				c_requestFromElevator <- request
			default:
				time.Sleep(time.Duration(LEASTRESPONSIVEPERIOD) * time.Microsecond)
			}
		}
	}()

	for {
		var input string
		fmt.Println("Enter request (C/H-floor): ")
		fmt.Scanln(&input)
		delimiter := "-"
		parts := strings.Split(input, delimiter)
		partToConvert := parts[1]
		floor, _ := strconv.Atoi(partToConvert)
		var returnRequest elevator.T_Request
		if parts[0] == "C" {
			returnRequest = elevator.T_Request{
				Id:        increment,
				State:     elevator.UNASSIGNED,
				Calltype:  elevator.CAB,
				Floor:     int8(floor),
				Direction: elevator.UP,
			}
			increment += 1
			c_requestFromElevator <- returnRequest
		} else if parts[0] == "H" {
			returnRequest = elevator.T_Request{
				Id:        increment,
				State:     elevator.UNASSIGNED,
				Calltype:  elevator.HALL,
				Floor:     int8(floor),
				Direction: elevator.UP,
			}
			increment += 1
			c_requestFromElevator <- returnRequest
		}
		time.Sleep(time.Duration(LEASTRESPONSIVEPERIOD) * time.Microsecond)
	}
}

// ***	END TEST FUNCTIONS	***//

func f_InitNode(config T_Config) T_Node {
	thisElevatorInfo := elevator.T_ElevatorInfo{
		Direction: elevator.NONE,
		Floor:     -1,
		State:     elevator.IDLE,
	}
	thisNodeInfo := T_NodeInfo{
		PRIORITY:     config.Priority,
		ElevatorInfo: thisElevatorInfo,
		Role:         MASTER,
	}

	thisElevator := elevator.T_Elevator{
		P_info:         &thisElevatorInfo,
		P_serveRequest: &elevator.T_Request{},
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
func f_WriteLogConnectedNodes(ops T_NodeOperations, connectedNodes []T_NodeInfo) {
	thisNode := f_GetNodeInfo(ops)
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
func f_WriteLogSlaveMessage(ops T_NodeOperations, slaveMessage T_SlaveMessage) {
	request := slaveMessage.Entry.Request
	thisNode := f_GetNodeInfo(ops)
	logStr := fmt.Sprintf("Node: | %d | %s | received SM from | %d | Request ID: | %d | State: | %s | Calltype: %s | Floor: %d | Direction: %s |",
		int(thisNode.PRIORITY), f_NodeRoleToString(thisNode.Role), int(slaveMessage.Transmitter.PRIORITY), int(request.Id), f_RequestStateToString(slaveMessage.Entry.Request.State), f_CallTypeToString(request.Calltype), int(request.Floor), f_DirectionToString(request.Direction))
	F_WriteLog(logStr)
}
func f_WriteLogMasterMessage(ops T_NodeOperations, masterMessage T_MasterMessage) {
	thisNode := f_GetNodeInfo(ops)
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
func f_FindNodeInfo(node uint8, connectedNodes []T_NodeInfo) T_NodeInfo {
	returnNode := T_NodeInfo{}
	for _, nodeInfo := range connectedNodes {
		if node == nodeInfo.PRIORITY {
			returnNode = nodeInfo
			break
		}
	}
	return returnNode
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
func f_FindEntry(id uint16, requestedNode uint8, globalQueue []T_GlobalQueueEntry) T_GlobalQueueEntry {
	returnEntry := T_GlobalQueueEntry{}
	for _, entry := range globalQueue {
		if id == entry.Request.Id && entry.RequestedNode == requestedNode {
			returnEntry = entry
		}
	}
	return returnEntry
}
func f_MasterVariableWatchDog(ops T_NodeOperations, c_lastAssignedEntry chan T_GlobalQueueEntry, c_assignmentSuccessfull chan bool, c_ackSentEntryToSlave chan T_AckObject, c_quit chan bool) {
	go f_SlaveVariableWatchDog(ops, c_quit)
	go func() {
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
						connectedNodes := f_GetConnectedNodes(ops)
						globalQueue := f_GetGlobalQueue(ops)
						updatedEntry := f_FindEntry(lastAssignedEntry.Request.Id, lastAssignedEntry.RequestedNode, globalQueue)
						updatedAssignedNode := f_FindNodeInfo(lastAssignedEntry.AssignedNode, connectedNodes)
						if updatedAssignedNode.ElevatorInfo.State != elevator.IDLE || updatedEntry.Request.State != elevator.ASSIGNED {
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
		default:
			thisNodeInfo := f_GetNodeInfo(ops)
			c_readGlobalQueue := make(chan []T_GlobalQueueEntry)
			c_writeGlobalQueue := make(chan []T_GlobalQueueEntry)
			c_quitGetSetGlobalQueue := make(chan bool)
			go f_GetAndSetGlobalQueue(ops, c_readGlobalQueue, c_writeGlobalQueue, c_quitGetSetGlobalQueue)
			globalQueue := <-c_readGlobalQueue

			unServicedEntry := T_GlobalQueueEntry{}
			unServicedEntryIndex := 0
			servicedEntry := T_GlobalQueueEntry{}
			servicedEntryIdex := 0
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
			c_writeGlobalQueue <- globalQueue
			c_quitGetSetGlobalQueue <- true

			if (servicedEntry != T_GlobalQueueEntry{}) {
				c_sentDoneEntryToSlave := make(chan bool)
				ackSentEntryToSlave := T_AckObject{
					ObjectToAcknowledge:        globalQueue,
					ObjectToSupportAcknowledge: thisNodeInfo,
					C_Acknowledgement:          c_sentDoneEntryToSlave,
				}
				F_WriteLog("Finds here")
				c_ackSentEntryToSlave <- ackSentEntryToSlave
				breakOutTimer := time.NewTicker(time.Duration(1000) * time.Millisecond)
				F_WriteLog("MASTER found done entry, waiting for sending to slave before removing")
				select {
				case <-c_sentDoneEntryToSlave:
					F_WriteLog("Removed entry: | " + strconv.Itoa(int(servicedEntry.Request.Id)) + " | " + strconv.Itoa(int(servicedEntry.RequestedNode)) + " | from global queue")
					globalQueue = append(globalQueue[:servicedEntryIdex], globalQueue[servicedEntryIdex+1:]...)
					f_SetGlobalQueue(ops, globalQueue)
				case <-breakOutTimer.C:
				}
			}
			time.Sleep(time.Duration(LEASTRESPONSIVEPERIOD) * time.Microsecond)
		}
	}
}
func f_SlaveVariableWatchDog(ops T_NodeOperations, c_quit chan bool) {
	for {
		select {
		case <-c_quit:
			return
		default:
			c_readConnectedNodes := make(chan []T_NodeInfo)
			c_writeConnectedNodes := make(chan []T_NodeInfo)
			c_quitGetSetConnectedNodes := make(chan bool)
			go f_GetAndSetConnectedNodes(ops, c_readConnectedNodes, c_writeConnectedNodes, c_quitGetSetConnectedNodes)
			connectedNodes := <-c_readConnectedNodes

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
			c_writeConnectedNodes <- connectedNodes
			c_quitGetSetConnectedNodes <- true

			time.Sleep(time.Duration(LEASTRESPONSIVEPERIOD) * time.Microsecond)
		}
	}
}
func f_MasterTimeManager(ops T_NodeOperations, c_quit chan bool) {
	go f_SlaveTimeManager(ops, c_quit)
	for {
		select {
		case <-c_quit:
			return
		default:
			c_readGlobalQueue := make(chan []T_GlobalQueueEntry)
			c_writeGlobalQueue := make(chan []T_GlobalQueueEntry)
			c_quitGetSetGlobalQueue := make(chan bool)
			go f_GetAndSetGlobalQueue(ops, c_readGlobalQueue, c_writeGlobalQueue, c_quitGetSetGlobalQueue)
			globalQueue := <-c_readGlobalQueue

			for i, entry := range globalQueue {
				if entry.TimeUntilReassign > 0 && entry.Request.State != elevator.UNASSIGNED {
					globalQueue[i].TimeUntilReassign -= 1
				}
			}
			//remove all entries being DONE and
			c_writeGlobalQueue <- globalQueue
			c_quitGetSetGlobalQueue <- true

			time.Sleep(1 * time.Second)
		}
	}
}
func f_SlaveTimeManager(ops T_NodeOperations, c_quit chan bool) {
	for {
		select {
		case <-c_quit:
			return
		default:
			c_readConnectedNodes := make(chan []T_NodeInfo)
			c_writeConnectedNodes := make(chan []T_NodeInfo)
			c_quitGetSetConnectedNodes := make(chan bool)
			go f_GetAndSetConnectedNodes(ops, c_readConnectedNodes, c_writeConnectedNodes, c_quitGetSetConnectedNodes)
			connectedNodes := <-c_readConnectedNodes

			for i := range connectedNodes {
				if connectedNodes[i].TimeUntilDisconnect > 0 {
					connectedNodes[i].TimeUntilDisconnect -= 1
				}
			}
			c_writeConnectedNodes <- connectedNodes
			c_quitGetSetConnectedNodes <- true

			time.Sleep(1 * time.Second)
		}
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
func f_AddEntryGlobalQueue(nodeOps T_NodeOperations, entryToAdd T_GlobalQueueEntry) {
	c_readGlobalQueue := make(chan []T_GlobalQueueEntry)
	c_writeGlobalQueue := make(chan []T_GlobalQueueEntry)
	c_quit := make(chan bool)
	go f_GetAndSetGlobalQueue(nodeOps, c_readGlobalQueue, c_writeGlobalQueue, c_quit)
	globalQueue := <-c_readGlobalQueue

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
	c_writeGlobalQueue <- globalQueue
	c_quit <- true
}
func f_UpdateGlobalQueueMaster(nodeOps T_NodeOperations, masterMessage T_MasterMessage) {
	for _, remoteEntry := range masterMessage.GlobalQueue {
		f_AddEntryGlobalQueue(nodeOps, remoteEntry)
	}
}
func f_UpdateGlobalQueueSlave(nodeOps T_NodeOperations, masterMessage T_MasterMessage) {
	entriesToRemove := []T_GlobalQueueEntry{}
	for _, remoteEntry := range masterMessage.GlobalQueue {
		f_AddEntryGlobalQueue(nodeOps, remoteEntry)
		if remoteEntry.Request.State == elevator.DONE {
			entriesToRemove = append(entriesToRemove, remoteEntry)
		}
	}
	c_readGlobalQueue := make(chan []T_GlobalQueueEntry)
	c_writeGlobalQueue := make(chan []T_GlobalQueueEntry)
	c_quit := make(chan bool)
	go f_GetAndSetGlobalQueue(nodeOps, c_readGlobalQueue, c_writeGlobalQueue, c_quit)
	oldGlobalQueue := <-c_readGlobalQueue
	newGlobalQueue := oldGlobalQueue
	for i, entry := range oldGlobalQueue {
		for _, entryToRemove := range entriesToRemove {
			if entry.Request.Id == entryToRemove.Request.Id && entry.RequestedNode == entryToRemove.RequestedNode {
				newGlobalQueue = append(oldGlobalQueue[:i], oldGlobalQueue[i+1:]...)
			}
		}
	}
	c_writeGlobalQueue <- newGlobalQueue
	c_quit <- true

}
func f_ElevatorManager(nodeOps T_NodeOperations, elevatorOps elevator.T_ElevatorOperations, c_shouldCheckIfAssigned chan bool, c_entryFromElevator chan T_GlobalQueueEntry) {
	c_requestFromElevator := make(chan elevator.T_Request)
	c_requestToElevator := make(chan elevator.T_Request)
	shouldCheckIfAssigned := true

	go elevator.F_RunElevator(elevatorOps, c_requestFromElevator, c_requestToElevator, ELEVATORPORT)
	//go f_simulateRequest(nodeOps, elevatorOps, c_requestFromElevator, c_requestToElevator)

	thisNodeInfo := f_GetNodeInfo(nodeOps)
	globalQueue := f_GetGlobalQueue(nodeOps)
	assignedEntry, _ := F_FindAssignedEntry(globalQueue, thisNodeInfo)
	for {
		select {
		case receivedRequest := <-c_requestFromElevator:
			newEntry := T_GlobalQueueEntry{}
			if receivedRequest.State == elevator.DONE {
				newEntry = T_GlobalQueueEntry{
					Request:           receivedRequest,
					RequestedNode:     assignedEntry.RequestedNode,
					AssignedNode:      assignedEntry.AssignedNode,
					TimeUntilReassign: 0,
				}
			} else if receivedRequest.State == elevator.ACTIVE {
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
				thisNodeInfo = f_GetNodeInfo(nodeOps)
				globalQueue = f_GetGlobalQueue(nodeOps)
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

	nodeOperations := T_NodeOperations{ //Make global for jonas
		c_readNodeInfo:         make(chan chan T_NodeInfo),
		c_writeNodeInfo:        make(chan T_NodeInfo),
		c_readAndWriteNodeInfo: make(chan chan T_NodeInfo),

		c_readGlobalQueue:         make(chan chan []T_GlobalQueueEntry),
		c_writeGlobalQueue:        make(chan []T_GlobalQueueEntry),
		c_readAndWriteGlobalQueue: make(chan chan []T_GlobalQueueEntry),

		c_readConnectedNodes:         make(chan chan []T_NodeInfo),
		c_writeConnectedNodes:        make(chan []T_NodeInfo),
		c_readAndWriteConnectedNodes: make(chan chan []T_NodeInfo),
	}
	elevatorOperations := elevator.T_ElevatorOperations{
		C_readElevator:         make(chan chan elevator.T_Elevator),
		C_writeElevator:        make(chan elevator.T_Elevator),
		C_readAndWriteElevator: make(chan chan elevator.T_Elevator),
	}

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
	c_quitMasterRoutines := make(chan bool)
	c_nodeIsSlave := make(chan bool)
	c_quitSlaveRoutines := make(chan bool)
	c_ackSentGlobalQueueToSlave := make(chan T_AckObject)

	go func() {
		go f_NodeOperationManager(&ThisNode, nodeOperations, elevatorOperations) //SHOULD BE THE ONLY REFERENCE TO ThisNode!
		go f_ElevatorManager(nodeOperations, elevatorOperations, c_shouldCheckIfAssigned, c_entryFromElevator)
		go F_ReceiveSlaveMessage(c_receiveSlaveMessage, nodeOperations, SLAVEPORT)
		go F_ReceiveMasterMessage(c_receiveMasterMessage, nodeOperations, MASTERPORT)
		go F_TransmitSlaveMessage(c_transmitSlaveMessage, SLAVEPORT)
		go F_TransmitMasterMessage(c_transmitMasterMessage, MASTERPORT)
		for {
			select {
			case <-c_nodeIsMaster:
				c_quitMasterRoutines = make(chan bool)
				c_quitSlaveRoutines = make(chan bool)
				go f_MasterVariableWatchDog(nodeOperations, c_lastAssignedEntry, c_assignmentWasSucessFull, c_ackSentGlobalQueueToSlave, c_quitMasterRoutines)
				go f_MasterTimeManager(nodeOperations, c_quitMasterRoutines)
			case <-c_nodeIsSlave:
				c_quitMasterRoutines = make(chan bool)
				c_quitSlaveRoutines = make(chan bool)
				go f_SlaveVariableWatchDog(nodeOperations, c_quitSlaveRoutines)
				go f_SlaveTimeManager(nodeOperations, c_quitSlaveRoutines)
			default:
				time.Sleep(time.Duration(LEASTRESPONSIVEPERIOD) * time.Microsecond)
			}
		}
	}()

	sendTimer := time.NewTicker(time.Duration(SENDPERIOD) * time.Millisecond)
	printGQTimer := time.NewTicker(time.Duration(2000) * time.Millisecond) //Test function
	assignState := ASSIGN
	nodeRole := f_GetNodeInfo(nodeOperations).Role
	if nodeRole == MASTER {
		c_nodeIsMaster <- true
	} else {
		c_nodeIsSlave <- true
	}

	for {
		nodeRole = f_GetNodeInfo(nodeOperations).Role
		switch nodeRole {
		case MASTER:
			select {
			case masterMessage := <-c_receiveMasterMessage:
				f_WriteLogMasterMessage(nodeOperations, masterMessage)
				f_UpdateConnectedNodes(nodeOperations, masterMessage.Transmitter)
				f_WriteLogConnectedNodes(nodeOperations, f_GetConnectedNodes(nodeOperations))
				thisNode := f_GetNodeInfo(nodeOperations)
				if masterMessage.Transmitter.PRIORITY != thisNode.PRIORITY {
					f_UpdateGlobalQueueMaster(nodeOperations, masterMessage)
				}

			case slaveMessage := <-c_receiveSlaveMessage:

				f_WriteLogSlaveMessage(nodeOperations, slaveMessage)
				f_UpdateConnectedNodes(nodeOperations, slaveMessage.Transmitter)
				f_WriteLogConnectedNodes(nodeOperations, f_GetConnectedNodes(nodeOperations))
				if slaveMessage.Entry.Request.Calltype != elevator.NONECALL {
					f_AddEntryGlobalQueue(nodeOperations, slaveMessage.Entry)
				}

			case entryFromElevator := <-c_entryFromElevator:
				f_AddEntryGlobalQueue(nodeOperations, entryFromElevator)
				if entryFromElevator.Request.State == elevator.DONE {
					c_shouldCheckIfAssigned <- true
				}

				thisNode := f_GetNodeInfo(nodeOperations)
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
				transmitterNodeInfo := f_GetNodeInfo(nodeOperations)
				masterMessage := T_MasterMessage{
					Transmitter: transmitterNodeInfo,
					GlobalQueue: f_GetGlobalQueue(nodeOperations),
				}
				c_transmitMasterMessage <- masterMessage
				sendTimer.Reset(time.Duration(SENDPERIOD) * time.Millisecond)

			case <-printGQTimer.C:
				globalQueue := f_GetGlobalQueue(nodeOperations)
				nodeInfo := f_GetNodeInfo(nodeOperations)
				F_WriteLog("Node: | " + strconv.Itoa(int(nodeInfo.PRIORITY)) + " | MASTER | has GQ:\n")
				for _, entry := range globalQueue {
					F_WriteLogGlobalQueueEntry(entry)
				}
				printGQTimer.Reset(time.Duration(2000) * time.Millisecond)

			default:
				c_readNodeInfo := make(chan T_NodeInfo)
				c_writeNodeInfo := make(chan T_NodeInfo)
				c_quitGetSetNodeInfo := make(chan bool)
				connectedNodes := f_GetConnectedNodes(nodeOperations)

				go f_GetAndSetNodeInfo(nodeOperations, c_readNodeInfo, c_writeNodeInfo, c_quitGetSetNodeInfo)
				oldNodeInfo := <-c_readNodeInfo
				newNodeInfo := f_AssignNewRole(oldNodeInfo, connectedNodes)
				c_writeNodeInfo <- newNodeInfo
				c_quitGetSetNodeInfo <- true

				thisNodeInfo := newNodeInfo
				f_UpdateConnectedNodes(nodeOperations, thisNodeInfo) //Update connected nodes with newnodeinfo

				//Need to be in own FSM
				switch assignState {
				case ASSIGN:
					c_readGlobalQueue := make(chan []T_GlobalQueueEntry)
					c_writeGlobalQueue := make(chan []T_GlobalQueueEntry)
					c_quitGetSetGlobalQueue := make(chan bool)
					connectedNodes := f_GetConnectedNodes(nodeOperations)
					avalibaleNodes := f_GetAvalibaleNodes(connectedNodes)
					go f_GetAndSetGlobalQueue(nodeOperations, c_readGlobalQueue, c_writeGlobalQueue, c_quitGetSetGlobalQueue)
					globalQueue := <-c_readGlobalQueue

					assignedEntry, assignedEntryIndex := F_AssignNewEntry(globalQueue, connectedNodes, avalibaleNodes)
					if (assignedEntry != T_GlobalQueueEntry{}) {
						globalQueue[assignedEntryIndex] = assignedEntry
						c_lastAssignedEntry <- assignedEntry
						F_WriteLog("Assigned request with ID: " + strconv.Itoa(int(assignedEntry.Request.Id)) + " assigned to node " + strconv.Itoa(int(assignedEntry.AssignedNode)))
						assignState = WAITFORACK
					}
					c_writeGlobalQueue <- globalQueue
					c_quitGetSetGlobalQueue <- true

				case WAITFORACK:
					select {
					case assigmentWasSucessFull := <-c_assignmentWasSucessFull:
						if assigmentWasSucessFull {
							assignState = ASSIGN
						}
					default:
					}
				}

				if newNodeInfo.Role == SLAVE {
					c_nodeIsSlave <- true
					close(c_quitMasterRoutines)
					assignState = ASSIGN
				}

			}

		case SLAVE:
			select {
			case masterMessage := <-c_receiveMasterMessage:
				f_WriteLogMasterMessage(nodeOperations, masterMessage)
				f_UpdateGlobalQueueSlave(nodeOperations, masterMessage)
				f_UpdateConnectedNodes(nodeOperations, masterMessage.Transmitter)

			case slaveMessage := <-c_receiveSlaveMessage:
				f_WriteLogSlaveMessage(nodeOperations, slaveMessage)
				f_UpdateConnectedNodes(nodeOperations, slaveMessage.Transmitter)

			case entryFromElevator := <-c_entryFromElevator:
				f_AddEntryGlobalQueue(nodeOperations, entryFromElevator)
				if entryFromElevator.Request.State == elevator.DONE {
					c_shouldCheckIfAssigned <- true
				}
				thisNode := f_GetNodeInfo(nodeOperations)
				F_WriteLog("Node: | " + strconv.Itoa(int(thisNode.PRIORITY)) + " | SLAVE | updated GQ entry:\n")
				F_WriteLogGlobalQueueEntry(entryFromElevator)

				transmitter := f_GetNodeInfo(nodeOperations)
				infoMessage := T_SlaveMessage{
					Transmitter: transmitter,
					Entry:       entryFromElevator,
				}
				c_transmitSlaveMessage <- infoMessage
			case <-sendTimer.C:
				transmitter := f_GetNodeInfo(nodeOperations)
				aliveMessage := T_SlaveMessage{
					Transmitter: transmitter,
					Entry:       T_GlobalQueueEntry{},
				}
				c_transmitSlaveMessage <- aliveMessage
				sendTimer.Reset(time.Duration(SENDPERIOD) * time.Millisecond)
			case <-printGQTimer.C:
				globalQueue := f_GetGlobalQueue(nodeOperations)
				nodeInfo := f_GetNodeInfo(nodeOperations)
				F_WriteLog("Node: | " + strconv.Itoa(int(nodeInfo.PRIORITY)) + " | SLAVE | has GQ:\n")
				for _, entry := range globalQueue {
					F_WriteLogGlobalQueueEntry(entry)
				}
				printGQTimer.Reset(time.Duration(2000) * time.Millisecond)
			default:
				c_readNodeInfo := make(chan T_NodeInfo)
				c_writeNodeInfo := make(chan T_NodeInfo)
				c_quitGetSetNodeInfo := make(chan bool)

				connectedNodes := f_GetConnectedNodes(nodeOperations)
				go f_GetAndSetNodeInfo(nodeOperations, c_readNodeInfo, c_writeNodeInfo, c_quitGetSetNodeInfo)
				oldNodeInfo := <-c_readNodeInfo
				newNodeInfo := f_AssignNewRole(oldNodeInfo, connectedNodes)
				c_writeNodeInfo <- newNodeInfo
				c_quitGetSetNodeInfo <- true

				thisNodeInfo := newNodeInfo
				f_UpdateConnectedNodes(nodeOperations, thisNodeInfo)

				if newNodeInfo.Role == MASTER {
					c_nodeIsMaster <- true
					close(c_quitSlaveRoutines)
				}
			}
		}
		time.Sleep(time.Duration(MOSTRESPONSIVEPERIOD) * time.Microsecond)
	}
}
