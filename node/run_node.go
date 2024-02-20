package node

import (
	"encoding/json"
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
	logFile, _ := os.OpenFile("log/debug.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
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
	MMMILLS = config.MMMills
	SLAVEPORT = config.SlavePort
	MASTERPORT = config.MasterPort
	ELEVATORPORT = config.ElevatorPort
}

func F_WriteLog(text string) bool {
	logFile, _ := os.OpenFile("log/debug.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println(text)
	return true
}
func f_WriteLogConnectedNodes(connectedNodes []T_NodeInfo) {
	logStr := "Updated connected nodes: "
	for _, nodeInfo := range ThisNode.ConnectedNodes {
		logStr += strconv.Itoa(nodeInfo.PRIORITY) + "\t"
	}
	F_WriteLog(logStr)
}
func f_WriteLogSlaveMessage(slaveMessage T_SlaveMessage) {
	request := slaveMessage.Entry.Request
	logStr := "Slavemessage from: " + strconv.Itoa(slaveMessage.Transmitter.PRIORITY) + "\t Content:"
	if request.Calltype == 0 {
		logStr += "CAB\t"
	} else {
		logStr += "HALL\t"
	}
	logStr += strconv.Itoa(request.Floor) + "\t"
	if request.Direction == 0 {
		logStr += "UP\t"
	} else if request.Direction == 1 {
		logStr += "DOWN\t"
	} else {
		logStr += "NONE"
	}
	F_WriteLog(logStr)
}
func f_WriteLogMasterMessage(masterMessage T_MasterMessage) {
	logStr := "MasterMessage from: " + strconv.Itoa(masterMessage.Transmitter.PRIORITY) + "\t as "
	if masterMessage.Transmitter.Role == MASTER {
		logStr += "MASTER\t"
	} else {
		logStr += "SLAVE\t"
	}
	//add more about GQ
	F_WriteLog(logStr)
}
func f_ChooseRole(thisNodeInfo T_NodeInfo, connectedNodes []T_NodeInfo) T_NodeRole {
	var returnRole T_NodeRole
	for _, remoteNodeInfo := range connectedNodes {
		if remoteNodeInfo.PRIORITY > thisNodeInfo.PRIORITY {
			returnRole = SLAVE
		} else {
			returnRole = MASTER
		}
	}
	return returnRole
}

// increments all timers in GQ and checks for any that has run out
func f_TimerWatchdog(ops T_NodeOperations, c_reassignEntry chan T_GlobalQueueEntry, c_quit chan bool) {
	for {
		globalQueue := f_GetGlobalQueue(ops)
		for _, element := range globalQueue {
			element.TimeUntilReassign -= 1
			if element.TimeUntilReassign == 0 && element.Request.State != elevator.DONE {
				c_reassignEntry <- element
			}
			time.Sleep(1 * time.Second)
		}
		select {
		case <-c_quit:
			F_WriteLog("Closed Watchdog goroutine in master")
			return
		default:
			continue
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

func f_AddEntryGlobalQueue(operations T_NodeOperations, entryToAdd T_GlobalQueueEntry) {
	thisGlobalQueue := f_GetGlobalQueue(operations)
	entryIsUnique := true
	var entryIndex int
	for i, entry := range thisGlobalQueue {
		if entryToAdd.Request.Id == entry.Request.Id && entryToAdd.RequestedNode == entry.RequestedNode { //random id generated to each entry
			entryIsUnique = false
			entryIndex = i
		}
	}
	if entryIsUnique {
		entryToAdd.AssignedNode.PRIORITY = 0
		thisGlobalQueue = append(thisGlobalQueue, entryToAdd)
		f_SetGlobalQueue(operations, thisGlobalQueue)
	} else { //should update the existing entry
		thisGlobalQueue[entryIndex] = entryToAdd
	}
}
func f_HandleElevator(ops T_NodeOperations, c_requestFromElevator chan elevator.T_Request, c_requestToElevator chan elevator.T_Request, c_quit chan bool) {
	for {
		globalQueue := f_GetGlobalQueue(ops)
		thisNodeInfo := f_GetNodeInfo(ops)
		transmitRequest := F_FindAssignedRequest(globalQueue, thisNodeInfo)
		if transmitRequest != (elevator.T_Request{}) {
			c_requestToElevator <- transmitRequest
		}
		select {
		case receivedRequest := <-c_requestFromElevator:
			//make a globalqueueentry, and add to globalqueue
			entry := T_GlobalQueueEntry{
				Request:           receivedRequest,
				RequestedNode:     f_GetNodeInfo(ops),
				AssignedNode:      T_NodeInfo{},
				TimeUntilReassign: REASSIGNTIME,
			}
			f_AddEntryGlobalQueue(ops, entry)
		case <-c_quit:
			F_WriteLog("Closed HandleElevator goroutine in master")
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
	c_newConnectedNodes := make(chan []T_NodeInfo)
	c_quit := make(chan bool)

	go f_NodeOperationManager(&ThisNode, c_nodeOpMsg) //SHOULD BE THE ONLY REFERENCE TO ThisNode!
	go F_ReceiveSlaveMessage(c_receiveSlaveMessage, c_nodeOpMsg, c_newConnectedNodes, SLAVEPORT)
	go F_ReceiveMasterMessage(c_receiveMasterMessage, c_nodeOpMsg, c_newConnectedNodes, MASTERPORT)
	go F_TransmitSlaveMessage(c_transmitSlaveMessage, SLAVEPORT)
	go F_TransmitMasterMessage(c_transmitMasterMessage, MASTERPORT)
	for {
		nodeRole := f_GetNodeInfo(c_nodeOpMsg).Role
		switch nodeRole {
		case MASTER:
			//Receive messages

			c_reassignEntry := make(chan T_GlobalQueueEntry)
			//Watchdog goroutine
			go f_TimerWatchdog(c_nodeOpMsg, c_reassignEntry, c_quit)
			//Message goroutine
			go func() {
				for {
					select {
					case newConnectedNodes := <-c_newConnectedNodes:
						f_SetConnectedNodes(c_nodeOpMsg, newConnectedNodes)
						f_WriteLogConnectedNodes(f_GetConnectedNodes(c_nodeOpMsg))

					case masterMessage := <-c_receiveMasterMessage:
						f_WriteLogMasterMessage(masterMessage)
						for _, remoteEntry := range masterMessage.GlobalQueue {
							f_AddEntryGlobalQueue(c_nodeOpMsg, remoteEntry)
						}
						//IMPORTANT: cannot really propagate to slave until it knows that the other master has received its GQ

						connectedNodes := f_GetConnectedNodes(c_nodeOpMsg)
						thisNodeInfo := f_GetNodeInfo(c_nodeOpMsg)
						thisNodeInfo.Role = f_ChooseRole(thisNodeInfo, connectedNodes)
						if thisNodeInfo.Role == SLAVE {
							close(c_quit) //closes all the master goroutines
						}
						f_SetNodeInfo(c_nodeOpMsg, thisNodeInfo)

					case slaveMessage := <-c_receiveSlaveMessage:
						f_WriteLogSlaveMessage(slaveMessage)
						f_AddEntryGlobalQueue(c_nodeOpMsg, slaveMessage.Entry)

						//Update elevator whereabout of the slave
						connectedNodes := f_GetConnectedNodes(c_nodeOpMsg)
						for i, nodeInfo := range connectedNodes {
							if slaveMessage.Transmitter.PRIORITY == nodeInfo.PRIORITY {
								connectedNodes[i] = slaveMessage.Transmitter
								f_SetConnectedNodes(c_nodeOpMsg, connectedNodes)
							}
						}
					case reassignEntry := <-c_reassignEntry:
						f_AddEntryGlobalQueue(c_nodeOpMsg, reassignEntry) //Demands that the old one is removed
					case <-c_quit:
						F_WriteLog("Closed Receive Message goroutine in master")
						return
					}
				}
			}()

			//Distribution goroutine
			go func() {
				for {
					//check for avalibale nodes
					var avalibaleNodes []T_NodeInfo
					connectedNodes := f_GetConnectedNodes(c_nodeOpMsg)
					for _, nodeInfo := range connectedNodes {
						if nodeInfo.ElevatorInfo.State == elevator.IDLE {
							avalibaleNodes = append(avalibaleNodes, nodeInfo)
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
					select {
					case <-c_quit:
						F_WriteLog("Closed Distribution goroutine in master")
						return
					default:
						time.Sleep(1 * time.Second)
					}
				}
			}()

			//Send MasterMessages
			go func() {
				for {
					masterMessage := T_MasterMessage{
						Transmitter: f_GetNodeInfo(c_nodeOpMsg),
						GlobalQueue: f_GetGlobalQueue(c_nodeOpMsg),
					}
					c_transmitMasterMessage <- masterMessage
					F_WriteLog("MasterMessage sent on port: " + strconv.Itoa(MASTERPORT))
					time.Sleep(time.Duration(MMMILLS) * time.Millisecond)
					select {
					case <-c_quit:
						F_WriteLog("Closed Transmit Message goroutine in master")
						return
					default:
						continue
					}
				}
			}()

			c_transmitElevatorRequest := make(chan elevator.T_Request)
			c_receiveElevatorRequest := make(chan elevator.T_Request)

			//go elevator.F_RunElevator(c_transmitElevatorRequest, c_receiveElevatorRequest)
			go f_HandleElevator(c_nodeOpMsg, c_receiveElevatorRequest, c_transmitElevatorRequest,c_quit) //MAKE ABLE TO QUIT

			for { //MAKE ABLE TO QUIT

				//Update own elevator information
				connectedNodes := f_GetConnectedNodes(c_nodeOpMsg)
				thisNodeInfo := f_GetNodeInfo(c_nodeOpMsg)
				//logStr := "Connected nodes: "
				for i, nodeInfo := range connectedNodes {
					//logStr += strconv.Itoa(connectedNodes[i].PRIORITY)
					if thisNodeInfo.PRIORITY == nodeInfo.PRIORITY {
						connectedNodes[i] = thisNodeInfo
						f_SetConnectedNodes(c_nodeOpMsg, connectedNodes)
						break
					}
				}
				select {
				case <-c_quit:
					F_WriteLog("Closed Miscellaneous goroutine in master")
					return
				default:
					time.Sleep(1 * time.Second)
				}

			}

		case SLAVE:

			//receive MasterMessage
			go func() { //MAKE ABLE TO QUIT
				for {

				}
			}()

		}
	}
}
