package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"the-elevator/node"
	"time"
)

/*
Initializes global configuration and logging mechanisms. Opens "log/debug1.log" for logging and "config/default.json" for configuration settings,
then decodes the configuration into global variables. It's run automatically at the start to configure system settings like floors, port numbers,
and operational timings based on the JSON file. Errors during file operations or JSON decoding are output to the console.

Prerequisites: "config/default.json" must exist with valid configurations. The system must have write access to the log file directory.

Returns: None. Sets global variables for system configuration and initializes logging.
*/
func init() {
	logFile, errOpenLog := os.OpenFile("log/debug1.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	configFile, errOpenConfig := os.Open("config/default.json")
	var config node.T_Config
	errReadConfig := json.NewDecoder(configFile).Decode(&config)
	if errOpenLog != nil || errOpenConfig != nil || errReadConfig != nil {
		fmt.Println("Errors opening/reading config or log")
	}
	defer logFile.Close()
	defer configFile.Close()

	node.ThisNode = node.F_InitNode(config)

	node.FLOORS = config.Floors
	node.REASSIGN_PERIOD = config.ReassignPeriod
	node.CONNECTION_PERIOD = config.ConnectionPeriod
	node.SEND_PERIOD = config.SendPeriod
	node.GETSET_PERIOD = config.GetSetPeriod
	node.SLAVE_PORT = config.SlavePort
	node.MASTER_PORT = config.MasterPort
	node.ELEVATOR_PORT = config.ElevatorPort
	node.ASSIGN_BREAKOUT_PERIOD = config.AssignBreakoutPeriod
	node.MOST_RESPONSIVE_PERIOD = config.MostResponsivePeriod
	node.MEDIUM_RESPONSIVE_PERIOD = config.MiddleResponsivePeriod
	node.LEAST_RESPONSIVE_PERIOD = config.LeastResponsivePeriod
	node.TERMINATION_PERIOD = config.TerminationPeriod
	node.MAX_ALLOWED_ELEVATOR_ERRORS = config.MaxAllowedElevatorErrors
	node.MAX_ALLOWED_NODE_ERRORS = config.MaxAllowedNodeErrors
}

func main() {
	fmt.Println("Checking for primaries...")
	go node.F_NodeOperationManager(&node.ThisNode) //Should be only reference to ThisNode

	c_isPrimary := make(chan bool)
	c_shouldTerminate := make(chan bool)
	c_nodeRunningWithoutErrors := make(chan bool)
	c_elevatorRunningWithoutErrors := make(chan bool)
	go node.F_RunBackup(c_isPrimary)
	for {
		select {
		case <-c_isPrimary:
			exec.Command("gnome-terminal", "--", "go", "run", "main.go").Run()
			fmt.Println("Switched to primary")
			go node.F_RunPrimary(c_nodeRunningWithoutErrors, c_elevatorRunningWithoutErrors)
			go node.F_CheckIfShouldTerminate(c_shouldTerminate, c_nodeRunningWithoutErrors, c_elevatorRunningWithoutErrors)
		case <-c_shouldTerminate:
			fmt.Println("Terminating...")
			time.Sleep(time.Duration(1) * time.Second)
			os.Exit(1)
		}
	}
}