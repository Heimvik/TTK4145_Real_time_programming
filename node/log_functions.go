package node

import (
	"fmt"
	"log"
	"os"
	"the-elevator/node/elevator"
)

/*
Writes a given text string to a log file.

Prerequisites: File system access to create or append to the log file in the specified directory.

Returns: A boolean value, true, indicating successful logging of the provided text.
*/
func F_WriteLog(text string) bool {
	logFile, _ := os.OpenFile("log/debug1.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Print(text)
	return true
}

/*
Converts a node role type to a string ("MASTER" or "SLAVE").

Prerequisites: None.

Returns: "MASTER" or "SLAVE" based on the node role.
*/
func f_NodeRoleToString(role T_MSNodeRole) string {
	switch role {
	case MSROLE_MASTER:
		return "MASTER"
	default:
		return "SLAVE"
	}
}

/*
Logs the connected nodes' information, including priority, role, elevator status, and disconnect timeout.

Prerequisites: An up-to-date list of connected nodes.

Returns: Nothing, but records the detailed status of all connected nodes in the log file.
*/
func f_WriteLogConnectedNodes(connectedNodes []T_NodeInfo) {
	thisNode := f_GetNodeInfo()
	logStr := fmt.Sprintf("Node: | %d | %s | has connected nodes | ", thisNode.PRIORITY, f_NodeRoleToString(thisNode.MSRole))
	for _, info := range connectedNodes {
		logStr += fmt.Sprintf("%d (Role: %s, ElevatorInfo: %+v, TimeUntilDisconnect: %d) | ",
			info.PRIORITY, f_NodeRoleToString(info.MSRole), info.ElevatorInfo, info.TimeUntilDisconnect)
	}
	F_WriteLog(logStr)
}

/*
Generates and logs a summary of all connected nodes, including priority, role, elevator status, and disconnect timer.

Prerequisites: Updated list of connected nodes.

Returns: Nothing, but logs connected nodes' status; returns nothing.
*/
func f_WriteLogGlobalQueueEntry(entry T_GlobalQueueEntry) {
	logStr := fmt.Sprintf("Entry: | %d | State: | %s | Calltype: %s | Floor: %d | Direction: %s | Reassigned in: %.2f | ",
		entry.Request.Id, f_RequestStateToString(entry.Request.State), f_CallTypeToString(entry.Request.Calltype), entry.Request.Floor, f_DirectionToString(entry.Request.Direction), float64(entry.TimeUntilReassign))
	logStr += fmt.Sprintf("Requested node: | %d | ",
		entry.RequestedNode)
	logStr += fmt.Sprintf("Assigned node: | %d | ",
		entry.AssignedNode)
	F_WriteLog(logStr)
}

/*
Translates an elevator call type to a string representation.

Prerequisites: A valid elevator call type.

Returns: "NONE", "CAB", "HALL", or "UNKNOWN" based on the call type.
*/
func f_CallTypeToString(callType elevator.T_CallType) string {
	switch callType {
	case 0:
		return "NONE"
	case 1:
		return "CAB"
	case 2:
		return "HALL"
	default:
		return "UNKNOWN"
	}
}

/*
Converts an elevator request state to its string equivalent.

Prerequisites: A valid elevator request state.

Returns: "UNASSIGNED", "ASSIGNED", "ACTIVE", "DONE", or "UNKNOWN" based on the state.
*/
func f_RequestStateToString(state elevator.T_RequestState) string {
	switch state {
	case elevator.REQUESTSTATE_UNASSIGNED:
		return "UNASSIGNED"
	case elevator.REQUESTSTATE_ASSIGNED:
		return "ASSIGNED"
	case elevator.REQUESTSTATE_ACTIVE:
		return "ACTIVE"
	case elevator.REQUESTSTATE_DONE:
		return "DONE"
	default:
		return "UNKNOWN"
	}
}

/*
Translates an elevator direction value into a string representation.

Prerequisites: A valid elevator direction value.

Returns: "UP", "DOWN", "NONE", or "UNKNOWN" based on the direction.
*/
func f_DirectionToString(direction elevator.T_ElevatorDirection) string {
	switch direction {
	case 1:
		return "UP"
	case -1:
		return "DOWN"
	case 0:
		return "NONE"
	default:
		return "UNKNOWN"
	}
}

/*
Logs detailed information about a received slave message, including request details and sender node, to track system communication and request distribution.

Prerequisites: A valid slave message to log.

Returns: Nothing, but logs detailed slave message information.
*/
func f_WriteLogSlaveMessage(slaveMessage T_SlaveMessage) {
	entryStr := fmt.Sprintf("Entry: | %d | State: | %s | Calltype: %s | Floor: %d | Direction: %s | Reassigned in: %.2f | ",
		slaveMessage.Entry.Request.Id, f_RequestStateToString(slaveMessage.Entry.Request.State), f_CallTypeToString(slaveMessage.Entry.Request.Calltype), slaveMessage.Entry.Request.Floor, f_DirectionToString(slaveMessage.Entry.Request.Direction), float64(slaveMessage.Entry.TimeUntilReassign))
	entryStr += fmt.Sprintf("Requested node: | %d | ",
		slaveMessage.Entry.RequestedNode)
	entryStr += fmt.Sprintf("Assigned node: | %d | ",
		slaveMessage.Entry.AssignedNode)
	thisNode := f_GetNodeInfo()
	logStr := fmt.Sprintf("Node: | %d | %s | received SM from | %d | Entry: %s",
		int(thisNode.PRIORITY), f_NodeRoleToString(thisNode.MSRole), int(slaveMessage.Transmitter.PRIORITY), entryStr)
	F_WriteLog(logStr)
}

/*
Logs a received master message, documenting the sender's details and the global queue status, aiding in the synchronization and debugging of the network's state.

Prerequisites: A valid master message to log.

Returns: Nothing, but logs detailed master message and global queue information.
*/
func f_WriteLogMasterMessage(masterMessage T_MasterMessage) {
	thisNode := f_GetNodeInfo()
	roleStr := f_NodeRoleToString(thisNode.MSRole)
	transmitterRoleStr := f_NodeRoleToString(masterMessage.Transmitter.MSRole)

	logStr := fmt.Sprintf("Node: | %d | %s | received MM from | %d | %s | GlobalQueue: [",
		thisNode.PRIORITY, roleStr, masterMessage.Transmitter.PRIORITY, transmitterRoleStr)

	for i, entry := range masterMessage.GlobalQueue {
		entryStr := fmt.Sprintf("Request ID: | %d | State: | %s | Calltype: %s | Floor: %d | Direction: %s | Reassigned in: %d | Requested node: | %d | Assigned node: | %d |",
			entry.Request.Id, f_RequestStateToString(entry.Request.State), f_CallTypeToString(entry.Request.Calltype), int(entry.Request.Floor),
			f_DirectionToString(entry.Request.Direction), int(entry.TimeUntilReassign),
			int(entry.RequestedNode), int(entry.AssignedNode))

		logStr += entryStr
		if i < len(masterMessage.GlobalQueue)-1 {
			logStr += ", "
		}
	}
	logStr += "]"

	F_WriteLog(logStr)
}
