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

func f_simulateRequest(nodeOps T_NodeOperations, elevatorOps elevator.T_ElevatorOperations, requestFromElevator chan elevator.T_Request, requestToElevator chan elevator.T_Request) {
	increment := uint16(0)
	go func() {
		for {
			select {
			case request := <-requestToElevator:
				c_readElevator0 := make(chan elevator.T_Elevator)
				c_writeElevator0 := make(chan elevator.T_Elevator)
				c_quitGetSetElevator0 := make(chan bool)
				go elevator.F_GetAndSetElevator(elevatorOps, c_readElevator0, c_writeElevator0, c_quitGetSetElevator0)
				currentElevator0 := <-c_readElevator0
				(*currentElevator0.P_info).State = elevator.MOVING
				c_writeElevator0 <- currentElevator0
				c_quitGetSetElevator0 <- true

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
				requestFromElevator <- request
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
			requestFromElevator <- returnRequest
		} else if parts[0] == "H" {
			returnRequest = elevator.T_Request{
				Id:        increment,
				State:     elevator.UNASSIGNED,
				Calltype:  elevator.HALL,
				Floor:     int8(floor),
				Direction: elevator.UP,
			}
			increment += 1
			requestFromElevator <- returnRequest
		}
	}
}

// ***	END TEST FUNCTIONS	***//

func f_InitNode(config T_Config) T_Node {
	thisElevatorInfo := elevator.T_ElevatorInfo{
		Direction: elevator.NONE,
		Floor:     1,
		State:     elevator.IDLE,
	}
	thisNodeInfo := T_NodeInfo{
		PRIORITY:     config.Priority,
		ElevatorInfo: thisElevatorInfo,
		Role:         MASTER,
	}
	var c_receiveRequest chan elevator.T_Request
	var c_distributeRequest chan elevator.T_Request
	var c_distributeInfo chan elevator.T_ElevatorInfo

	thisElevator := elevator.T_Elevator{
		P_info:              &thisElevatorInfo,
		C_receiveRequest:    c_receiveRequest,
		C_distributeRequest: c_distributeRequest,
		C_distributeInfo:    c_distributeInfo,
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
}

func F_WriteLog(text string) bool {
	logFile, _ := os.OpenFile("log/debug1.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Print(text)
	return true
}
func nodeRoleToString(role T_NodeRole) string {
	switch role {
	case MASTER:
		return "MASTER"
	default:
		return "SLAVE"
	}
}
func f_WriteLogConnectedNodes(ops T_NodeOperations, connectedNodes []T_NodeInfo) {
	thisNode := f_GetNodeInfo(ops)
	logStr := fmt.Sprintf("Node: | %d | %s | has connected nodes | ", thisNode.PRIORITY, nodeRoleToString(thisNode.Role))
	for _, info := range connectedNodes {
		logStr += fmt.Sprintf("%d (Role: %s, ElevatorInfo: %+v, TimeUntilDisconnect: %d) | ",
			info.PRIORITY, nodeRoleToString(info.Role), info.ElevatorInfo, info.TimeUntilDisconnect)
	}
	F_WriteLog(logStr)
}
func F_WriteLogGlobalQueueEntry(entry T_GlobalQueueEntry) {
	logStr := fmt.Sprintf("Entry: | %d | Calltype: %s | Floor: %d | Direction: %s | Reassigned in: %.2f | ",
		entry.Request.Id, callTypeToString(entry.Request.Calltype), entry.Request.Floor, directionToString(entry.Request.Direction), float64(entry.TimeUntilReassign))
	logStr += fmt.Sprintf("Requested node: | %d | ",
		entry.RequestedNode)
	logStr += fmt.Sprintf("Assigned node: | %d | ",
		entry.AssignedNode)
	F_WriteLog(logStr)
}
func callTypeToString(callType elevator.T_Call) string {
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

func directionToString(direction elevator.T_ElevatorDirection) string {
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
	logStr := fmt.Sprintf("Node: | %d | %s | received SM from | %d | Request: | %d | Calltype: %s | Floor: %d | Direction: %s |",
		thisNode.PRIORITY, nodeRoleToString(thisNode.Role), slaveMessage.Transmitter.PRIORITY, request.Id, callTypeToString(request.Calltype), request.Floor, directionToString(request.Direction))
	F_WriteLog(logStr)
}
func f_WriteLogMasterMessage(ops T_NodeOperations, masterMessage T_MasterMessage) {
	thisNode := f_GetNodeInfo(ops)
	roleStr := nodeRoleToString(thisNode.Role)
	transmitterRoleStr := nodeRoleToString(masterMessage.Transmitter.Role)

	logStr := fmt.Sprintf("Node: | %d | %s | received MM from | %d | %s | GlobalQueue: [",
		thisNode.PRIORITY, roleStr, masterMessage.Transmitter.PRIORITY, transmitterRoleStr)

	// Iterate over the GlobalQueue to add details for each entry
	for i, entry := range masterMessage.GlobalQueue {
		entryStr := fmt.Sprintf("Request: | %d | Calltype: %s | Floor: %d | Direction: %s | Reassigned in: %d | Requested node: | %d | Assigned node: | %d |",
			entry.Request.Id, callTypeToString(entry.Request.Calltype), entry.Request.Floor,
			directionToString(entry.Request.Direction), entry.TimeUntilReassign,
			entry.RequestedNode, entry.AssignedNode)

		// Append this entry's details to the log string
		logStr += entryStr
		if i < len(masterMessage.GlobalQueue)-1 {
			logStr += ", " // Add a comma separator between entries, except after the last one
		}
	}
	logStr += "]" // Close the GlobalQueue information

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
func f_RemoveNode(nodes []T_NodeInfo, nodeToRemove T_NodeInfo) []T_NodeInfo {
	for i, nodeInfo := range nodes {
		if nodeInfo.PRIORITY == nodeToRemove.PRIORITY {
			return append(nodes[:i], nodes[i+1:]...)
		}
	}
	return nodes
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
func f_MasterVariableWatchDog(ops T_NodeOperations, c_lastAssignedNode chan T_NodeInfo, c_assignmentSuccessfull chan bool, c_quit chan bool) {
	go f_SlaveVariableWatchDog(ops, c_quit)
	go func() {
		for {
		PollLastAssigned:
			select {
			case lastAssignedNode := <-c_lastAssignedNode:
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
						updatedAssignedNode := f_FindNodeInfo(lastAssignedNode.PRIORITY, connectedNodes)
						if updatedAssignedNode.ElevatorInfo.State != elevator.IDLE {
							c_assignmentSuccessfull <- true
							F_WriteLog("Found ack")
							assignBreakoutTimer.Stop()
							break PollLastAssigned
						}
					}
				}
			}
		}
	}()

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
			if (servicedEntry != T_GlobalQueueEntry{}) {
				F_WriteLog("Removed entry: | " + strconv.Itoa(int(servicedEntry.Request.Id)) + " | " + strconv.Itoa(int(servicedEntry.RequestedNode)) + " | from global queue")
				globalQueue = append(globalQueue[:servicedEntryIdex], globalQueue[servicedEntryIdex+1:]...)
			}
			c_writeGlobalQueue <- globalQueue
			c_quitGetSetGlobalQueue <- true
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
				if entry.TimeUntilReassign > 0 && entry.AssignedNode != 0 {
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
		entryToAdd.AssignedNode = 0
		//entryToAdd.TimeUntilReassign = REASSIGNTIME //Should be set by slave
		globalQueue = append(globalQueue, entryToAdd)
		c_writeGlobalQueue <- globalQueue
	} else { //should update the existing entry
		globalQueue[entryIndex] = entryToAdd
		c_writeGlobalQueue <- globalQueue
	}
	c_quit <- true
}
func f_ElevatorManager(nodeOps T_NodeOperations, elevatorOps elevator.T_ElevatorOperations, c_entryFromElevator chan T_GlobalQueueEntry) {
	c_requestFromElevator := make(chan elevator.T_Request)
	c_requestToElevator := make(chan elevator.T_Request)

	//go elevator.F_RunElevator(elevatorOps, c_requestFromElevator, c_requestToElevator)
	go f_simulateRequest(nodeOps, elevatorOps, c_requestFromElevator, c_requestToElevator)

	go func() {
		for {
			thisNodeInfo := f_GetNodeInfo(nodeOps)
			c_readGlobalQueue := make(chan []T_GlobalQueueEntry)
			c_writeGlobalQueue := make(chan []T_GlobalQueueEntry)
			c_quit := make(chan bool)
			go f_GetAndSetGlobalQueue(nodeOps, c_readGlobalQueue, c_writeGlobalQueue, c_quit)

			globalQueue := <-c_readGlobalQueue
			assignedEntry, assignedEntryIndex := F_FindAssignedEntry(globalQueue, thisNodeInfo)
			if (assignedEntry != T_GlobalQueueEntry{}) {
				assignedEntry.Request.State = elevator.ACTIVE //Cannot be done by this, ACTIVE has to be set in GQ when the node is actually in MOVING
				globalQueue[assignedEntryIndex] = assignedEntry
				c_requestToElevator <- assignedEntry.Request //NB! Depending on that elevator is polling in IDLE
			}
			c_writeGlobalQueue <- globalQueue
			c_quit <- true
		}
	}()

	for {
		select {
		case receivedRequest := <-c_requestFromElevator:
			thisNodeInfo := f_GetNodeInfo(nodeOps)
			entry := T_GlobalQueueEntry{}
			if receivedRequest.State == elevator.DONE {
				entry = T_GlobalQueueEntry{
					Request:           receivedRequest,
					RequestedNode:     thisNodeInfo.PRIORITY,
					AssignedNode:      thisNodeInfo.PRIORITY,
					TimeUntilReassign: 0,
				}
			} else {
				entry = T_GlobalQueueEntry{
					Request:           receivedRequest,
					RequestedNode:     thisNodeInfo.PRIORITY,
					AssignedNode:      0,
					TimeUntilReassign: REASSIGNTIME,
				}
			}
			c_entryFromElevator <- entry
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
	c_lastAssignedNode := make(chan T_NodeInfo)
	c_assignmentWasSucessFull := make(chan bool)
	c_nodeIsMaster := make(chan bool)
	c_quitMasterRoutines := make(chan bool)
	c_nodeIsSlave := make(chan bool)
	c_quitSlaveRoutines := make(chan bool)

	go func() {
		go F_ReceiveSlaveMessage(c_receiveSlaveMessage, nodeOperations, SLAVEPORT)
		go F_ReceiveMasterMessage(c_receiveMasterMessage, nodeOperations, MASTERPORT)
		go F_TransmitSlaveMessage(c_transmitSlaveMessage, SLAVEPORT)
		go F_TransmitMasterMessage(c_transmitMasterMessage, MASTERPORT)
		for {
			select {
			case <-c_nodeIsMaster:
				close(c_quitSlaveRoutines)
				c_quitMasterRoutines = make(chan bool)
				c_quitSlaveRoutines = make(chan bool)
				go f_MasterVariableWatchDog(nodeOperations, c_lastAssignedNode, c_assignmentWasSucessFull, c_quitMasterRoutines)
				go f_MasterTimeManager(nodeOperations, c_quitMasterRoutines)
			case <-c_nodeIsSlave:
				close(c_quitMasterRoutines)
				c_quitMasterRoutines = make(chan bool)
				c_quitSlaveRoutines = make(chan bool)
				go f_SlaveVariableWatchDog(nodeOperations, c_quitSlaveRoutines)
				go f_SlaveTimeManager(nodeOperations, c_quitSlaveRoutines)
			}
		}
	}()

	go f_NodeOperationManager(&ThisNode, nodeOperations, elevatorOperations) //SHOULD BE THE ONLY REFERENCE TO ThisNode!
	go f_ElevatorManager(nodeOperations, elevatorOperations, c_entryFromElevator)
	sendTimer := time.NewTicker(time.Duration(SENDPERIOD) * time.Millisecond)
	printGQTimer := time.NewTicker(time.Duration(2000) * time.Millisecond) //Test function
	assignState := ASSIGN
	nodeRole := f_GetNodeInfo(nodeOperations).Role
	if nodeRole == MASTER {
		c_nodeIsMaster <- true
	} else {
		c_nodeIsSlave <- true
	}
	ackinc := 0
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
					for _, remoteEntry := range masterMessage.GlobalQueue {
						f_AddEntryGlobalQueue(nodeOperations, remoteEntry)
					}
				}
				//IMPORTANT: cannot really propagate to slave until it knows that the other master has received its GQ

			case slaveMessage := <-c_receiveSlaveMessage:
				f_WriteLogSlaveMessage(nodeOperations, slaveMessage)
				f_UpdateConnectedNodes(nodeOperations, slaveMessage.Transmitter)
				f_WriteLogConnectedNodes(nodeOperations, f_GetConnectedNodes(nodeOperations))
				if slaveMessage.Entry.Request.Calltype != elevator.NONECALL {
					f_AddEntryGlobalQueue(nodeOperations, slaveMessage.Entry)
				}

			case entryFromElevator := <-c_entryFromElevator:
				f_AddEntryGlobalQueue(nodeOperations, entryFromElevator)

				thisNode := f_GetNodeInfo(nodeOperations)
				F_WriteLog("Node: | " + strconv.Itoa(int(thisNode.PRIORITY)) + " | MASTER | updated GQ entry:\n")
				F_WriteLogGlobalQueueEntry(entryFromElevator)

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
						c_lastAssignedNode <- f_FindNodeInfo(assignedEntry.AssignedNode, connectedNodes)
						F_WriteLog("Assigned request with ID: " + strconv.Itoa(int(assignedEntry.Request.Id)) + " assigned to node " + strconv.Itoa(int(assignedEntry.AssignedNode)))
						assignState = WAITFORACK
					}
					c_writeGlobalQueue <- globalQueue
					c_quitGetSetGlobalQueue <- true

				case WAITFORACK:
					select {
					case assigmentWasSucessFull := <-c_assignmentWasSucessFull:
						ackinc = 0
						if assigmentWasSucessFull {
							assignState = ASSIGN
						} else {
							//assignState = ASSIGN
							//An entry, assigned but not resent and confirmed will be reassigned by VairableWatchdog
							//However, if it never leaves IDLE, but somehow is not ready (i.e. local elevator is not in idle)
							//it can lead to deadlock. Unsure of what to do
							//The connectiontime is less than the breakouttime, meaning we will disconnect if elevator.state is not updated
						}
					default:
						ackinc++
						if ackinc < 10 {
							fmt.Println("Waiting for ACK...")
						}
					}
				}

				if newNodeInfo.Role == SLAVE {
					c_nodeIsSlave <- true
					assignState = ASSIGN
					fmt.Println("Node " + strconv.Itoa(int(newNodeInfo.PRIORITY)) + "entered SLAVE mode")
				}
			}

		case SLAVE:
			select {
			case masterMessage := <-c_receiveMasterMessage:
				f_WriteLogMasterMessage(nodeOperations, masterMessage)
				for _, remoteEntry := range masterMessage.GlobalQueue {
					f_AddEntryGlobalQueue(nodeOperations, remoteEntry)
				}
				f_UpdateConnectedNodes(nodeOperations, masterMessage.Transmitter)
			case slaveMessage := <-c_receiveSlaveMessage:
				f_WriteLogSlaveMessage(nodeOperations, slaveMessage)
				f_UpdateConnectedNodes(nodeOperations, slaveMessage.Transmitter)

			case entryFromElevator := <-c_entryFromElevator:
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
					fmt.Println("Node " + strconv.Itoa(int(newNodeInfo.PRIORITY)) + "entered MASTER mode")
				}
				/*
					KNOWN BUG
					In the situation:
					Node: | 1 | MASTER | updated GQ entry:
					2024/03/03 03:48:06 run_node.go:139: Entry: | 1 | Calltype: CAB | Floor: 2 | Direction: UP | Reassigned in: 15.00 | Requested node: | 1 | Assigned node: | 0 |
					2024/03/03 03:48:06 run_node.go:139: Assigned request with ID: 1 assigned to node 2
					2024/03/03 03:48:06 run_node.go:139: Getting ack from last assinged...
					2024/03/03 03:48:06 run_node.go:139: Node: | 1 | MASTER | has GQ:
					2024/03/03 03:48:06 run_node.go:139: Entry: | 0 | Calltype: CAB | Floor: 1 | Direction: UP | Reassigned in: 12.00 | Requested node: | 1 | Assigned node: | 1 |
					2024/03/03 03:48:06 run_node.go:139: Entry: | 1 | Calltype: CAB | Floor: 2 | Direction: UP | Reassigned in: 15.00 | Requested node: | 1 | Assigned node: | 2 |
					2024/03/03 03:48:06 run_node.go:139: Node: | 2 | SLAVE | received MM from | 1 | MASTER | GlobalQueue: [Request: | 0 | Calltype: CAB | Floor: 1 | Direction: UP | Reassigned in: 11 | Requested node: | 1 | Assigned node: | 1 |, Request: | 1 | Calltype: CAB | Floor: 2 | Direction: UP | Reassigned in: 14 | Requested node: | 1 | Assigned node: | 2 |]
					2024/03/03 03:48:06 run_node.go:139: Node: | 1 | MASTER | received MM from | 1 | MASTER | GlobalQueue: [Request: | 0 | Calltype: CAB | Floor: 1 | Direction: UP | Reassigned in: 11 | Requested node: | 1 | Assigned node: | 1 |, Request: | 1 | Calltype: CAB | Floor: 2 | Direction: UP | Reassigned in: 14 | Requested node: | 1 | Assigned node: | 2 |]
					2024/03/03 03:48:06 run_node.go:139: Node: | 1 | MASTER | has connected nodes | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) | 2 (Role: SLAVE, ElevatorInfo: {Direction:0 Floor:1 State:0}, TimeUntilDisconnect: 3) |
					2024/03/03 03:48:07 run_node.go:139: Node: | 2 | SLAVE | received SM from | 2 | Request: | 0 | Calltype: NONE | Floor: 0 | Direction: NONE |
					2024/03/03 03:48:07 run_node.go:139: Node: | 1 | MASTER | received SM from | 2 | Request: | 0 | Calltype: NONE | Floor: 0 | Direction: NONE |
					2024/03/03 03:48:07 run_node.go:139: Node: | 1 | MASTER | has connected nodes | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) | 2 (Role: SLAVE, ElevatorInfo: {Direction:0 Floor:1 State:0}, TimeUntilDisconnect: 4) |
					2024/03/03 03:48:07 run_node.go:139: Node: | 1 | MASTER | received MM from | 1 | MASTER | GlobalQueue: [Request: | 0 | Calltype: CAB | Floor: 1 | Direction: UP | Reassigned in: 10 | Requested node: | 1 | Assigned node: | 1 |, Request: | 1 | Calltype: CAB | Floor: 2 | Direction: UP | Reassigned in: 13 | Requested node: | 1 | Assigned node: | 2 |]
					2024/03/03 03:48:07 run_node.go:139: Node: | 2 | SLAVE | received MM from | 1 | MASTER | GlobalQueue: [Request: | 0 | Calltype: CAB | Floor: 1 | Direction: UP | Reassigned in: 10 | Requested node: | 1 | Assigned node: | 1 |, Request: | 1 | Calltype: CAB | Floor: 2 | Direction: UP | Reassigned in: 13 | Requested node: | 1 | Assigned node: | 2 |]
					2024/03/03 03:48:07 run_node.go:139: Node: | 1 | MASTER | has connected nodes | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) | 2 (Role: SLAVE, ElevatorInfo: {Direction:0 Floor:1 State:0}, TimeUntilDisconnect: 3) |
					2024/03/03 03:48:07 run_node.go:139: Found assigned request with ID: 1 assigned to node 1
					2024/03/03 03:48:08 run_node.go:139: Node: | 2 | SLAVE | received SM from | 2 | Request: | 0 | Calltype: NONE | Floor: 0 | Direction: NONE |
					2024/03/03 03:48:08 run_node.go:139: Node: | 1 | MASTER | received SM from | 2 | Request: | 0 | Calltype: NONE | Floor: 0 | Direction: NONE |
					2024/03/03 03:48:08 run_node.go:139: Node: | 1 | MASTER | has connected nodes | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) | 2 (Role: SLAVE, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) |
					2024/03/03 03:48:08 run_node.go:139: Found ack
					2024/03/03 03:48:08 run_node.go:139: Node: | 1 | MASTER | has GQ:
					2024/03/03 03:48:08 run_node.go:139: Entry: | 0 | Calltype: CAB | Floor: 1 | Direction: UP | Reassigned in: 10.00 | Requested node: | 1 | Assigned node: | 1 |
					2024/03/03 03:48:08 run_node.go:139: Entry: | 1 | Calltype: CAB | Floor: 2 | Direction: UP | Reassigned in: 13.00 | Requested node: | 1 | Assigned node: | 2 |
					2024/03/03 03:48:08 run_node.go:139: Node: | 2 | SLAVE | received MM from | 1 | MASTER | GlobalQueue: [Request: | 0 | Calltype: CAB | Floor: 1 | Direction: UP | Reassigned in: 9 | Requested node: | 1 | Assigned node: | 1 |, Request: | 1 | Calltype: CAB | Floor: 2 | Direction: UP | Reassigned in: 12 | Requested node: | 1 | Assigned node: | 2 |]
					2024/03/03 03:48:08 run_node.go:139: Node: | 1 | MASTER | received MM from | 1 | MASTER | GlobalQueue: [Request: | 0 | Calltype: CAB | Floor: 1 | Direction: UP | Reassigned in: 9 | Requested node: | 1 | Assigned node: | 1 |, Request: | 1 | Calltype: CAB | Floor: 2 | Direction: UP | Reassigned in: 12 | Requested node: | 1 | Assigned node: | 2 |]
					2024/03/03 03:48:08 run_node.go:139: Found assigned request with ID: 1 assigned to node 1
					2024/03/03 03:48:08 run_node.go:139: Node: | 1 | MASTER | has connected nodes | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) | 2 (Role: SLAVE, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 3) |
					2024/03/03 03:48:09 run_node.go:139: Node: | 1 | MASTER | received MM from | 1 | MASTER | GlobalQueue: [Request: | 0 | Calltype: CAB | Floor: 1 | Direction: UP | Reassigned in: 8 | Requested node: | 1 | Assigned node: | 1 |, Request: | 1 | Calltype: CAB | Floor: 2 | Direction: UP | Reassigned in: 11 | Requested node: | 1 | Assigned node: | 2 |]
					2024/03/03 03:48:09 run_node.go:139: Node: | 1 | MASTER | has connected nodes | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) | 2 (Role: SLAVE, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 2) |
					2024/03/03 03:48:10 run_node.go:139: Node: | 1 | MASTER | has GQ:
					2024/03/03 03:48:10 run_node.go:139: Entry: | 0 | Calltype: CAB | Floor: 1 | Direction: UP | Reassigned in: 8.00 | Requested node: | 1 | Assigned node: | 1 |
					2024/03/03 03:48:10 run_node.go:139: Entry: | 1 | Calltype: CAB | Floor: 2 | Direction: UP | Reassigned in: 11.00 | Requested node: | 1 | Assigned node: | 2 |
					2024/03/03 03:48:10 run_node.go:139: Node: | 1 | MASTER | received MM from | 1 | MASTER | GlobalQueue: [Request: | 0 | Calltype: CAB | Floor: 1 | Direction: UP | Reassigned in: 7 | Requested node: | 1 | Assigned node: | 1 |, Request: | 1 | Calltype: CAB | Floor: 2 | Direction: UP | Reassigned in: 10 | Requested node: | 1 | Assigned node: | 2 |]
					2024/03/03 03:48:10 run_node.go:139: Node: | 1 | MASTER | has connected nodes | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) | 2 (Role: SLAVE, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 1) |
					2024/03/03 03:48:11 run_node.go:139: Node 2 disconnected
					2024/03/03 03:48:11 run_node.go:139: Node: | 1 | MASTER | received MM from | 1 | MASTER | GlobalQueue: [Request: | 0 | Calltype: CAB | Floor: 1 | Direction: UP | Reassigned in: 6 | Requested node: | 1 | Assigned node: | 1 |, Request: | 1 | Calltype: CAB | Floor: 2 | Direction: UP | Reassigned in: 9 | Requested node: | 1 | Assigned node: | 2 |]
					2024/03/03 03:48:11 run_node.go:139: Node: | 1 | MASTER | has connected nodes | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) |
					2024/03/03 03:48:12 run_node.go:139: Node: | 1 | MASTER | has GQ:
					2024/03/03 03:48:12 run_node.go:139: Entry: | 0 | Calltype: CAB | Floor: 1 | Direction: UP | Reassigned in: 6.00 | Requested node: | 1 | Assigned node: | 1 |
					2024/03/03 03:48:12 run_node.go:139: Entry: | 1 | Calltype: CAB | Floor: 2 | Direction: UP | Reassigned in: 9.00 | Requested node: | 1 | Assigned node: | 2 |
					2024/03/03 03:48:12 run_node.go:139: Node: | 1 | MASTER | received MM from | 1 | MASTER | GlobalQueue: [Request: | 0 | Calltype: CAB | Floor: 1 | Direction: UP | Reassigned in: 5 | Requested node: | 1 | Assigned node: | 1 |, Request: | 1 | Calltype: CAB | Floor: 2 | Direction: UP | Reassigned in: 8 | Requested node: | 1 | Assigned node: | 2 |]
					2024/03/03 03:48:12 run_node.go:139: Node: | 1 | MASTER | has connected nodes | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) |
					2024/03/03 03:48:13 run_node.go:139: Node: | 1 | MASTER | updated GQ entry:
					2024/03/03 03:48:13 run_node.go:139: Entry: | 0 | Calltype: CAB | Floor: 1 | Direction: UP | Reassigned in: 0.00 | Requested node: | 1 | Assigned node: | 1 |
					2024/03/03 03:48:13 run_node.go:139: Removed entry: | 0 | 1 | from global queue
					2024/03/03 03:48:13 run_node.go:139: Ended GetSet goroutine of GQ because of deadlock

					Happens because try to write the same request to node 2s elevator twise, witch causes the deadlock.
					Node 2 (slave) writes its own globalQueue in elevatorManager
					TODO: Find out why it gets assigned twise, its update to the GQ should hinder this.
					Conseptual thing: SlaveNode SHOLD NOT make changes to the GQ it receives from master
					Remove all SLAVES references to GQ() modifiaction?
				*/
			}
		}
	}
}
