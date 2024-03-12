package node

import (
	"strconv"
	"the-elevator/node/elevator"
	"time"
)

func f_AbsInt(x int8) int8 {
	if x < 0 {
		return -x
	}
	return x
}

func f_EntriesAreEqual(e1 T_GlobalQueueEntry, e2 T_GlobalQueueEntry) bool {
	return ((e1.Request.Id == e2.Request.Id) && (e1.RequestedNode == e2.RequestedNode))
}

func f_GlobalQueueAreEqual(q1 []T_GlobalQueueEntry, q2 []T_GlobalQueueEntry) bool {
	if len(q1) != len(q2) {
		return false
	}

	for i := 0; i < len(q1); i++ {
		if q1[i] != q2[i] {
			return false
		}
	}
	return true
}

func f_FindPossibleRequests() []elevator.T_Request {
	possibleCalls := []elevator.T_Call{elevator.CAB, elevator.HALL}
	possibleFloors := []int8{}
	for i := 0; i < int(FLOORS); i++ {
		possibleFloors = append(possibleFloors, int8(i))
	}
	possibleDirections := []elevator.T_ElevatorDirection{-1, 1}
	possibleRequests := make([]elevator.T_Request, 0)
	for _, floor := range possibleFloors {
		for _, call := range possibleCalls {
			if call == elevator.HALL {
				for _, direction := range possibleDirections {
					if !(floor == FLOORS-1 && direction == elevator.UP) || !(floor == 0 && direction == elevator.DOWN) {
						possibleRequests = append(possibleRequests, elevator.T_Request{0, 0, call, floor, direction})
					}
				}
			} else if call == elevator.CAB {
				possibleRequests = append(possibleRequests, elevator.T_Request{0, 0, call, floor, elevator.NONE})
			}
		}
	}
	return possibleRequests
}

func f_FindNotPresentRequests(globalQueue []T_GlobalQueueEntry, possibleRequests []elevator.T_Request) []elevator.T_Request {
	notPresentRequests := make([]elevator.T_Request, 0)
	if len(globalQueue) == 0 {
		return possibleRequests
	}
	for _, request := range possibleRequests {
		found := false
		for _, entry := range globalQueue {
			if request.Floor == entry.Request.Floor {
				if request.Calltype == elevator.HALL && entry.Request.Calltype == elevator.HALL && request.Direction == entry.Request.Direction {
					found = true
					break
				} else if request.Calltype == elevator.CAB && entry.Request.Calltype == elevator.CAB {
					found = true
					break
				}
			}
		}
		if !found {
			notPresentRequests = append(notPresentRequests, request)
		}
	}
	return notPresentRequests
}
func f_ClosestElevatorNode(floor int8, nodes []T_NodeInfo) uint8 {
	var closestNode T_NodeInfo
	closestDifference := int8(FLOORS)
	for _, nodeInfo := range nodes {
		currentDifference := f_AbsInt(int8(nodeInfo.ElevatorInfo.Floor) - floor)
		if currentDifference < closestDifference {
			closestDifference = currentDifference
			closestNode = nodeInfo
		}
		if currentDifference > FLOORS {
			F_WriteLog("Error: Found floordifference larger than max floors")
		}
	}
	return closestNode.PRIORITY
}

func F_AssembleEntryFromRequest(receivedRequest elevator.T_Request, thisNodeInfo T_NodeInfo, assignedEntry T_GlobalQueueEntry) T_GlobalQueueEntry {
	returnEntry := T_GlobalQueueEntry{}
	if receivedRequest.State == elevator.DONE {
		F_WriteLog("Node: | " + strconv.Itoa(int(thisNodeInfo.PRIORITY)) + " | request resent DONE")
		returnEntry = T_GlobalQueueEntry{
			Request:           receivedRequest,
			RequestedNode:     assignedEntry.RequestedNode,
			AssignedNode:      assignedEntry.AssignedNode,
			TimeUntilReassign: 0,
		}
	} else if receivedRequest.State == elevator.ACTIVE {
		F_WriteLog("Node: | " + strconv.Itoa(int(thisNodeInfo.PRIORITY)) + " | request resent ACTIVE")
		returnEntry = T_GlobalQueueEntry{
			Request:           receivedRequest,
			RequestedNode:     assignedEntry.RequestedNode,
			AssignedNode:      assignedEntry.AssignedNode,
			TimeUntilReassign: REASSIGNTIME,
		}
	} else if receivedRequest.State == elevator.UNASSIGNED {
		returnEntry = T_GlobalQueueEntry{
			Request:           receivedRequest,
			RequestedNode:     thisNodeInfo.PRIORITY,
			AssignedNode:      0,
			TimeUntilReassign: REASSIGNTIME,
		}
	} else {
		F_WriteLog("Error: Received Assigned request from elevator")
	}
	return returnEntry
}

func F_AssignNewEntry(globalQueue []T_GlobalQueueEntry, connectedNodes []T_NodeInfo, avalibaleNodes []T_NodeInfo) (T_GlobalQueueEntry, int) {
	assignedEntry := T_GlobalQueueEntry{}
	assignedEntryIndex := -1
	for i, entry := range globalQueue {
		if (entry.Request.State == elevator.UNASSIGNED) && len(avalibaleNodes) > 0 {
			chosenNode := uint8(0)
			switch entry.Request.Calltype {
			case elevator.HALL:
				chosenNode = f_ClosestElevatorNode(entry.Request.Floor, avalibaleNodes)
			case elevator.CAB:
				chosenNode = entry.RequestedNode
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
			F_WriteLog("Found assigned request with ID: " + strconv.Itoa(int(entry.Request.Id)) + " assigned to node " + strconv.Itoa(int(entry.AssignedNode)))
			return entry, i
		}
	}
	return T_GlobalQueueEntry{}, -1
}

func f_FindEntry(entryToFind T_GlobalQueueEntry, globalQueue []T_GlobalQueueEntry) T_GlobalQueueEntry {
	for _, entry := range globalQueue {
		if f_EntriesAreEqual(entryToFind, entry) {
			return entry
		}
	}
	return T_GlobalQueueEntry{}
}

func f_UpdateGlobalQueue(c_getSetGlobalQueueInterface chan T_GetSetGlobalQueueInterface, getSetGlobalQueueInterface T_GetSetGlobalQueueInterface, masterMessage T_MasterMessage) {
	entriesToRemove := []T_GlobalQueueEntry{}
	for _, remoteEntry := range masterMessage.GlobalQueue {
		f_AddEntryGlobalQueue(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, remoteEntry)
		if remoteEntry.Request.State == elevator.DONE {
			entriesToRemove = append(entriesToRemove, remoteEntry)
		}
	}

	if len(entriesToRemove) > 0 {
		c_getSetGlobalQueueInterface <- getSetGlobalQueueInterface
		globalQueue := <-getSetGlobalQueueInterface.c_get
		globalQueue = f_RemoveEntryGlobalQueue(globalQueue, entriesToRemove)
		getSetGlobalQueueInterface.c_set <- globalQueue
	}
}

func f_RemoveEntryGlobalQueue(globalQueue []T_GlobalQueueEntry, entriesToRemove []T_GlobalQueueEntry) []T_GlobalQueueEntry {
	newGlobalQueue := globalQueue
	for i, entry := range globalQueue {
		for _, entryToRemove := range entriesToRemove {
			if entry.Request.Id == entryToRemove.Request.Id && entry.RequestedNode == entryToRemove.RequestedNode {
				newGlobalQueue = append(globalQueue[:i], globalQueue[i+1:]...)
			}
		}
	}
	return newGlobalQueue
}
func f_RemoveFinishedEntry(c_ackSentEntryToSlave chan T_AckObject, globalQueue []T_GlobalQueueEntry, thisNodeInfo T_NodeInfo, finishedEntry T_GlobalQueueEntry, finishedEntryIndex int) []T_GlobalQueueEntry {
	c_sentDoneEntryToSlave := make(chan bool)
	ackSentEntryToSlave := T_AckObject{
		ObjectToAcknowledge:        globalQueue,
		ObjectToSupportAcknowledge: thisNodeInfo,
		C_Acknowledgement:          c_sentDoneEntryToSlave,
	}
	c_ackSentEntryToSlave <- ackSentEntryToSlave
	breakOutTimer := time.NewTicker(time.Duration(1000) * time.Millisecond)
	F_WriteLog("MASTER found done entry waiting for sending to slave before removing")
	select {
	case <-c_sentDoneEntryToSlave:
		F_WriteLog("Removed entry: | " + strconv.Itoa(int(finishedEntry.Request.Id)) + " | " + strconv.Itoa(int(finishedEntry.RequestedNode)) + " | from global queue")
		globalQueue = append(globalQueue[:finishedEntryIndex], globalQueue[finishedEntryIndex+1:]...)
		return globalQueue
	case <-breakOutTimer.C:
		return globalQueue
	}
}

func f_ReassignUnfinishedEntry(globalQueue []T_GlobalQueueEntry, unFinishedEntry T_GlobalQueueEntry, unFinishedEntryIndex int) []T_GlobalQueueEntry {
	unFinishedEntry.Request.State = elevator.UNASSIGNED
	entryToReassign := T_GlobalQueueEntry{
		Request:           unFinishedEntry.Request,
		RequestedNode:     unFinishedEntry.RequestedNode,
		AssignedNode:      0,
		TimeUntilReassign: REASSIGNTIME,
	}
	globalQueue[unFinishedEntryIndex] = entryToReassign
	F_WriteLog("Reassigned entry: | " + strconv.Itoa(int(unFinishedEntry.Request.Id)) + " | " + strconv.Itoa(int(unFinishedEntry.RequestedNode)) + " | in global queue")
	return globalQueue
}

func f_AddEntryGlobalQueue(c_getSetGlobalQueueInterface chan T_GetSetGlobalQueueInterface, getSetGlobalQueueInterface T_GetSetGlobalQueueInterface, entryToAdd T_GlobalQueueEntry) {
	c_getSetGlobalQueueInterface <- getSetGlobalQueueInterface
	globalQueue := <-getSetGlobalQueueInterface.c_get

	entryIsUnique := true
	entryIndex := 0
	for i, entry := range globalQueue {
		if f_EntriesAreEqual(entryToAdd, entry) {
			entryIsUnique = false
			entryIndex = i
			break
		}
	}
	if entryIsUnique && entryToAdd.Request.State != elevator.DONE {
		globalQueue = append(globalQueue, entryToAdd)
	} else if !entryIsUnique {
		if entryToAdd.Request.State >= globalQueue[entryIndex].Request.State || entryToAdd.TimeUntilReassign < globalQueue[entryIndex].TimeUntilReassign { //only allow forward entry states //>=?
			globalQueue[entryIndex] = entryToAdd
		} else {
			F_WriteLog("Disallowed backward information")
		}
	}
	getSetGlobalQueueInterface.c_set <- globalQueue
}
