package node

import "the-elevator/elevator"

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

func F_AssignRequest(undistributedRequest T_GlobalQueueEntry, avalibaleNodes []T_NodeInfo) T_GlobalQueueEntry {

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
		Id:                undistributedRequest.Id,
		State:             elevator.ASSIGNED,
		Request:           undistributedRequest.Request,
		RequestedNode:     undistributedRequest.RequestedNode,
		AssignedNode:      chosenNode,
		TimeUntilReassign: REASSIGNTIME,
	}
	return distributedRequest
}
