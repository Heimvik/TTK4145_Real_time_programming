package node

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

func f_TestDistribution() {
	var elevators []*T_Elevator
	for i := 0; i <= 2; i++ {
		newElevator := T_Elevator{
			Floor:     4 - i,
			Direction: Down,
			Avalibale: true,
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

func f_TestCommunication(port int) {
	// Setup
	c_transmitMessage := make(chan T_Message)
	c_receivedMessage := make(chan T_Message)
	c_connectedNodes := make(chan []*T_Node)

	F_ReceiveMessages(c_receivedMessage, c_connectedNodes, port)
	F_TransmitMessages(c_transmitMessage, port)

	go func() {
		helloMsg := T_Message{thisNode, "Hello from " + strconv.Itoa(thisNode.PRIORITY)}
		for {
			c_transmitMessage <- helloMsg
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		received := <-c_receivedMessage
		fmt.Println(received.TestStr)
		connectedNodes := <-c_connectedNodes
		for _, node := range connectedNodes {
			fmt.Println("Connected nodes: " + strconv.Itoa(node.PRIORITY))
		}
		time.Sleep(1 * time.Second)
	}()

}

//***	END TEST FUNCTIONS	***//

var thisNode T_Node

func F_Init(initType string) {
	configFile, err := os.Open("../config/" + initType + ".json")
	if err != nil {
		//kys eller prøv annen fil
		fmt.Println("Could not open file!")
		return
	}
	defer configFile.Close()

	// Decode the JSON data into the Config struct
	var config T_Config
	json.NewDecoder(configFile).Decode(&config)

	// Init thisNode
	thisNode.PRIORITY = config.Priority
	thisNode.ELEVATOR = &T_Elevator{config.Priority, 0, 2, Idle, true}
	PORT = config.Port
	IP = config.Ip
	FLOORS = config.Floors
}

// should contain the main master/slave fsm in Run() function, to be called from main
func F_Run() {
	f_TestCommunication(PORT)
}
