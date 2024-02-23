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
func f_ChooseRole(thisNodeInfo T_NodeInfo, connectedNodes []T_NodeInfo) T_NodeRole {
	var returnRole T_NodeRole
	for _, remoteNodeInfo := range connectedNodes {
		if remoteNodeInfo.PRIORITY < thisNodeInfo.PRIORITY {
			returnRole = SLAVE
		} else {
			returnRole = MASTER
		}
	}
	return returnRole
}
func f_RemoveNode(nodes []T_NodeInfo, nodeToRemove T_NodeInfo) []T_NodeInfo {
	for i, nodeInfo := range nodes {
		if nodeInfo.PRIORITY == nodeToRemove.PRIORITY {
			return append(nodes[:i], nodes[i+1:]...)
		}
	}
	return nodes
}

// decrements all timers in GQ and checks for any that has run out, will always try to reassign/remove GQ elements/connectednodes
func f_TimeManager(ops T_NodeOperations, c_quit chan bool, c_send chan bool) {
	for {
		globalQueue := f_GetGlobalQueue(ops)
		for _, element := range globalQueue {
			element.TimeUntilReassign -= 1
			if element.TimeUntilReassign == 0 && element.Request.State != elevator.DONE {
				f_AddEntryGlobalQueue(ops, element)
				//Remove old element?
			}
		}

		oldConnectedNodes := f_GetConnectedNodes(ops)
		for _, element := range oldConnectedNodes {
			element.TimeUntilDisconnect -= 1
			fmt.Println(strconv.Itoa(element.PRIORITY) + " | " + strconv.FormatFloat(float64(element.TimeUntilDisconnect), 'f', -1, 32))
			if element.TimeUntilDisconnect == 0 {
				newConnectedNodes := f_RemoveNode(oldConnectedNodes, element)
				f_SetConnectedNodes(ops, newConnectedNodes)
				break
			}
		}
		select {
		case <-c_quit:
			F_WriteLog("Closed TimeManager goroutine in master")
			return
		default:
			time.Sleep(1 * time.Second)
		}
	}
}
func f_NodeOperationManager(node *T_Node, ops T_NodeOperations) {
	for {
		select {
		case responseChan := <-ops.c_readNodeInfo:
			responseChan <- node.Info

		case newNodeInfo := <-ops.c_writeNodeInfo:
			node.Info = newNodeInfo

		case responseChan := <-ops.c_readGlobalQueue:
			responseChan <- node.GlobalQueue

		case newGlobalQueue := <-ops.c_writeGlobalQueue:
			node.GlobalQueue = newGlobalQueue

		case responseChan := <-ops.c_readConnectedNodes:
			responseChan <- node.ConnectedNodes

		case newConnectedNodes := <-ops.c_writeConnectedNodes:
			node.ConnectedNodes = newConnectedNodes

		case responseChan := <-ops.c_readElevator:
			responseChan <- *node.P_ELEVATOR

		case newElevator := <-ops.c_writeElevator:
			*node.P_ELEVATOR = newElevator
		}
	}
}
func f_GetNodeInfo(ops T_NodeOperations) T_NodeInfo {
	responseChan := make(chan T_NodeInfo)
	ops.c_readNodeInfo <- responseChan // Send the response channel to the NodeOperationManager
	nodeInfo := <-responseChan         // Receive the node info from the response channel
	return nodeInfo
}
func f_SetNodeInfo(ops T_NodeOperations, nodeInfo T_NodeInfo) {
	ops.c_writeNodeInfo <- nodeInfo // Send the nodeInfo directly to be written
}
func f_GetGlobalQueue(ops T_NodeOperations) []T_GlobalQueueEntry {
	responseChan := make(chan []T_GlobalQueueEntry)
	ops.c_readGlobalQueue <- responseChan // Send the response channel to the NodeOperationManager
	globalQueue := <-responseChan         // Receive the global queue from the response channel
	return globalQueue
}
func f_SetGlobalQueue(ops T_NodeOperations, globalQueue []T_GlobalQueueEntry) {
	ops.c_writeGlobalQueue <- globalQueue // Send the globalQueue directly to be written
}
func f_GetConnectedNodes(ops T_NodeOperations) []T_NodeInfo {
	responseChan := make(chan []T_NodeInfo)
	ops.c_readConnectedNodes <- responseChan // Send the response channel to the NodeOperationManager
	connectedNodes := <-responseChan         // Receive the connected nodes from the response channel
	return connectedNodes
}
func f_SetConnectedNodes(ops T_NodeOperations, connectedNodes []T_NodeInfo) {
	ops.c_writeConnectedNodes <- connectedNodes // Send the connectedNodes directly to be written
}
func f_GetElevator(ops T_NodeOperations) elevator.T_Elevator {
	responseChan := make(chan elevator.T_Elevator)
	ops.c_readElevator <- responseChan // Send the response channel to the NodeOperationManager
	elevator := <-responseChan         // Receive the connected nodes from the response channel
	return elevator
}
func f_SetElevator(ops T_NodeOperations, elevator elevator.T_Elevator) {
	ops.c_writeElevator <- elevator // Send the connectedNodes directly to be written
}
func f_UpdateConnectedNodes(ops T_NodeOperations, currentNode T_NodeInfo) {
	oldConnectedNodes := f_GetConnectedNodes(ops)
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
		f_SetConnectedNodes(ops, connectedNodes)
	} else {
		currentNode.TimeUntilDisconnect = CONNECTIONTIME
		oldConnectedNodes[nodeIndex] = currentNode
		f_SetConnectedNodes(ops, oldConnectedNodes)
	}
}

func f_AddEntryGlobalQueue(ops T_NodeOperations, entryToAdd T_GlobalQueueEntry) {
	thisGlobalQueue := f_GetGlobalQueue(ops)
	entryIsUnique := true
	entryIndex := 0
	for i, entry := range thisGlobalQueue {
		if entryToAdd.Request.Id == entry.Request.Id && entryToAdd.RequestedNode == entry.RequestedNode { //random id generated to each entry
			entryIsUnique = false
			entryIndex = i
			break
		}
	}
	if entryIsUnique {
		entryToAdd.AssignedNode.PRIORITY = 0
		//entryToAdd.TimeUntilReassign = REASSIGNTIME //Should be set by slave
		thisGlobalQueue = append(thisGlobalQueue, entryToAdd)
		f_SetGlobalQueue(ops, thisGlobalQueue)
	} else { //should update the existing entry
		thisGlobalQueue[entryIndex] = entryToAdd
		f_SetGlobalQueue(ops, thisGlobalQueue)
	}
}
func f_ElevatorManager(ops T_NodeOperations, c_entryFromElevator chan T_GlobalQueueEntry, c_quit chan bool) {

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
		case <-c_quit:
			F_WriteLog("Closed ElevatorManager goroutine in master")
			return
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
	//to run the main FSM
	c_nodeOpMsg := T_NodeOperations{
		c_readNodeInfo:        make(chan chan T_NodeInfo),
		c_writeNodeInfo:       make(chan T_NodeInfo),
		c_readGlobalQueue:     make(chan chan []T_GlobalQueueEntry),
		c_writeGlobalQueue:    make(chan []T_GlobalQueueEntry),
		c_readConnectedNodes:  make(chan chan []T_NodeInfo),
		c_writeConnectedNodes: make(chan []T_NodeInfo),
		c_readElevator:        make(chan chan elevator.T_Elevator),
	}
	c_receiveSlaveMessage := make(chan T_SlaveMessage)
	c_receiveMasterMessage := make(chan T_MasterMessage)
	c_transmitMasterMessage := make(chan T_MasterMessage)
	c_transmitSlaveMessage := make(chan T_SlaveMessage)

	c_quit := make(chan bool)
	c_send := make(chan bool)
	c_entryFromElevator := make(chan T_GlobalQueueEntry)

	go f_TimeManager(c_nodeOpMsg, c_quit, c_send)
	go f_NodeOperationManager(&ThisNode, c_nodeOpMsg) //SHOULD BE THE ONLY REFERENCE TO ThisNode!
	go f_ElevatorManager(c_nodeOpMsg, c_entryFromElevator, c_quit)
	go F_ReceiveSlaveMessage(c_receiveSlaveMessage, c_nodeOpMsg, SLAVEPORT)
	go F_ReceiveMasterMessage(c_receiveMasterMessage, c_nodeOpMsg, MASTERPORT)
	go F_TransmitSlaveMessage(c_transmitSlaveMessage, SLAVEPORT)
	go F_TransmitMasterMessage(c_transmitMasterMessage, MASTERPORT)
	sendTimer := time.NewTicker(time.Duration(SENDPERIOD) * time.Millisecond)
	for {
		nodeRole := f_GetNodeInfo(c_nodeOpMsg).Role
		switch nodeRole {
		case MASTER:
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
				connectedNodes := f_GetConnectedNodes(c_nodeOpMsg)
				thisNodeInfo := f_GetNodeInfo(c_nodeOpMsg)
				thisNodeInfo.Role = f_ChooseRole(thisNodeInfo, connectedNodes)
				f_SetNodeInfo(c_nodeOpMsg, thisNodeInfo)

				f_UpdateConnectedNodes(c_nodeOpMsg, f_GetNodeInfo(c_nodeOpMsg))

				//check for avalibale nodes
				var avalibaleNodes []T_NodeInfo
				updatedThisNode := false
				for i, nodeInfo := range connectedNodes {
					if nodeInfo.ElevatorInfo.State == elevator.IDLE {
						avalibaleNodes = append(avalibaleNodes, nodeInfo)
					}
					if thisNodeInfo.PRIORITY == nodeInfo.PRIORITY {
						if !updatedThisNode {
							connectedNodes[i] = thisNodeInfo
							f_SetConnectedNodes(c_nodeOpMsg, connectedNodes)
							updatedThisNode = true
						} else {
							F_WriteLog("Error, thisNode appended to CN more than one")
						}
						continue
					}
				}

				//check for first entry that is unassigned
				globalQueue := f_GetGlobalQueue(c_nodeOpMsg)
				for i, entry := range globalQueue {
					if (entry.Request.State == elevator.UNASSIGNED || entry.AssignedNode.PRIORITY == 0) && len(avalibaleNodes) > 0 { //OR for redundnacy, both should not be different in theory
						assignedEntry := F_AssignUnassignedRequest(entry, avalibaleNodes)
						globalQueue := f_GetGlobalQueue(c_nodeOpMsg)
						globalQueue[i] = assignedEntry
						f_SetGlobalQueue(c_nodeOpMsg, globalQueue)
						break
					}
				}
			}

		case SLAVE:

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
				thisNodeInfo.Role = f_ChooseRole(thisNodeInfo, connectedNodes)
				f_SetNodeInfo(c_nodeOpMsg, thisNodeInfo)

				f_UpdateConnectedNodes(c_nodeOpMsg, f_GetNodeInfo(c_nodeOpMsg))

			}
		}
	}
}
