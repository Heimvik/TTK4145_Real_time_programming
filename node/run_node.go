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
func f_TimerWatchdog(c_nodeOpMsg chan T_NodeOperationMessage, c_reassignEntry chan T_GlobalQueueEntry) {
	globalQueue := f_GetGlobalQueue(c_nodeOpMsg)
	for _, element := range globalQueue {
		element.TimeUntilReassign -= 1
		if element.TimeUntilReassign == 0 && element.State != DONE {
			c_reassignEntry <- element
		}
		time.Sleep(1 * time.Second)
	}
}

func f_NodeOperationManager(node *T_Node, c_nodeReadMsg chan T_NodeOperationMessage, c_nodeWriteMsg chan T_NodeOperationMessage) {
	//IMPORTANT: ONLY IN THIS GOROUTINE CAN GLOBAL NODE VARIABLES BE USED, rest has do be channel dependent
	for {
		select {
		case readOperation := <-c_nodeReadMsg:
			switch readOperation.Type {
			case ReadNodeInfo:
				readOperation.Result <- node.Info
			case ReadGlobalQueue:
				readOperation.Result <- node.GlobalQueue
			case ReadConnectedNodes:
				readOperation.Result <- node.ConnectedNodes
			case ReadElevator:
				readOperation.Result <- node.P_ELEVATOR
			}
		case writeOperation := <-c_nodeWriteMsg:
			switch writeOperation.Type {
			case WriteNodeInfo:
				node.Info = writeOperation.Data.(T_NodeInfo)
			case WriteGlobalQueue:
				node.GlobalQueue = writeOperation.Data.([]T_GlobalQueueEntry)
			case WriteConnectedNodes:
				node.ConnectedNodes = writeOperation.Data.([]T_NodeInfo)
			}
		}
	}
}

func f_GetNodeInfo(c_nodeOpMsg chan T_NodeOperationMessage) T_NodeInfo {
	c_nodeInfo := make(chan interface{})
	c_nodeOpMsg <- T_NodeOperationMessage{Type: ReadNodeInfo, Result: c_nodeInfo}
	nodeInfo := <-c_nodeInfo
	return nodeInfo.(T_NodeInfo)
}
func f_SetNodeInfo(c_nodeOpMsg chan T_NodeOperationMessage, thisNodeInfo T_NodeInfo) {
	c_nodeOpMsg <- T_NodeOperationMessage{Type: WriteNodeInfo, Data: thisNodeInfo}
}
func f_GetGlobalQueue(c_nodeOpMsg chan T_NodeOperationMessage) []T_GlobalQueueEntry {
	c_globalQueue := make(chan interface{})
	c_nodeOpMsg <- T_NodeOperationMessage{Type: ReadGlobalQueue, Result: c_globalQueue}
	globalQueueResult := <-c_globalQueue
	return globalQueueResult.([]T_GlobalQueueEntry)
}
func f_SetGlobalQueue(c_nodeOpMsg chan T_NodeOperationMessage, globalQueue []T_GlobalQueueEntry) {
	c_nodeOpMsg <- T_NodeOperationMessage{Type: WriteGlobalQueue, Data: globalQueue}
}
func f_AddEntryGlobalQueue(c_nodeReadMsg chan T_NodeOperationMessage, c_nodeWriteMsg chan T_NodeOperationMessage, entryToAdd T_GlobalQueueEntry) {
	thisGlobalQueue := f_GetGlobalQueue(c_nodeReadMsg)
	entryIsUnique := true
	for _, entry := range thisGlobalQueue {
		if entryToAdd.Id == entry.Id { //random id generated to each entry
			entryIsUnique = false
		}
	}
	if entryIsUnique {
		entryToAdd.AssignedNode.PRIORITY = 0
		entryToAdd.State = UNASSIGNED
		thisGlobalQueue = append(thisGlobalQueue, entryToAdd)
		f_SetGlobalQueue(c_nodeWriteMsg, thisGlobalQueue)
	} else {
		F_WriteLog("Discarded request " + strconv.Itoa(entryToAdd.Id))
	}
}
func f_GetConnectedNodes(c_nodeOpMsg chan T_NodeOperationMessage) []T_NodeInfo {
	c_connectedNodes := make(chan interface{})
	c_nodeOpMsg <- T_NodeOperationMessage{Type: ReadConnectedNodes, Result: c_connectedNodes}
	connectedNodesResult := <-c_connectedNodes
	return connectedNodesResult.([]T_NodeInfo)
}
func f_SetConnectedNodes(c_nodeOpMsg chan T_NodeOperationMessage, connectedNodes []T_NodeInfo) {
	c_nodeOpMsg <- T_NodeOperationMessage{Type: WriteConnectedNodes, Data: connectedNodes}
}

// should contain the main master/slave fsm in Run() function, to be called from main
func F_RunNode() {
	//to run the main FSM
	c_nodeReadMsg := make(chan T_NodeOperationMessage)
	c_nodeWriteMsg := make(chan T_NodeOperationMessage)
	go f_NodeOperationManager(&ThisNode, c_nodeReadMsg, c_nodeWriteMsg) //SHOULD BE THE ONLY REFERENCE TO ThisNode!
	for {
		nodeRole := f_GetNodeInfo(c_nodeReadMsg).Role
		switch nodeRole {
		case MASTER:

			go F_TransmitMasterMessage(c_nodeReadMsg, MASTERPORT)

			//Receive messages
			c_slaveMessages := make(chan T_SlaveMessage)
			c_masterMessages := make(chan T_MasterMessage)

			c_newConnectedNodes := make(chan []T_NodeInfo)
			c_oldConnectedNodes := make(chan []T_NodeInfo)

			c_reassignEntry := make(chan T_GlobalQueueEntry)

			go F_ReceiveSlaveMessage(c_slaveMessages, c_oldConnectedNodes, c_newConnectedNodes, SLAVEPORT)
			go F_ReceiveMasterMessage(c_masterMessages, c_oldConnectedNodes, c_newConnectedNodes, MASTERPORT)
			go f_TimerWatchdog(c_nodeReadMsg, c_reassignEntry)
			go func() {
				for {
					c_oldConnectedNodes <- f_GetConnectedNodes(c_nodeReadMsg)
					select {
					case newConnectedNodes := <-c_newConnectedNodes:
						f_SetConnectedNodes(c_nodeWriteMsg, newConnectedNodes)
						f_WriteLogConnectedNodes(f_GetConnectedNodes(c_nodeReadMsg))

					case masterMessage := <-c_masterMessages:
						for _, remoteEntry := range masterMessage.GlobalQueue {
							f_AddEntryGlobalQueue(c_nodeReadMsg, c_nodeWriteMsg, remoteEntry)
						}
						//IMPORTANT: cannot really propagate to slave until it knows that the other master has received its GQ

						connectedNodes := f_GetConnectedNodes(c_nodeReadMsg)
						thisNodeInfo := f_GetNodeInfo(c_nodeReadMsg)
						thisNodeInfo.Role = f_ChooseRole(thisNodeInfo, connectedNodes)
						f_SetNodeInfo(c_nodeWriteMsg, thisNodeInfo)

					case slaveMessage := <-c_slaveMessages:
						f_AddEntryGlobalQueue(c_nodeReadMsg, c_nodeWriteMsg, slaveMessage.Entry)

						//Update elevator whereabout of the slave
						connectedNodes := f_GetConnectedNodes(c_nodeReadMsg)
						for i, nodeInfo := range connectedNodes {
							if slaveMessage.Transmitter.PRIORITY == nodeInfo.PRIORITY {
								connectedNodes[i] = slaveMessage.Transmitter
								f_SetConnectedNodes(c_nodeWriteMsg, connectedNodes)
							}
						}
					case reassignEntry := <-c_reassignEntry:
						f_AddEntryGlobalQueue(c_nodeReadMsg, c_nodeWriteMsg, reassignEntry) //Demands that the old one is removed
					}
				}
			}()

			//Calculate and assign elements in GQ
			go func() {
				for {
					//check for avalibale nodes
					var avalibaleNodes []T_NodeInfo
					connectedNodes := f_GetConnectedNodes(c_nodeReadMsg)
					for _, nodeInfo := range connectedNodes {
						if nodeInfo.ElevatorInfo.State == elevator.IDLE {
							avalibaleNodes = append(avalibaleNodes, nodeInfo)
						}
					}

					//check for first entry that is unassigned
					globalQueue := f_GetGlobalQueue(c_nodeReadMsg)
					for i, entry := range globalQueue {
						if (entry.State == UNASSIGNED || entry.AssignedNode.PRIORITY == 0) && len(avalibaleNodes) > 0 { //OR for redundnacy, both should not be different in theory
							assignedEntry := F_AssignRequest(entry, avalibaleNodes)
							globalQueue := f_GetGlobalQueue(c_nodeReadMsg)
							globalQueue[i] = assignedEntry
							f_SetGlobalQueue(c_nodeWriteMsg, globalQueue)
							break
						}
					}

					time.Sleep(1 * time.Second)
					//Programmet på få tid til å resolve om en node er connected eller ikke før det assignes en ny
				}
			}()

			//Send MasterMessages
			go func() {
				for {
					break
				}
			}()

			for {
				//Update own elevator information
				connectedNodes := f_GetConnectedNodes(c_nodeReadMsg)
				thisNodeInfo := f_GetNodeInfo(c_nodeReadMsg)
				for i, nodeInfo := range connectedNodes {
					if thisNodeInfo.PRIORITY == nodeInfo.PRIORITY {
						connectedNodes[i] = thisNodeInfo
						f_SetConnectedNodes(c_nodeWriteMsg, connectedNodes)
						break
					}
				}
			}

		case SLAVE:
			//receive MasterMessage

		}
	}
}
