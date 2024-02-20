package node

import (
	"strconv"
	"the-elevator/elevator"
)

//"the-elevator/elevator"

//Should take in:
//The request struct containing:
// - cab/hall
// - elevator id
// - floor
// - diraction
//
//Connected elevators and their struct:
// - elevator id
// - state (idle, moving, dooropen)
// - floor
// - diraction

// Should give out the elevator that should serve the request
func f_AbsInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func f_ClosestElevatorNode(floor int, nodes []T_NodeInfo) T_NodeInfo {
	var closestNode T_NodeInfo
	closestFloor := FLOORS
	for _, nodeInfo := range nodes {
		currentDifference := f_AbsInt(nodeInfo.ElevatorInfo.Floor - floor)
		if currentDifference < closestFloor {
			closestFloor = nodeInfo.ElevatorInfo.Floor
			closestNode = nodeInfo
		}
	}
	return closestNode
}

func F_AssignUnassignedRequest(undistributedRequest T_GlobalQueueEntry, avalibaleNodes []T_NodeInfo) T_GlobalQueueEntry {

	var distributedRequest T_GlobalQueueEntry
	var chosenNode T_NodeInfo
	switch undistributedRequest.Request.Calltype {
	case elevator.HALL:
		chosenNode = f_ClosestElevatorNode(undistributedRequest.Request.Floor, avalibaleNodes)
	case elevator.CAB:
		if undistributedRequest.RequestedNode.ElevatorInfo.State == elevator.IDLE {
			chosenNode = undistributedRequest.RequestedNode
		} else {
			chosenNode = f_ClosestElevatorNode(undistributedRequest.Request.Floor, avalibaleNodes)
		}
	}

	distributedRequest = T_GlobalQueueEntry{
		Request:           undistributedRequest.Request,
		RequestedNode:     undistributedRequest.RequestedNode,
		AssignedNode:      chosenNode,
		TimeUntilReassign: REASSIGNTIME,
	}
	return distributedRequest
}

func F_FindAssignedRequest(globalQueue []T_GlobalQueueEntry, thisNodeInfo T_NodeInfo) elevator.T_Request {
	var returnRequest elevator.T_Request
	for _, entry := range globalQueue {
		if entry.AssignedNode.PRIORITY == thisNodeInfo.PRIORITY && entry.Request.State == elevator.ASSIGNED {
			F_WriteLog("Entry with ID: " + strconv.Itoa(entry.Request.Id) + " from " + strconv.Itoa(entry.RequestedNode.PRIORITY))
			returnRequest = entry.Request
		} else {
			returnRequest = elevator.T_Request{}
		}
	}
	return returnRequest
}
