package node

import (
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
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
		Floor:     1,
		State:     elevator.IDLE,
	}
	thisNodeInfo := T_NodeInfo{
		PRIORITY:     config.Priority,
		ElevatorInfo: thisElevatorInfo,
		MSRole:       MASTER,
		PBRole:       BACKUP,
	}
	var c_receiveRequest chan elevator.T_Request
	var c_distributeRequest chan elevator.T_Request

	thisElevator := elevator.T_Elevator{
		P_info:              &thisElevatorInfo,
		C_receiveRequest:    c_receiveRequest,
		C_distributeRequest: c_distributeRequest,
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
func f_NodeRoleToString(role T_MasterSlaveRole) string {
	switch role {
	case MASTER:
		return "MASTER"
	default:
		return "SLAVE"
	}
}
func f_WriteLogConnectedNodes(ops T_NodeOperations, connectedNodes []T_NodeInfo) {
	thisNode := f_GetNodeInfo(ops)
	logStr := fmt.Sprintf("Node: | %d | %s | has connected nodes | ", thisNode.PRIORITY, f_NodeRoleToString(thisNode.MSRole))
	for _, info := range connectedNodes {
		logStr += fmt.Sprintf("%d (Role: %s, ElevatorInfo: %+v, TimeUntilDisconnect: %d) | ",
			info.PRIORITY, f_NodeRoleToString(info.MSRole), info.ElevatorInfo, info.TimeUntilDisconnect)
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
		int(thisNode.PRIORITY), f_NodeRoleToString(thisNode.MSRole), int(slaveMessage.Transmitter.PRIORITY), int(request.Id), f_RequestStateToString(slaveMessage.Entry.Request.State), f_CallTypeToString(request.Calltype), int(request.Floor), f_DirectionToString(request.Direction))
	F_WriteLog(logStr)
}
func f_WriteLogMasterMessage(ops T_NodeOperations, masterMessage T_MasterMessage) {
	thisNode := f_GetNodeInfo(ops)
	roleStr := f_NodeRoleToString(thisNode.MSRole)
	transmitterRoleStr := f_NodeRoleToString(masterMessage.Transmitter.MSRole)

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
	var returnRole T_MasterSlaveRole = MASTER
	for _, remoteNodeInfo := range connectedNodes {
		if remoteNodeInfo.PRIORITY < thisNodeInfo.PRIORITY {
			returnRole = SLAVE
		}
	}
	newNodeInfo := T_NodeInfo{
		PRIORITY:            thisNodeInfo.PRIORITY,
		MSRole:              returnRole,
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

func f_CheckGlobalQueueEntryStatus(c_getSetGlobalQueueInterface chan T_GetSetGlobalQueueInterface, getSetGlobalQueueInterface T_GetSetGlobalQueueInterface, c_ackSentEntryToSlave chan T_AckObject, c_quit chan bool) {
	for {
		select {
		case <-c_quit:
			F_WriteLog("Closed: f_CheckGlobalQueueEntryStatus")
			return
		default:
			thisNodeInfo := f_GetNodeInfo()
			c_getSetGlobalQueueInterface <- getSetGlobalQueueInterface
			globalQueue := <-getSetGlobalQueueInterface.c_get

			doneEntry, doneEntryIndex := T_GlobalQueueEntry{}, 0
			for i, entry := range globalQueue {
				if entry.TimeUntilReassign == 0 {
					doneEntry = globalQueue[i]
					//fmt.Println(strconv.Itoa(int(doneEntry.RequestedNode)))
					doneEntryIndex = i
				}
			}
			if (doneEntry.Request.State != elevator.DONE && doneEntry != T_GlobalQueueEntry{}) {
				globalQueue = f_ReassignUnfinishedEntry(globalQueue, doneEntry, doneEntryIndex)
			}
			getSetGlobalQueueInterface.c_set <- globalQueue

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

func f_CheckConnectedNodesStatus(c_getSetConnectedNodesInterface chan T_GetSetConnectedNodesInterface, getSetConnectedNodesInterface T_GetSetConnectedNodesInterface) { //begge
	for {
		select {
		default:
			c_getSetConnectedNodesInterface <- getSetConnectedNodesInterface
			connectedNodes := <-getSetConnectedNodesInterface.c_get

			nodeToDisconnect, nodeToDisconnectIndex := T_NodeInfo{}, 0
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

func f_DecrementTimeUntilReassign(c_getSetGlobalQueueInterface chan T_GetSetGlobalQueueInterface, getSetGlobalQueueInterface T_GetSetGlobalQueueInterface, c_quit chan bool) { //Master
	for {
		select {
		case <-c_quit:
			F_WriteLog("Closed: f_DecrementTimeUntilReassign")
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

func f_DecrementTimeUntilDisconnect(c_getSetConnectedNodesInterface chan T_GetSetConnectedNodesInterface, getSetConnectedNodesInterface T_GetSetConnectedNodesInterface) { //Begge
	for {
		select {
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

func f_CheckIfShouldAssign(c_getSetGlobalQueueInterface chan T_GetSetGlobalQueueInterface, getSetGlobalQueueInterface T_GetSetGlobalQueueInterface, c_ackAssignmentSucessFull chan T_AckObject, c_assignState chan T_AssignState, c_quit chan bool) {
	assignState := ASSIGN
	c_assignmentSuccessfull := make(chan bool)
	ackAssignmentSucessFull := T_AckObject{
		C_Acknowledgement: c_assignmentSuccessfull,
	}
	for {
		select {
		case assignState = <-c_assignState:
		case <-c_quit:
			F_WriteLog("Closed: f_CheckIfShouldAssign")
			return
		default:
			switch assignState {
			case ASSIGN:
				connectedNodes := f_GetConnectedNodes()
				avalibaleNodes := f_GetAvalibaleNodes(connectedNodes)
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
	} else {
		if entryToAdd.Request.State >= globalQueue[entryIndex].Request.State || entryToAdd.TimeUntilReassign <= globalQueue[entryIndex].TimeUntilReassign { //only allow forward entry states //>=?
			globalQueue[entryIndex] = entryToAdd
		} else {
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

func f_ElevatorManager(c_shouldCheckIfAssigned chan bool, c_entryFromElevator chan T_GlobalQueueEntry) {
	c_requestFromElevator := make(chan elevator.T_Request)
	c_requestToElevator := make(chan elevator.T_Request)
	shouldCheckIfAssigned := true

	//go elevator.F_RunElevator(elevatorOps, c_requestFromElevator, c_requestToElevator, ELEVATORPORT,ELEVATORPORT)
	go elevator.F_SimulateRequest(elevatorOperations, c_requestFromElevator, c_requestToElevator)

	thisNodeInfo := f_GetNodeInfo()
	globalQueue := f_GetGlobalQueue()
	assignedEntry, _ := F_FindAssignedEntry(globalQueue, thisNodeInfo)
	for {
		select {
		case requestFromElevator := <-c_requestFromElevator:
			entryFromElevator := F_AssembleEntryFromRequest(requestFromElevator, thisNodeInfo, assignedEntry)
			c_entryFromElevator <- entryFromElevator
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

// START DEV
func F_ProcessPairManager() {
	//has init
	nodeOperations := T_NodeOperations{
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
	getSetGlobalQueueInterface := T_GetSetGlobalQueueInterface{
		c_get: make(chan []T_GlobalQueueEntry),
		c_set: make(chan []T_GlobalQueueEntry),
	}
	getSetConnectedNodesInterface := T_GetSetConnectedNodesInterface{
		c_get: make(chan []T_NodeInfo),
		c_set: make(chan []T_NodeInfo),
	}

	go f_NodeOperationManager(&ThisNode, nodeOperations, elevatorOperations) //SHOULD BE THE ONLY REFERENCE TO ThisNode!

	c_isPrimary := make(chan bool)
	go F_RunBackup(nodeOperations, c_isPrimary)
	select {
	case <-c_isPrimary:
		err := exec.Command("gnome-terminal", "--", "go", "run", "main.go").Run()
		if err != nil {
			fmt.Println("Error starting BACKUP")
		}
		F_RunPrimary(nodeOperations, elevatorOperations)
	}
}

//IMPORTANT:
//-global variables should ALWAYS be handled by server to operate onn good data
//-all receive from channles should be organized in for-select!!! -> walk trough code and do

func F_RunBackup(nodeOps T_NodeOperations, c_isPrimary chan bool) {
	//constantly check if we receive messages
	F_WriteLog("Started as BACKUP")
	c_quitBackupRoutines := make(chan bool)
	c_receiveSlaveMessage := make(chan T_SlaveMessage)
	c_receiveMasterMessage := make(chan T_MasterMessage)

	go F_ReceiveSlaveMessage(c_receiveSlaveMessage, nodeOps, MASTERPORT, c_quitBackupRoutines)
	go F_ReceiveMasterMessage(c_receiveMasterMessage, nodeOps, SLAVEPORT, c_quitBackupRoutines)

	PBTimer := time.NewTicker(time.Duration(CONNECTIONTIME) * time.Second)
	for {
		select {
		case <-PBTimer.C:
			F_WriteLog("Switched to PRIMARY")
			c_isPrimary <- true
			close(c_quitBackupRoutines)
			return
		case masterMessage := <-c_receiveMasterMessage:
			thisNodeInfo := f_GetNodeInfo(nodeOps)
			f_SetGlobalQueue(nodeOps, masterMessage.GlobalQueue)
			if thisNodeInfo.PRIORITY == masterMessage.Transmitter.PRIORITY && thisNodeInfo.MSRole == MASTER {
				f_SetNodeInfo(nodeOps, masterMessage.Transmitter)
				PBTimer.Reset(time.Duration(CONNECTIONTIME) * time.Millisecond)
			}
		case slaveMessage := <-c_receiveSlaveMessage:
			thisNodeInfo := f_GetNodeInfo(nodeOps)
			if thisNodeInfo.PRIORITY == slaveMessage.Transmitter.PRIORITY && thisNodeInfo.MSRole == SLAVE {
				f_SetNodeInfo(nodeOps, slaveMessage.Transmitter)
				PBTimer.Reset(time.Duration(CONNECTIONTIME) * time.Millisecond)
			}
		}
	}
}

func F_RunPrimary(nodeOperations T_NodeOperations, elevatorOperations elevator.T_ElevatorOperations) {

	//END DEV

	c_getSetNodeInfoInterface := make(chan T_GetSetNodeInfoInterface)
	c_getSetGlobalQueueInterface := make(chan T_GetSetGlobalQueueInterface)
	c_getSetConnectedNodesInterface := make(chan T_GetSetConnectedNodesInterface)

	//to run the main FSM
	c_nodeIsMaster := make(chan bool)
	c_quitMasterRoutines := make(chan bool)
	c_nodeIsSlave := make(chan bool)

	c_receiveSlaveMessage := make(chan T_SlaveMessage)
	c_receiveMasterMessage := make(chan T_MasterMessage)
	c_quitPrimaryReceive := make(chan bool)
	c_transmitMasterMessage := make(chan T_MasterMessage)
	c_transmitSlaveMessage := make(chan T_SlaveMessage)

	c_entryFromElevator := make(chan T_GlobalQueueEntry)
	c_shouldCheckIfAssigned := make(chan bool)

	c_assignState := make(chan T_AssignState)
	c_ackAssignmentSucessFull := make(chan T_AckObject)
	c_ackSentGlobalQueueToSlave := make(chan T_AckObject)

	go func() {
		go f_ElevatorManager(nodeOperations, elevatorOperations, c_shouldCheckIfAssigned, c_entryFromElevator)
		go F_ReceiveSlaveMessage(c_receiveSlaveMessage, nodeOperations, SLAVEPORT, c_quitPrimaryReceive)
		go F_ReceiveMasterMessage(c_receiveMasterMessage, nodeOperations, MASTERPORT, c_quitPrimaryReceive)
		go F_TransmitSlaveMessage(c_transmitSlaveMessage, SLAVEPORT)
		go F_TransmitMasterMessage(c_transmitMasterMessage, MASTERPORT)

		go f_DecrementTimeUntilDisconnect(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface)
		go f_CheckConnectedNodesStatus(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface)
		for {
			select {
			case <-c_nodeIsMaster:
				go f_DecrementTimeUntilReassign(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, c_quitMasterRoutines)
				go f_CheckGlobalQueueEntryStatus(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, c_ackSentGlobalQueueToSlave, c_quitMasterRoutines)
				go f_CheckIfShouldAssign(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, c_ackAssignmentSucessFull, c_assignState, c_quitMasterRoutines)
				go f_CheckAssignedNodeState(c_ackAssignmentSucessFull, c_quitMasterRoutines)
				c_assignState <- ASSIGN
				F_WriteLog("Started all master routines")

			case <-c_nodeIsSlave:
				close(c_quitMasterRoutines)
				c_quitMasterRoutines = make(chan bool)

			default:
				time.Sleep(time.Duration(LEASTRESPONSIVEPERIOD) * time.Microsecond)
			}
		}
	}()

	sendTimer := time.NewTicker(time.Duration(SENDPERIOD) * time.Millisecond)
	printGQTimer := time.NewTicker(time.Duration(2000) * time.Millisecond) //Test function
	assignState := ASSIGN
	nodeRole := f_GetNodeInfo(nodeOperations).MSRole
	if nodeRole == MASTER {
		c_nodeIsMaster <- true
	} else {
		c_nodeIsSlave <- true
	}

	for {
		nodeRole := f_GetNodeInfo().MSRole
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
				transmitterNodeInfo := ackSentGlobalQueueToSlave.ObjectToSupportAcknowledge.(T_NodeInfo)
				globalQueue := ackSentGlobalQueueToSlave.ObjectToAcknowledge.([]T_GlobalQueueEntry)
				masterMessage := T_MasterMessage{
					Transmitter: transmitterNodeInfo,
					GlobalQueue: globalQueue,
				}
				c_transmitMasterMessage <- masterMessage
				ackSentGlobalQueueToSlave.C_Acknowledgement <- true

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
				f_UpdateConnectedNodes(nodeOperations, thisNodeInfo)

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

				if newNodeInfo.MSRole == SLAVE {
					c_nodeIsSlave <- true
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

				if newNodeInfo.MSRole == MASTER {
					c_nodeIsMaster <- true
				}
			}
		}
		time.Sleep(time.Duration(MOSTRESPONSIVEPERIOD) * time.Microsecond)
	}
}
