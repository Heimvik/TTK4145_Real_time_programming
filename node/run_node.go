package node

import (
	"encoding/json"
	"os"
	"strconv"
	"the-elevator/node/elevator"
	"time"
)

// ***	START TEST FUNCTIONS	***//

// ***	END TEST FUNCTIONS	***//

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

func f_CheckAssignedNodeState(c_lastAssignedEntry chan T_GlobalQueueEntry, c_assignmentSuccessfull chan bool, c_quit chan bool) {
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
				}
			}
		case <-c_quit:
			return
		}
		time.Sleep(time.Duration(MEDIUMRESPONSIVEPERIOD) * time.Microsecond)
	}
}

func f_CheckGlobalQueueEntryStatus(c_getSetGlobalQueueInterface chan T_GetSetGlobalQueueInterface, getSetGlobalQueueInterface T_GetSetGlobalQueueInterface, c_ackSentEntryToSlave chan T_AckObject, c_quit chan bool) {
	for {
		select {
		case <-c_quit:
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

func f_ElevatorManager(c_shouldCheckIfAssigned chan bool, c_entryFromElevator chan T_GlobalQueueEntry) {
	c_requestFromElevator := make(chan elevator.T_Request)
	c_requestToElevator := make(chan elevator.T_Request)
	shouldCheckIfAssigned := true

	//go elevator.F_RunElevator(elevatorOps, c_requestFromElevator, c_requestToElevator, ELEVATORPORT)
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

//IMPORTANT:
//-global variables should ALWAYS be handled by server to operate onn good data

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

	c_quitDecrementTimeUntilReassign := make(chan bool)
	c_quitCheckGlobalQueueEntryStatus := make(chan bool)
	c_quitCheckAssignedNodeState := make(chan bool)

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

		go f_DecrementTimeUntilDisconnect(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface)
		go f_CheckConnectedNodesStatus(c_getSetConnectedNodesInterface, getSetConnectedNodesInterface)
		for {
			select {
			case <-c_nodeIsMaster:
				go f_DecrementTimeUntilReassign(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, c_quitDecrementTimeUntilReassign)
				go f_CheckGlobalQueueEntryStatus(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, c_ackSentGlobalQueueToSlave, c_quitCheckGlobalQueueEntryStatus)
				go f_CheckAssignedNodeState(c_lastAssignedEntry, c_assignmentWasSucessFull, c_quitCheckAssignedNodeState)

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
					c_quitDecrementTimeUntilReassign <- true
					c_quitCheckGlobalQueueEntryStatus <- true
					c_quitCheckAssignedNodeState <- true
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
				}
			}
		}
		time.Sleep(time.Duration(MOSTRESPONSIVEPERIOD) * time.Microsecond)
	}
}
