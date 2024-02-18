package node

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
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

// ***	END TEST FUNCTIONS	***//
// nano .gitconfig
var thisNode T_Node

func F_Init(configName string) {
	logFile, err := os.OpenFile("log/debug.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
		return
	}
	defer logFile.Close()

	configFile, err := os.Open("config/" + configName + ".json")
	if err != nil {
		//kys eller prøv annen fil
		fmt.Println("Could not open file!")
		return
	}
	defer configFile.Close()
	var config T_Config
	json.NewDecoder(configFile).Decode(&config)

	// Init thisNode
	thisNodeInfo := &T_NodeInfo{
		PRIORITY: config.Priority,
		Role:     Master,
	}
	thisNode.Info = thisNodeInfo

	//thisNode.ELEVATOR = &elevator.T_Elevator{config.Priority, 0, Idle}
	PORT = config.Port
	IP = config.Ip
	FLOORS = config.Floors
}

func F_WriteLog(text string) bool {
	logFile, err := os.OpenFile("log/debug.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
		return false
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println(text)
	return true
}

// should contain the main master/slave fsm in Run() function, to be called from main
func F_Run() {
	f_TestCommunication(PORT)

	//run local elevator

	//to run the main FSM
	switch thisNode.Info.Role {
	case Master:
		//coms to receive orders
		//decide who will have the order
		//distribute the order
	case Slave:
		//
	}

}
