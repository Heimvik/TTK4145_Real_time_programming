package node

import (
	"fmt"
	"strconv"
	"the-elevator/node/elevator"
	"time"
)

/*
Monitors acknowledgment from the last assigned node to confirm successful assignment of a request, retrying or updating status based on response.

Prerequisites: A global queue with an entry assigned to a node expecting acknowledgment.

Returns: Nothing, but updates assignment status based on acknowledgment or lack thereof.
*/
func f_CheckAssignedNodeState(c_ackAssignmentSucessFull chan T_AckObject, c_receivedActiveEntry chan T_GlobalQueueEntry, c_quit chan bool) {
	for {
	PollLastAssigned:
		select {
		case ackAssignmentSucessFull := <-c_ackAssignmentSucessFull:
			lastAssignedEntry := ackAssignmentSucessFull.ObjectToAcknowledge.(T_GlobalQueueEntry)
			F_WriteLog("Getting ack from last assinged...")
			assignBreakoutTimer := time.NewTicker(time.Duration(ASSIGN_BREAKOUT_PERIOD) * time.Second)
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
					if updatedAssignedNode.ElevatorInfo.State == elevator.ELEVATORSTATE_MOVING && assignedEntry.Request.State == elevator.REQUESTSTATE_ACTIVE {
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
		time.Sleep(time.Duration(MOST_RESPONSIVE_PERIOD) * time.Microsecond)
	}
}

/*
Periodically evaluates changes in the global queue to detect stalls or unchanged conditions, signaling potential issues if no updates occur within a set period.

Prerequisites: An initialized global queue for monitoring.

Returns: Nothing, but sends a status signal indicating system health based on global queue activity.
*/
func f_CheckGlobalQueueEntryStatus(c_getSetGlobalQueueInterface chan T_GetSetGlobalQueueInterface, getSetGlobalQueueInterface T_GetSetGlobalQueueInterface, c_ackSentEntryToSlave chan T_AckObject, c_immobileNode chan uint8, c_nodeWithoutError chan bool, c_quit chan bool) {
	checkChangesToGlobalQueue := time.NewTicker(time.Duration(TERMINATION_PERIOD) * time.Second)
	previousGlobalQueue := f_GetGlobalQueue()
	for {
		select {
		case <-c_quit:
			F_WriteLog("Closed: f_CheckGlobalQueueEntryStatus")
			return
		case <-checkChangesToGlobalQueue.C:
			currentGlobalQueue := f_GetGlobalQueue()
			if f_GlobalQueueAreEqual(previousGlobalQueue, currentGlobalQueue) && len(currentGlobalQueue) != 0 && f_GlobalQueueShouldEmpty(currentGlobalQueue) {
				F_WriteLog(fmt.Sprintf("No changes in globalQueue for %d seconds", TERMINATION_PERIOD))
				c_nodeWithoutError <- false
			} else {
				previousGlobalQueue = currentGlobalQueue
				checkChangesToGlobalQueue.Reset(time.Duration(TERMINATION_PERIOD) * time.Second)
			}
		case immobileNode := <-c_immobileNode:
			c_getSetGlobalQueueInterface <- getSetGlobalQueueInterface
			globalQueue := <-getSetGlobalQueueInterface.c_get
			for entryIndex, entry := range globalQueue {
				if entry.AssignedNode == immobileNode {
					globalQueue = f_ReassignUnfinishedEntry(globalQueue, entry, entryIndex)
				}
			}
			getSetGlobalQueueInterface.c_set <- globalQueue
		default:
			thisNodeInfo := f_GetNodeInfo()
			c_getSetGlobalQueueInterface <- getSetGlobalQueueInterface
			globalQueue := <-getSetGlobalQueueInterface.c_get

			doneEntry, doneEntryIndex := f_FindDoneEntry(globalQueue)
			if (doneEntry.Request.State != elevator.REQUESTSTATE_DONE && doneEntry != T_GlobalQueueEntry{}) {
				globalQueue = f_ReassignUnfinishedEntry(globalQueue, doneEntry, doneEntryIndex)
			}
			getSetGlobalQueueInterface.c_set <- globalQueue

			if (doneEntry.Request.State == elevator.REQUESTSTATE_DONE && doneEntry != T_GlobalQueueEntry{}) {
				globalQueue = f_RemoveFinishedEntry(c_ackSentEntryToSlave, globalQueue, thisNodeInfo, doneEntry, doneEntryIndex)
				f_SetGlobalQueue(globalQueue)
			}

			time.Sleep(time.Duration(LEAST_RESPONSIVE_PERIOD) * time.Microsecond)
		}
	}
}

/*
Regularly inspects the list of connected nodes to remove any that have exceeded their allowed disconnect time, ensuring an up-to-date network state.

Prerequisites: An initialized list of connected nodes with current disconnect timers.

Returns: Nothing, but updates the list of connected nodes by removing those considered disconnected.
*/
func f_CheckConnectedNodesStatus(c_getSetConnectedNodesInterface chan T_GetSetConnectedNodesInterface, getSetConnectedNodesInterface T_GetSetConnectedNodesInterface, c_immobileNode chan uint8) {
	allTimesAtFloorChange := make(map[uint8]time.Time)
	previousConnectedNodes := f_GetConnectedNodes()
	immobileNodes := make(map[uint8]bool)
	for {
		connectedNodes := f_GetConnectedNodes()
		if f_GetNodeInfo().MSRole == MSROLE_MASTER {
			for _, nodeInfo := range connectedNodes {
				previousNodeInfo := f_FindNodeInfo(nodeInfo.PRIORITY, previousConnectedNodes)
				if (previousNodeInfo != T_NodeInfo{}) {
					if nodeInfo.ElevatorInfo.State == elevator.ELEVATORSTATE_MOVING && previousNodeInfo.ElevatorInfo.State == elevator.ELEVATORSTATE_IDLE {
						allTimesAtFloorChange[nodeInfo.PRIORITY] = time.Now()
					}
					if nodeInfo.ElevatorInfo.State == elevator.ELEVATORSTATE_MOVING && previousNodeInfo.ElevatorInfo.Floor != nodeInfo.ElevatorInfo.Floor {
						allTimesAtFloorChange[nodeInfo.PRIORITY] = time.Now()
					}
				}
			}
			previousConnectedNodes = f_CopyConnectedNodes(connectedNodes)
			for node, timeAtFloorChange := range allTimesAtFloorChange {
				timeNow := time.Now()
				if timeNow.Sub(timeAtFloorChange) > time.Duration(IMMOBILE_PERIOD*float64(time.Second)) && f_FindNodeInfo(node, connectedNodes).ElevatorInfo.State == elevator.ELEVATORSTATE_MOVING {
					immobileNodes[node] = false
				}
			}
			for node, isHandled := range immobileNodes {
				if !isHandled {
					c_immobileNode <- node
					immobileNodes[node] = true
					allTimesAtFloorChange[node] = time.Now()
				}
			}
		}

		c_getSetConnectedNodesInterface <- getSetConnectedNodesInterface
		connectedNodes = <-getSetConnectedNodesInterface.c_get
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

		time.Sleep(time.Duration(MOST_RESPONSIVE_PERIOD) * time.Microsecond)
	}
}

/*
Continuously decrements the time until reassignment for each global queue entry, facilitating timely reassignment of unresolved requests.

Prerequisites: An initialized global queue with entries that have reassignment timers.

Returns: Nothing, but modifies reassignment timers for entries in the global queue.
*/
func f_DecrementTimeUntilReassign(c_getSetGlobalQueueInterface chan T_GetSetGlobalQueueInterface, getSetGlobalQueueInterface T_GetSetGlobalQueueInterface, c_quit chan bool) {
	for {
		select {
		case <-c_quit:
			F_WriteLog("Closed: f_DecrementTimeUntilReassign")
			return
		default:
			c_getSetGlobalQueueInterface <- getSetGlobalQueueInterface
			globalQueue := <-getSetGlobalQueueInterface.c_get
			for i, entry := range globalQueue {
				if entry.TimeUntilReassign > 0 && entry.Request.State != elevator.REQUESTSTATE_UNASSIGNED {
					globalQueue[i].TimeUntilReassign -= 1
				}
			}
			getSetGlobalQueueInterface.c_set <- globalQueue
			time.Sleep(1 * time.Second)
		}
	}
}

/*
Periodically reduces the disconnect timers for each connected node, aiding in the timely removal of inactive or unresponsive nodes.

Prerequisites: An initialized list of connected nodes with current disconnect timers.

Returns: Nothing, but adjusts the disconnect timers for nodes in the list of connected nodes.
*/
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

/*
Evaluates whether new elevator requests in the global queue should be assigned to nodes, updating assignment state based on system and elevator availability.

Prerequisites: An initialized global queue and a list of connected, available nodes.

Returns: Nothing, but potentially modifies the global queue with new assignments.
*/
func f_CheckIfShouldAssign(c_getSetGlobalQueueInterface chan T_GetSetGlobalQueueInterface, getSetGlobalQueueInterface T_GetSetGlobalQueueInterface, c_ackAssignmentSucessFull chan T_AckObject, c_assignState chan T_AssignState, c_elevatorWithoutError chan bool, c_quit chan bool) {
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

				assignedEntry, assignedEntryIndex := f_AssignNewEntry(globalQueue, avalibaleNodes)
				if (assignedEntry != T_GlobalQueueEntry{}) {
					globalQueue[assignedEntryIndex] = assignedEntry
					ackAssignmentSucessFull.ObjectToAcknowledge = assignedEntry
					c_ackAssignmentSucessFull <- ackAssignmentSucessFull
					F_WriteLog("Assigned request with ID: " + strconv.Itoa(int(assignedEntry.Request.Id)) + " assigned to node " + strconv.Itoa(int(assignedEntry.AssignedNode)))
					assignState = ASSIGNSTATE_WAITFORACK
					F_WriteLog("Assignstate: 1")
					f_WriteLogGlobalQueueEntry(globalQueue[assignedEntryIndex])
				}
				getSetGlobalQueueInterface.c_set <- globalQueue
				if (assignedEntry != T_GlobalQueueEntry{}) {
					f_WriteLogGlobalQueueEntry(f_GetGlobalQueue()[assignedEntryIndex])
				}
			case ASSIGNSTATE_WAITFORACK:
				select {
				case assigmentWasSucessFull := <-ackAssignmentSucessFull.C_Acknowledgement:
					if assigmentWasSucessFull {
						F_WriteLog("Assignstate: 0")
						assignState = ASSIGNSTATE_ASSIGN
					} else {
						assignState = ASSIGNSTATE_ASSIGN
						c_elevatorWithoutError <- false
						F_WriteLog("Assignstate: 0")
					}
				case <-c_quit:
					F_WriteLog("Closed: f_CheckIfShouldAssign")
					return
				default:
				}
			}
		}
		time.Sleep(time.Duration(LEAST_RESPONSIVE_PERIOD) * time.Microsecond)
	}
}

/*
Manages elevator request assignments by listening for new requests from the elevator, updating assigned requests, and ensuring the elevator processes its currently assigned request.

Prerequisites: An initialized global queue and node information for request handling.

Returns: Nothing, but facilitates the communication and assignment management between the elevator and the control system.
*/
func f_ElevatorManager(c_shouldCheckIfAssigned chan bool, c_entryFromElevator chan T_GlobalQueueEntry, c_getSetElevatorInterface chan elevator.T_GetSetElevatorInterface, c_elevatorWithoutErrors chan bool) {
	c_requestFromElevator := make(chan elevator.T_Request)
	c_requestToElevator := make(chan elevator.T_Request)
	shouldCheckIfAssigned := true

	go elevator.F_RunElevator(elevatorOperations, c_getSetElevatorInterface, c_requestFromElevator, c_requestToElevator, ELEVATOR_PORT, c_elevatorWithoutErrors)

	thisNodeInfo := f_GetNodeInfo()
	globalQueue := f_GetGlobalQueue()
	assignedEntry, _ := f_FindAssignedEntry(globalQueue, thisNodeInfo)
	for {
		select {
		case requestFromElevator := <-c_requestFromElevator:
			entryFromElevator := f_AssembleEntryFromRequest(requestFromElevator, thisNodeInfo, assignedEntry)
			c_entryFromElevator <- entryFromElevator
		case <-c_shouldCheckIfAssigned:
			shouldCheckIfAssigned = true
		default:
			if shouldCheckIfAssigned {
				thisNodeInfo = f_GetNodeInfo()
				globalQueue = f_GetGlobalQueue()
				assignedEntry, _ = f_FindAssignedEntry(globalQueue, thisNodeInfo)
				if (assignedEntry != T_GlobalQueueEntry{}) {
					c_requestToElevator <- assignedEntry.Request
					F_WriteLog("Found assigned entry!")
					shouldCheckIfAssigned = false
				}
			}
			time.Sleep(time.Duration(MOST_RESPONSIVE_PERIOD) * time.Microsecond)
		}
	}
}

/*
Monitors the system for error conditions that could necessitate termination, using periodic checks on both the node and elevator operational statuses.

Prerequisites: Initialized channels for receiving status updates from both node and elevator operations.

Returns: Signals termination through a channel if predefined error thresholds are exceeded.
*/
func F_CheckIfShouldTerminate(c_shouldTerminate chan bool, c_nodeRunningWithoutErrors chan bool, c_elevatorRunningWithoutErrors chan bool) {
	nodeErrors, elevatorErrors := 0, 0
	nodeDeadlockTicker := time.NewTicker(time.Duration(TERMINATION_PERIOD) * time.Second)
	elevatorDeadlockTicker := time.NewTicker(time.Duration(TERMINATION_PERIOD) * time.Second)
	for {
		select {
		case nodeWithoutErrors := <-c_nodeRunningWithoutErrors:
			if nodeWithoutErrors {
				nodeDeadlockTicker.Reset(time.Duration(TERMINATION_PERIOD) * time.Second)
			} else {
				nodeErrors += 1
			}
		case elevatorWithoutErrors := <-c_elevatorRunningWithoutErrors:
			if elevatorWithoutErrors {
				elevatorDeadlockTicker.Reset(time.Duration(TERMINATION_PERIOD) * time.Second)
			} else {
				elevatorErrors += 1
			}
		case <-nodeDeadlockTicker.C:
			fmt.Printf("Node failed to behave correctly for %d seconds", TERMINATION_PERIOD)
			c_shouldTerminate <- true

		case <-elevatorDeadlockTicker.C:
			fmt.Printf("Elevator failed to behave correctly for %d seconds", TERMINATION_PERIOD)
			c_shouldTerminate <- true

		default:
			if nodeErrors > MAX_ALLOWED_NODE_ERRORS || elevatorErrors > MAX_ALLOWED_ELEVATOR_ERRORS {
				fmt.Println("Too many errors from node or elevator")
				time.Sleep(1 * time.Second)
				c_shouldTerminate <- true
			}
		}
	}

}

/*
Initiates and manages the backup operation mode, monitoring for primary node presence and handling messages from both slave and master nodes to maintain system state.

Prerequisites: Configuration for network communication ports must be set.

Returns: Nothing, but transitions the node to a primary role upon detecting absence of a primary node, or updates system state based on received network messages.
*/
func F_RunBackup(c_isPrimary chan bool) {
	c_quitBackupRoutines := make(chan bool)
	c_receiveSlaveMessage := make(chan T_SlaveMessage)
	c_receiveMasterMessage := make(chan T_MasterMessage)

	go f_ReceiveSlaveMessage(c_receiveSlaveMessage, SLAVE_PORT, c_quitBackupRoutines)
	go f_ReceiveMasterMessage(c_receiveMasterMessage, MASTER_PORT, c_quitBackupRoutines)

	PBTicker := time.NewTicker(time.Duration(CONNECTION_PERIOD) * time.Second)
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
				PBTicker.Reset(time.Duration(CONNECTION_PERIOD) * time.Second)
			}
		case slaveMessage := <-c_receiveSlaveMessage:
			thisNodeInfo := f_GetNodeInfo()
			if thisNodeInfo.PRIORITY == slaveMessage.Transmitter.PRIORITY {
				fmt.Println("Slave primary found")
				f_SetNodeInfo(slaveMessage.Transmitter)
				PBTicker.Reset(time.Duration(CONNECTION_PERIOD) * time.Second)
			}
		}
	}
}

/*
Starts primary operation mode, managing global queue, node states, elevator requests, and inter-node communication to orchestrate the entire elevator system.

Prerequisites: Proper initialization of global and node-specific configurations.

Returns: Nothing, but continuously updates the system state based on elevator and node activities.
*/
func F_RunPrimary(c_nodeRunningWithoutErrors chan bool, c_elevatorRunningWithoutErrors chan bool) {
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
	c_immobileNodes := make(chan uint8)
	c_ackAssignmentSucessFull := make(chan T_AckObject)
	c_receivedActiveEntry := make(chan T_GlobalQueueEntry)
	c_ackSentGlobalQueueToSlave := make(chan T_AckObject)

	go func() {
		go f_GetSetNodeInfo(c_getSetNodeInfoInterface)
		go f_GetSetGlobalQueue(c_getSetGlobalQueueInterface)
		go f_GetSetConnectedNodes(c_getSetConnectedNodesInterface)

		go f_ElevatorManager(c_shouldCheckIfAssigned, c_entryFromElevator, c_getSetElevatorInterface, c_elevatorRunningWithoutErrors)
		go f_ReceiveSlaveMessage(c_receiveSlaveMessage, SLAVE_PORT, c_quitReceive)
		go f_ReceiveMasterMessage(c_receiveMasterMessage, MASTER_PORT, c_quitReceive)
		go f_TransmitSlaveMessage(c_transmitSlaveMessage, SLAVE_PORT)
		go f_TransmitMasterMessage(c_transmitMasterMessage, MASTER_PORT)

		go f_DecrementTimeUntilDisconnect(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface)
		go f_CheckConnectedNodesStatus(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface, c_immobileNodes)
		for {
			select {
			case <-c_nodeIsMaster:
				go f_DecrementTimeUntilReassign(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, c_quitMasterRoutines)
				go f_CheckGlobalQueueEntryStatus(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, c_ackSentGlobalQueueToSlave, c_immobileNodes, c_nodeRunningWithoutErrors, c_quitMasterRoutines)
				go f_CheckIfShouldAssign(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, c_ackAssignmentSucessFull, c_assignState, c_elevatorRunningWithoutErrors, c_quitMasterRoutines)
				go f_CheckAssignedNodeState(c_ackAssignmentSucessFull, c_receivedActiveEntry, c_quitMasterRoutines)
				c_assignState <- ASSIGNSTATE_ASSIGN
				F_WriteLog("Started all master routines")

			case <-c_nodeIsSlave:
				close(c_quitMasterRoutines)
				c_quitMasterRoutines = make(chan bool)

			default:
				time.Sleep(time.Duration(LEAST_RESPONSIVE_PERIOD) * time.Microsecond)
			}
		}
	}()

	sendTicker := time.NewTicker(time.Duration(SEND_PERIOD) * time.Millisecond)
	lightsTicker := time.NewTicker(time.Duration(500) * time.Millisecond)
	logTicker := time.NewTicker(time.Duration(2000) * time.Millisecond)

	c_nodeIsMaster <- true

	for {
		nodeRole := f_GetNodeInfo().MSRole
		switch nodeRole {
		case MSROLE_MASTER:
			select {
			case masterMessage := <-c_receiveMasterMessage:
				if masterMessage.Transmitter.PRIORITY != f_GetNodeInfo().PRIORITY {
					f_UpdateGlobalQueue(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, masterMessage)
					f_UpdateConnectedNodes(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface, masterMessage.Transmitter)
				}

			case slaveMessage := <-c_receiveSlaveMessage:
				f_UpdateConnectedNodes(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface, slaveMessage.Transmitter)
				if slaveMessage.Entry.Request.Calltype != elevator.CALLTYPE_NONECALL {
					f_AddEntryGlobalQueue(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, slaveMessage.Entry)
				}
				if slaveMessage.Entry.Request.State == elevator.REQUESTSTATE_ACTIVE {
					c_receivedActiveEntry <- slaveMessage.Entry
				}

			case entryFromElevator := <-c_entryFromElevator:
				f_AddEntryGlobalQueue(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, entryFromElevator)
				if entryFromElevator.Request.State == elevator.REQUESTSTATE_DONE {
					c_shouldCheckIfAssigned <- true
				} else if entryFromElevator.Request.State == elevator.REQUESTSTATE_ACTIVE {
					c_receivedActiveEntry <- entryFromElevator
				}
				F_WriteLog("Node: | " + strconv.Itoa(int(f_GetNodeInfo().PRIORITY)) + " | MASTER | updated GQ entry:\n")
				f_WriteLogGlobalQueueEntry(entryFromElevator)

			case ackSentGlobalQueueToSlave := <-c_ackSentGlobalQueueToSlave:
				transmitterNodeInfo := ackSentGlobalQueueToSlave.ObjectToSupportAcknowledge.(T_NodeInfo)
				globalQueue := ackSentGlobalQueueToSlave.ObjectToAcknowledge.([]T_GlobalQueueEntry)
				masterMessage := T_MasterMessage{
					Transmitter: transmitterNodeInfo,
					GlobalQueue: f_CopyGlobalQueue(globalQueue),
				}
				c_transmitMasterMessage <- masterMessage
				ackSentGlobalQueueToSlave.C_Acknowledgement <- true

			case <-sendTicker.C:
				f_TransmitMasterInfo(c_transmitMasterMessage)
				sendTicker.Reset(time.Duration(SEND_PERIOD) * time.Millisecond)

			case <-lightsTicker.C:
				f_UpdateLights()
				lightsTicker.Reset(time.Duration(500) * time.Millisecond)

			case <-logTicker.C:
				globalQueue := f_GetGlobalQueue()
				nodeInfo := f_GetNodeInfo()
				connectedNodes := f_GetConnectedNodes()
				f_WriteLogConnectedNodes(connectedNodes)
				F_WriteLog("Node: | " + strconv.Itoa(int(nodeInfo.PRIORITY)) + " | MASTER | has GQ:\n")
				for _, entry := range globalQueue {
					f_WriteLogGlobalQueueEntry(entry)
				}
				logTicker.Reset(time.Duration(2000) * time.Millisecond)

			default:
				connectedNodes := f_GetConnectedNodes()
				c_getSetNodeInfoInterface <- getSetNodeInfoInterface
				oldNodeInfo := <-getSetNodeInfoInterface.c_get
				newNodeInfo := f_AssignNewRole(oldNodeInfo, connectedNodes)
				getSetNodeInfoInterface.c_set <- newNodeInfo
				thisNodeInfo := newNodeInfo
				f_UpdateConnectedNodes(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface, thisNodeInfo)
				if newNodeInfo.MSRole == MSROLE_SLAVE {
					f_TransmitMasterInfo(c_transmitMasterMessage)
					c_nodeIsSlave <- true
				}
			}

		case MSROLE_SLAVE:
			select {
			case masterMessage := <-c_receiveMasterMessage:
				f_UpdateGlobalQueue(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, masterMessage)
				f_UpdateConnectedNodes(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface, masterMessage.Transmitter)

			case slaveMessage := <-c_receiveSlaveMessage:
				if slaveMessage.Transmitter.PRIORITY != f_GetNodeInfo().PRIORITY {
					f_UpdateConnectedNodes(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface, slaveMessage.Transmitter)
				}

			case entryFromElevator := <-c_entryFromElevator:
				f_AddEntryGlobalQueue(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, entryFromElevator)
				if entryFromElevator.Request.State == elevator.REQUESTSTATE_DONE {
					c_shouldCheckIfAssigned <- true
				}
				thisNode := f_GetNodeInfo()
				F_WriteLog("Node: | " + strconv.Itoa(int(thisNode.PRIORITY)) + " | SLAVE | updated GQ entry:\n")
				f_WriteLogGlobalQueueEntry(entryFromElevator)
				infoMessage := T_SlaveMessage{
					Transmitter: f_GetNodeInfo(),
					Entry:       entryFromElevator,
				}
				c_transmitSlaveMessage <- infoMessage

			case <-sendTicker.C:
				f_TransmitSlaveInfo(c_transmitSlaveMessage)
				sendTicker.Reset(time.Duration(SEND_PERIOD) * time.Millisecond)

			case <-lightsTicker.C:
				f_UpdateLights()
				lightsTicker.Reset(time.Duration(500) * time.Millisecond)

			case <-logTicker.C:
				globalQueue := f_GetGlobalQueue()
				nodeInfo := f_GetNodeInfo()
				connectedNodes := f_GetConnectedNodes()
				f_WriteLogConnectedNodes(connectedNodes)
				F_WriteLog("Node: | " + strconv.Itoa(int(nodeInfo.PRIORITY)) + " | SLAVE | has GQ:\n")
				for _, entry := range globalQueue {
					f_WriteLogGlobalQueueEntry(entry)
				}
				logTicker.Reset(time.Duration(2000) * time.Millisecond)
			default:
				connectedNodes := f_GetConnectedNodes()
				c_getSetNodeInfoInterface <- getSetNodeInfoInterface
				oldNodeInfo := <-getSetNodeInfoInterface.c_get
				newNodeInfo := f_AssignNewRole(oldNodeInfo, connectedNodes)
				getSetNodeInfoInterface.c_set <- newNodeInfo
				thisNodeInfo := newNodeInfo
				f_UpdateConnectedNodes(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface, thisNodeInfo)

				if newNodeInfo.MSRole == MSROLE_MASTER {
					f_TransmitSlaveInfo(c_transmitSlaveMessage)
					c_nodeIsMaster <- true
				}
			}
		}
		c_nodeRunningWithoutErrors <- true
		time.Sleep(time.Duration(MOST_RESPONSIVE_PERIOD) * time.Microsecond)
	}
}
