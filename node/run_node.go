package node

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"the-elevator/elevator"
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
	p_thisNodeInfo := &T_NodeInfo{
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
		P_info:     p_thisNodeInfo,
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

	PORT = config.Port
	IP = config.Ip
	FLOORS = config.Floors
	REASSIGNTIME = config.ReassignTime
}

func F_WriteLog(text string) bool {
	logFile, _ := os.OpenFile("log/debug.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println(text)
	return true
}
func f_WriteLogConnectedNodes(connectedNodes []*T_NodeInfo) {
	logStr := "Updated connected nodes: "
	for _, nodeInfo := range ThisNode.ConnectedNodes {
		logStr += strconv.Itoa(nodeInfo.PRIORITY) + "\t"
	}
	F_WriteLog(logStr)
}
func f_WriteLogSlaveMessage(slaveMessage T_SlaveMessage) {
	request := slaveMessage.RequestInfo.Request
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

// should contain the main master/slave fsm in Run() function, to be called from main
func F_RunNode() {

	//to run the main FSM
	switch ThisNode.P_info.Role {
	case MASTER:
		//receive SlaveMessage
		c_slaveMessages := make(chan T_SlaveMessage)
		c_newConnectedNodes := make(chan []*T_NodeInfo)
		c_oldConnectedNodes := make(chan []*T_NodeInfo)
		go F_ReceiveSlaveMessage(c_slaveMessages, c_oldConnectedNodes, c_newConnectedNodes, PORT)
		go func() {
			for {
				c_oldConnectedNodes <- ThisNode.ConnectedNodes
				select {
				case ThisNode.ConnectedNodes = <-c_newConnectedNodes:
					f_WriteLogConnectedNodes(ThisNode.ConnectedNodes)
				case slaveMessage := <-c_slaveMessages:
					f_WriteLogSlaveMessage(slaveMessage)

					//Update global queue
					entryIsUnique := true
					for _, requestInfo := range ThisNode.GlobalQueue {
						if slaveMessage.RequestInfo.Id == requestInfo.Id { //random id generated
							entryIsUnique = false
						}
					}
					if entryIsUnique {
						ThisNode.GlobalQueue = append(ThisNode.GlobalQueue, slaveMessage.RequestInfo)
					}else{
						F_WriteLog("Discarded request "+strconv.Itoa(slaveMessage.RequestInfo.Id))
					}

					//Update thisnode.connectedNodes with state of the elevator it received messages from
					for _,p_node

				}
			}
		}()
		//add entry to global queue
		ThisNode.GlobalQueue
		//calculate distribution of lowest order

		//update enrty in global queue

		//send MasterMessage

	case SLAVE:
		//receive MasterMessage
		var masterMessage T_MasterMessage
		c_newConnectedNodes := make(chan []*T_NodeInfo)
		c_oldConnectedNodes := make(chan []*T_NodeInfo)

		go func() {
			masterMessage.I_Receive(c_oldConnectedNodes, c_newConnectedNodes, PORT)
			for {
				c_oldConnectedNodes <- ThisNode.ConnectedNodes
				ThisNode.ConnectedNodes = <-c_newConnectedNodes
			}
		}()
	}

}
