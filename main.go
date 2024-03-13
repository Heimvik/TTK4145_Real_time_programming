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
Initializes the log with filename and overwrites any existing content.

Prerequisites: None

Returns: None
*/
func f_StartLog() {
	logFile, err := os.OpenFile("log/debug1.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("Error opening/creating log file:", err)
		return
	}
	defer logFile.Close()
}

/*
Restarts the log every *parameter* minutes to avoid unnecessary memory consumption

Prerequisites: None

Returns: None, but resets log after given minutes.
*/
func f_RestartLogEachMinutes(minutes int) {
	for range time.Tick(time.Duration(minutes) * time.Minute) {
		f_StartLog()
	}
}

/*
Initializes global configuration and logging mechanisms.

Prerequisites: "config/default.json" must exist with valid configurations. The system must have write access to the log file directory.

Returns: None. Sets global variables for system configuration and initializes logging.
*/
func init() {
	f_StartLog()
	go f_RestartLogEachMinutes(5)

	configFile, errOpenConfig := os.Open("config/default.json")
	var config node.T_Config
	errReadConfig := json.NewDecoder(configFile).Decode(&config)
	if errOpenConfig != nil || errReadConfig != nil {
		fmt.Println("Errors opening/reading config or log")
	}
	defer configFile.Close()

	node.ThisNode = node.F_InitNode(config)

	node.FLOORS = config.Floors
	node.REASSIGN_PERIOD = config.ReassignPeriod
	node.CONNECTION_PERIOD = config.ConnectionPeriod
	node.IMMOBILE_PERIOD = config.ImmobilePeriod
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
