package node

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"the-elevator/node/elevator"
	"time"
)

func init() {
	logFile, _ := os.OpenFile("log/debug1.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	defer logFile.Close()

	configFile, _ := os.Open("config/default.json")
	defer configFile.Close()

	var config T_Config
	json.NewDecoder(configFile).Decode(&config)

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

func f_CheckAssignedNodeState(c_ackAssignmentSucessFull chan T_AckObject, c_receivedActiveEntry chan T_GlobalQueueEntry, c_quit chan bool) {
	for {
	PollLastAssigned:
		select {
		case ackAssignmentSucessFull := <-c_ackAssignmentSucessFull:
			lastAssignedEntry := ackAssignmentSucessFull.ObjectToAcknowledge.(T_GlobalQueueEntry)
			F_WriteLog("Getting ack from last assinged...")
			assignBreakoutTimer := time.NewTicker(time.Duration(ASSIGNBREAKOUTPERIOD) * time.Second)
			for {
				select {
				case <-assignBreakoutTimer.C:
					ackAssignmentSucessFull.C_Acknowledgement <- false
					assignBreakoutTimer.Stop()
					break PollLastAssigned
				case <-c_quit:
					F_WriteLog("Closed: f_CheckAssignedNodeState")
					return
				case receivedActiveEntry := <-c_receivedActiveEntry:
					assignedEntry := f_FindEntry(lastAssignedEntry, f_GetGlobalQueue())
					if f_EntriesAreEqual(assignedEntry, receivedActiveEntry) {
						ackAssignmentSucessFull.C_Acknowledgement <- true
						F_WriteLog("Found ack (received ACTIVE)")
						assignBreakoutTimer.Stop()
						break PollLastAssigned
					}
				default:
					assignedEntry := f_FindEntry(lastAssignedEntry, f_GetGlobalQueue())
					updatedAssignedNode := f_FindNodeInfo(lastAssignedEntry.AssignedNode, f_GetConnectedNodes())
					if updatedAssignedNode.ElevatorInfo.State == elevator.MOVING && assignedEntry.Request.State == elevator.ACTIVE {
						ackAssignmentSucessFull.C_Acknowledgement <- true
						F_WriteLog("Found ack (changes in connected nodes)")
						assignBreakoutTimer.Stop()
						break PollLastAssigned
					}
				}
			}
		case <-c_receivedActiveEntry:
			F_WriteLog("ACTIVE entry already handeled")
		case <-c_quit:
			F_WriteLog("Closed: f_CheckAssignedNodeState")
			return
		}
		time.Sleep(time.Duration(MOSTRESPONSIVEPERIOD) * time.Microsecond)
	}
}

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
					doneEntryIndex = i
				}
			}
			if (doneEntry.Request.State != elevator.DONE && doneEntry != T_GlobalQueueEntry{}) {
				globalQueue = f_ReassignUnfinishedEntry(globalQueue, doneEntry, doneEntryIndex)
			}
			getSetGlobalQueueInterface.c_set <- globalQueue

			if (doneEntry.Request.State == elevator.DONE && doneEntry != T_GlobalQueueEntry{}) {
				globalQueue = f_RemoveFinishedEntry(c_ackSentEntryToSlave, globalQueue, thisNodeInfo, doneEntry, doneEntryIndex)
				f_SetGlobalQueue(globalQueue)
			}

			time.Sleep(time.Duration(LEASTRESPONSIVEPERIOD) * time.Microsecond)
		}
	}
}

func f_CheckConnectedNodesStatus(c_getSetConnectedNodesInterface chan T_GetSetConnectedNodesInterface, getSetConnectedNodesInterface T_GetSetConnectedNodesInterface) { //begge
	for {
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

func f_CheckIfShouldAssign(c_getSetGlobalQueueInterface chan T_GetSetGlobalQueueInterface, getSetGlobalQueueInterface T_GetSetGlobalQueueInterface, c_ackAssignmentSucessFull chan T_AckObject, c_assignState chan T_AssignState, c_quit chan bool) {
	assignState := ASSIGNSTATE_ASSIGN
	c_assignmentSuccessfull := make(chan bool)
	ackAssignmentSucessFull := T_AckObject{
		C_Acknowledgement: c_assignmentSuccessfull,
	}
	for {
		select {
		case assignState = <-c_assignState:
			F_WriteLog("Assignstate: " + strconv.Itoa(int(assignState)))
		case <-c_quit:
			F_WriteLog("Closed: f_CheckIfShouldAssign")
			return
		default:
			switch assignState {
			case ASSIGNSTATE_ASSIGN:
				connectedNodes := f_GetConnectedNodes()
				avalibaleNodes := f_GetAvalibaleNodes(connectedNodes)
				c_getSetGlobalQueueInterface <- getSetGlobalQueueInterface
				globalQueue := <-getSetGlobalQueueInterface.c_get

				assignedEntry, assignedEntryIndex := F_AssignNewEntry(globalQueue, connectedNodes, avalibaleNodes)
				if (assignedEntry != T_GlobalQueueEntry{}) {
					globalQueue[assignedEntryIndex] = assignedEntry
					ackAssignmentSucessFull.ObjectToAcknowledge = assignedEntry
					c_ackAssignmentSucessFull <- ackAssignmentSucessFull
					F_WriteLog("Assigned request with ID: " + strconv.Itoa(int(assignedEntry.Request.Id)) + " assigned to node " + strconv.Itoa(int(assignedEntry.AssignedNode)))
					assignState = ASSIGNSTATE_WAITFORACK
					F_WriteLog("Assignstate: 1")
					F_WriteLogGlobalQueueEntry(globalQueue[assignedEntryIndex])
				}
				getSetGlobalQueueInterface.c_set <- globalQueue
				if (assignedEntry != T_GlobalQueueEntry{}) {
					F_WriteLogGlobalQueueEntry(f_GetGlobalQueue()[assignedEntryIndex])
				}
			case ASSIGNSTATE_WAITFORACK:
				select {
				case assigmentWasSucessFull := <-ackAssignmentSucessFull.C_Acknowledgement:
					if assigmentWasSucessFull {
						F_WriteLog("Assignstate: 0")
						assignState = ASSIGNSTATE_ASSIGN
					} else {
						assignState = ASSIGNSTATE_ASSIGN
						F_WriteLog("Assignstate: 0")
					}
				default:
				}
			}

		}
		time.Sleep(time.Duration(LEASTRESPONSIVEPERIOD) * time.Microsecond)
	}
}

/*
Psudocode for Lightsmanager. All elevetors should show the same lights under normal circumstances, except
cab light, which is not shared.

We will first try the folowin approach:
Deciding if lights should be on shuld be determined by any !DONE entries in f_getGlobalQueue.
This locic is implemented in the for loop below.

Potential problems:
This is dependent of polling the GQ, and not message based, such that it might not register sudden changes in
lights, but might not be a problem. An alternative solution would be to syncronize with a c_entryDone channel
everywhere we are reamoving from GQ (after removal). In this case deciding if lights should be on/off should
be determined by the signaling of unassigned/done entries further out in the program.

*/

func f_TurnOnLight(entry T_GlobalQueueEntry, entryIsUnique bool) {
	thisNodeInfo := f_GetNodeInfo()

	//Turn on light for new entry in global queue
	if entry.Request.State != elevator.DONE && entryIsUnique {
		if entry.Request.Calltype == elevator.HALL && entry.Request.Direction == elevator.DOWN{
			elevator.F_SetButtonLamp(elevator.BT_HallDown, int(entry.Request.Floor), true)

		}else if entry.Request.Calltype == elevator.HALL && entry.Request.Direction == elevator.UP{
			elevator.F_SetButtonLamp(elevator.BT_HallUp, int(entry.Request.Floor), true)

		}else if entry.Request.Calltype == elevator.CAB && entry.AssignedNode == thisNodeInfo.PRIORITY{
			elevator.F_SetButtonLamp(elevator.BT_Cab, int(entry.Request.Floor), true)
		}
	}
}

func f_TurnOffLight(entry T_GlobalQueueEntry) {
	thisNodeInfo := f_GetNodeInfo()

	//Turn off light for done entry in global queue
	if entry.Request.State == elevator.DONE {
		if entry.Request.Calltype == elevator.HALL && entry.Request.Direction == elevator.DOWN{
			elevator.F_SetButtonLamp(elevator.BT_HallDown, int(entry.Request.Floor), false)

		}else if entry.Request.Calltype == elevator.HALL && entry.Request.Direction == elevator.UP{
			elevator.F_SetButtonLamp(elevator.BT_HallUp, int(entry.Request.Floor), false)

		}else if entry.Request.Calltype == elevator.CAB && entry.AssignedNode == thisNodeInfo.PRIORITY{
			elevator.F_SetButtonLamp(elevator.BT_Cab, int(entry.Request.Floor), false)
		}
	}
}

func f_ElevatorManager(c_shouldCheckIfAssigned chan bool, c_entryFromElevator chan T_GlobalQueueEntry, c_getSetElevatorInterface chan elevator.T_GetSetElevatorInterface) {
	c_requestFromElevator := make(chan elevator.T_Request)
	c_requestToElevator := make(chan elevator.T_Request)
	shouldCheckIfAssigned := true

	go elevator.F_RunElevator(elevatorOperations, c_getSetElevatorInterface, c_requestFromElevator, c_requestToElevator, ELEVATORPORT)
	//go elevator.F_SimulateRequest(elevatorOperations, c_requestFromElevator, c_requestToElevator)

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
					c_requestToElevator <- assignedEntry.Request
					shouldCheckIfAssigned = false
				}
			}
			time.Sleep(time.Duration(MOSTRESPONSIVEPERIOD) * time.Microsecond)
		}
	}
}

func F_ProcessPairManager() {
	fmt.Println("Checking for primaries...")
	go f_NodeOperationManager(&ThisNode) //Should be only reference to ThisNode

	c_isPrimary := make(chan bool)
	go f_RunBackup(c_isPrimary)
	select {
	case <-c_isPrimary:
		err := exec.Command("gnome-terminal", "--", "go", "run", "main.go").Run()
		if err != nil {
			F_WriteLog("Error starting BACKUP")
		}
		F_WriteLog("Switched to primary")
		f_RunPrimary()
	}
}

func f_RunBackup(c_isPrimary chan bool) {
	c_quitBackupRoutines := make(chan bool)
	c_receiveSlaveMessage := make(chan T_SlaveMessage)
	c_receiveMasterMessage := make(chan T_MasterMessage)

	go F_ReceiveSlaveMessage(c_receiveSlaveMessage, SLAVEPORT, c_quitBackupRoutines)
	go F_ReceiveMasterMessage(c_receiveMasterMessage, MASTERPORT, c_quitBackupRoutines)

	PBTicker := time.NewTicker(time.Duration(CONNECTIONTIME) * time.Second)
	for {
		select {
		case <-PBTicker.C:
			fmt.Println("No primary found, initiates primary role")
			close(c_quitBackupRoutines)
			c_isPrimary <- true
			return
		case masterMessage := <-c_receiveMasterMessage:
			thisNodeInfo := f_GetNodeInfo()
			f_SetGlobalQueue(masterMessage.GlobalQueue)
			if thisNodeInfo.PRIORITY == masterMessage.Transmitter.PRIORITY {
				fmt.Println("Master primary found")
				f_SetNodeInfo(masterMessage.Transmitter)
				PBTicker.Reset(time.Duration(CONNECTIONTIME) * time.Second)
			}
		case slaveMessage := <-c_receiveSlaveMessage:
			thisNodeInfo := f_GetNodeInfo()
			if thisNodeInfo.PRIORITY == slaveMessage.Transmitter.PRIORITY {
				fmt.Println("Slave primary found")
				f_SetNodeInfo(slaveMessage.Transmitter)
				PBTicker.Reset(time.Duration(CONNECTIONTIME) * time.Second)
			}
		}
	}
}

func f_RunPrimary() {
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
	c_getSetElevatorInterface := make(chan elevator.T_GetSetElevatorInterface)

	c_nodeIsMaster := make(chan bool)
	c_quitMasterRoutines := make(chan bool)
	c_quitReceive := make(chan bool)
	c_nodeIsSlave := make(chan bool)

	c_receiveSlaveMessage := make(chan T_SlaveMessage)
	c_receiveMasterMessage := make(chan T_MasterMessage)
	c_transmitMasterMessage := make(chan T_MasterMessage)
	c_transmitSlaveMessage := make(chan T_SlaveMessage)

	c_entryFromElevator := make(chan T_GlobalQueueEntry)
	c_shouldCheckIfAssigned := make(chan bool)

	c_assignState := make(chan T_AssignState)
	c_ackAssignmentSucessFull := make(chan T_AckObject)
	c_receivedActiveEntry := make(chan T_GlobalQueueEntry)
	c_ackSentGlobalQueueToSlave := make(chan T_AckObject)

	go func() {
		go f_GetSetNodeInfo(c_getSetNodeInfoInterface)
		go f_GetSetGlobalQueue(c_getSetGlobalQueueInterface)
		go f_GetSetConnectedNodes(c_getSetConnectedNodesInterface)

		go f_ElevatorManager(c_shouldCheckIfAssigned, c_entryFromElevator, c_getSetElevatorInterface)
		go F_ReceiveSlaveMessage(c_receiveSlaveMessage, SLAVEPORT, c_quitReceive)
		go F_ReceiveMasterMessage(c_receiveMasterMessage, MASTERPORT, c_quitReceive)
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
				go f_CheckAssignedNodeState(c_ackAssignmentSucessFull, c_receivedActiveEntry, c_quitMasterRoutines)
				c_assignState <- ASSIGNSTATE_ASSIGN
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
	logTimer := time.NewTicker(time.Duration(2000) * time.Millisecond)

	c_nodeIsMaster <- true

	for {
		nodeRole := f_GetNodeInfo().MSRole
		switch nodeRole {
		case MSROLE_MASTER:
			select {
			case masterMessage := <-c_receiveMasterMessage:
				if masterMessage.Transmitter.PRIORITY != f_GetNodeInfo().PRIORITY {
					f_UpdateGlobalQueueMM(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, masterMessage)
					f_UpdateConnectedNodes(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface, masterMessage.Transmitter)
				}
				f_WriteLogMasterMessage(masterMessage)

			case slaveMessage := <-c_receiveSlaveMessage:
				f_UpdateConnectedNodes(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface, slaveMessage.Transmitter)
				f_UpdateGlobalQueueSM(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, slaveMessage, c_receivedActiveEntry)
				f_WriteLogSlaveMessage(slaveMessage)

			case entryFromElevator := <-c_entryFromElevator:
				f_AddEntryGlobalQueue(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, entryFromElevator)
				if entryFromElevator.Request.State == elevator.DONE {
					c_shouldCheckIfAssigned <- true
				} else if entryFromElevator.Request.State == elevator.ACTIVE {
					c_receivedActiveEntry <- entryFromElevator
				}
				F_WriteLog("Node: | " + strconv.Itoa(int(f_GetNodeInfo().PRIORITY)) + " | MASTER | updated GQ entry:\n")
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
				masterMessage := T_MasterMessage{
					Transmitter: f_GetNodeInfo(),
					GlobalQueue: f_GetGlobalQueue(),
				}
				c_transmitMasterMessage <- masterMessage
				sendTimer.Reset(time.Duration(SENDPERIOD) * time.Millisecond)

			case <-logTimer.C:
				globalQueue := f_GetGlobalQueue()
				nodeInfo := f_GetNodeInfo()
				connectedNodes := f_GetConnectedNodes()
				f_WriteLogConnectedNodes(connectedNodes)
				F_WriteLog("Node: | " + strconv.Itoa(int(nodeInfo.PRIORITY)) + " | MASTER | has GQ:\n")
				for _, entry := range globalQueue {
					F_WriteLogGlobalQueueEntry(entry)
				}
				logTimer.Reset(time.Duration(2000) * time.Millisecond)

			default:
				connectedNodes := f_GetConnectedNodes()
				c_getSetNodeInfoInterface <- getSetNodeInfoInterface
				oldNodeInfo := <-getSetNodeInfoInterface.c_get
				newNodeInfo := f_AssignNewRole(oldNodeInfo, connectedNodes)
				getSetNodeInfoInterface.c_set <- newNodeInfo
				thisNodeInfo := newNodeInfo
				f_UpdateConnectedNodes(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface, thisNodeInfo)
				if newNodeInfo.MSRole == MSROLE_SLAVE {
					c_nodeIsSlave <- true
				}
			}

		case MSROLE_SLAVE:
			select {
			case masterMessage := <-c_receiveMasterMessage:
				f_UpdateGlobalQueueMM(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, masterMessage)
				f_UpdateConnectedNodes(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface, masterMessage.Transmitter)
				f_WriteLogMasterMessage(masterMessage)

			case slaveMessage := <-c_receiveSlaveMessage:
				if slaveMessage.Transmitter.PRIORITY != f_GetNodeInfo().PRIORITY {
					f_UpdateConnectedNodes(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface, slaveMessage.Transmitter)
				}
				f_WriteLogSlaveMessage(slaveMessage)

			case entryFromElevator := <-c_entryFromElevator:
				f_AddEntryGlobalQueue(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, entryFromElevator)
				if entryFromElevator.Request.State == elevator.DONE {
					c_shouldCheckIfAssigned <- true
				}
				thisNode := f_GetNodeInfo()
				F_WriteLog("Node: | " + strconv.Itoa(int(thisNode.PRIORITY)) + " | SLAVE | updated GQ entry:\n")
				F_WriteLogGlobalQueueEntry(entryFromElevator)
				infoMessage := T_SlaveMessage{
					Transmitter: f_GetNodeInfo(),
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
			case <-logTimer.C:
				globalQueue := f_GetGlobalQueue()
				nodeInfo := f_GetNodeInfo()
				connectedNodes := f_GetConnectedNodes()
				f_WriteLogConnectedNodes(connectedNodes)
				F_WriteLog("Node: | " + strconv.Itoa(int(nodeInfo.PRIORITY)) + " | SLAVE | has GQ:\n")
				for _, entry := range globalQueue {
					F_WriteLogGlobalQueueEntry(entry)
				}
				logTimer.Reset(time.Duration(2000) * time.Millisecond)
			default:
				connectedNodes := f_GetConnectedNodes()
				c_getSetNodeInfoInterface <- getSetNodeInfoInterface
				oldNodeInfo := <-getSetNodeInfoInterface.c_get
				newNodeInfo := f_AssignNewRole(oldNodeInfo, connectedNodes)
				getSetNodeInfoInterface.c_set <- newNodeInfo
				thisNodeInfo := newNodeInfo
				f_UpdateConnectedNodes(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface, thisNodeInfo)

				if newNodeInfo.MSRole == MSROLE_MASTER {
					c_nodeIsMaster <- true
				}
			}
		}
		time.Sleep(time.Duration(MOSTRESPONSIVEPERIOD) * time.Microsecond)
	}
}
