package node

import (
	"fmt"
	"log"
	"os"
	"the-elevator/node/elevator"
)

func F_WriteLog(text string) bool {
	logFile, _ := os.OpenFile("log/debug1.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Print(text)
	return true
}
func f_NodeRoleToString(role T_MSNodeRole) string {
	switch role {
	case MASTER:
		return "MASTER"
	default:
		return "SLAVE"
	}
}
func f_WriteLogConnectedNodes(connectedNodes []T_NodeInfo) {
	thisNode := f_GetNodeInfo()
	logStr := fmt.Sprintf("Node: | %d | %s | has connected nodes | ", thisNode.PRIORITY, f_NodeRoleToString(thisNode.MSRole))
	for _, info := range connectedNodes {
		logStr += fmt.Sprintf("%d (Role: %s, ElevatorInfo: %+v, TimeUntilDisconnect: %d) | ",
			info.PRIORITY, f_NodeRoleToString(info.MSRole), info.ElevatorInfo, info.TimeUntilDisconnect)
	}
	F_WriteLog(logStr)
}
func F_WriteLogGlobalQueueEntry(entry T_GlobalQueueEntry) {
	logStr := fmt.Sprintf("Entry: | %d | State: | %s | Calltype: %s | Floor: %d | Direction: %s | Reassigned in: %.2f | ",
		entry.Request.Id, f_RequestStateToString(entry.Request.State), f_CallTypeToString(entry.Request.Calltype), entry.Request.Floor, f_DirectionToString(entry.Request.Direction), float64(entry.TimeUntilReassign))
	logStr += fmt.Sprintf("Requested node: | %d | ",
		entry.RequestedNode)
	logStr += fmt.Sprintf("Assigned node: | %d | ",
		entry.AssignedNode)
	F_WriteLog(logStr)
}
func f_CallTypeToString(callType elevator.T_Call) string {
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
func f_RequestStateToString(state elevator.T_RequestState) string {
	switch state {
	case elevator.UNASSIGNED:
		return "UNASSIGNED"
	case elevator.ASSIGNED:
		return "ASSIGNED"
	case elevator.ACTIVE:
		return "ACTIVE"
	case elevator.DONE:
		return "DONE"
	default:
		return "UNKNOWN"
	}
}

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
func f_WriteLogSlaveMessage(slaveMessage T_SlaveMessage) {
	request := slaveMessage.Entry.Request
	thisNode := f_GetNodeInfo()
	logStr := fmt.Sprintf("Node: | %d | %s | received SM from | %d | Request ID: | %d | State: | %s | Calltype: %s | Floor: %d | Direction: %s |",
		int(thisNode.PRIORITY), f_NodeRoleToString(thisNode.MSRole), int(slaveMessage.Transmitter.PRIORITY), int(request.Id), f_RequestStateToString(slaveMessage.Entry.Request.State), f_CallTypeToString(request.Calltype), int(request.Floor), f_DirectionToString(request.Direction))
	F_WriteLog(logStr)
}
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
