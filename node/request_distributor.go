package node

import (
	"strconv"
	"the-elevator/node/elevator"
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
func f_AbsInt(x int8) int8 {
	if x < 0 {
		return -x
	}
	return x
}

func f_ClosestElevatorNode(floor int8, nodes []T_NodeInfo) uint8 {
	var closestNode T_NodeInfo
	closestFloor := FLOORS
	for _, nodeInfo := range nodes {
		currentDifference := f_AbsInt(nodeInfo.ElevatorInfo.Floor - floor)
		if currentDifference < closestFloor {
			closestFloor = nodeInfo.ElevatorInfo.Floor
			closestNode = nodeInfo
		}
		if currentDifference > FLOORS {
			F_WriteLog("Error: Found floordifference larger than max floors")
		}
	}
	return closestNode.PRIORITY
}

func F_AssignNewEntry(globalQueue []T_GlobalQueueEntry, connectedNodes []T_NodeInfo, avalibaleNodes []T_NodeInfo) (T_GlobalQueueEntry, int) {
	assignedEntry := T_GlobalQueueEntry{}
	assignedEntryIndex := -1
	for i, entry := range globalQueue {
		if (entry.Request.State == elevator.UNASSIGNED) && len(avalibaleNodes) > 0 { //OR for redundnacy, both should not be different in theory
			chosenNode := uint8(0)
			switch entry.Request.Calltype {
			case elevator.HALL:
				chosenNode = f_ClosestElevatorNode(entry.Request.Floor, avalibaleNodes)
			case elevator.CAB:
				elevatorAvalibale := false
				for _, nodeInfo := range connectedNodes {
					if nodeInfo.PRIORITY == entry.RequestedNode && nodeInfo.ElevatorInfo.State == elevator.IDLE {
						elevatorAvalibale = true
					}
				}
				if elevatorAvalibale {
					chosenNode = entry.RequestedNode
				} else {
					chosenNode = f_ClosestElevatorNode(entry.Request.Floor, avalibaleNodes)
				}
			}

			entry.Request.State = elevator.ASSIGNED
			assignedEntry = T_GlobalQueueEntry{
				Request:           entry.Request,
				RequestedNode:     entry.RequestedNode,
				AssignedNode:      chosenNode,
				TimeUntilReassign: REASSIGNTIME,
			}
			assignedEntryIndex = i
			break
		}
	}
	return assignedEntry, assignedEntryIndex
}

func F_FindAssignedEntry(globalQueue []T_GlobalQueueEntry, thisNodeInfo T_NodeInfo) (T_GlobalQueueEntry, int) {
	for i, entry := range globalQueue {
		if entry.AssignedNode == thisNodeInfo.PRIORITY && entry.Request.State == elevator.ASSIGNED {
			F_WriteLog("Found assigned request with ID: " + strconv.Itoa(int(entry.Request.Id)) + " assigned to node " + strconv.Itoa(int(entry.RequestedNode)))
			return entry, i // Return both index and entry
		}
	}
	return T_GlobalQueueEntry{}, -1 // Return -1 and an empty entry if not found
}
