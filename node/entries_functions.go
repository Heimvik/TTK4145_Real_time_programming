package node

import (
	"strconv"
	"the-elevator/node/elevator"
	"time"
)

/*
Calculates and returns the absolute value of an int8 number.

Prerequisites: None.

Returns: The absolute value of the input int8 number.
*/
func f_AbsInt(x int8) int8 {
	if x < 0 {
		return -x
	}
	return x
}

/*
Checks if two elevator queue entries are identical based on their request ID and requested node.

Prerequisites: None.

Returns: A boolean indicating whether the two entries are equal.
*/
func f_EntriesAreEqual(e1 T_GlobalQueueEntry, e2 T_GlobalQueueEntry) bool {
	return ((e1.Request.Id == e2.Request.Id) && (e1.RequestedNode == e2.RequestedNode))
}

/*
Compares two global queue arrays to determine if they contain identical entries in the same order.

Prerequisites: None.

Returns: A boolean indicating whether the two global queues are equal.
*/
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

/*
Determines if all entries in the global queue are obstructed, based on the elevator state of their assigned nodes.

Prerequisites: The global queue and the list of connected nodes must be initialized.

Returns: A boolean indicating whether the entire global queue is obstructed.
*/
func f_GlobalQueueShouldEmpty(globalQueue []T_GlobalQueueEntry) bool {
	obstructedNodes := 0
	assignedToUnConNodes := 0
	connectedNodes := f_GetConnectedNodes()

	for _, entry := range globalQueue {
		assignedNode := f_FindNodeInfo(entry.AssignedNode, connectedNodes)
		if assignedNode.ElevatorInfo.Obstructed {
			obstructedNodes += 1
		}
		if (assignedNode == T_NodeInfo{}) {
			assignedToUnConNodes += 1
		}
	}
	isObstructed := (len(globalQueue) == obstructedNodes)
	isAssignedToUnCon := (len(globalQueue) == assignedToUnConNodes)
	return !isObstructed && !isAssignedToUnCon
}

/*
Searches the global queue for the first entry that has TimeUntilReassign equals 0, indicating it is done.

Prerequisites: An initialized global queue with at least one entry.

Returns: The first done entry found in the global queue and its index; returns an empty entry and 0 if none are found.
*/
func f_FindDoneEntry(globalQueue []T_GlobalQueueEntry) (T_GlobalQueueEntry, int) {
	doneEntry, doneEntryIndex := T_GlobalQueueEntry{}, 0
	for i, entry := range globalQueue {
		if entry.TimeUntilReassign == 0 {
			doneEntry = globalQueue[i]
			doneEntryIndex = i
		}
	}
	return doneEntry, doneEntryIndex
}

/*
Creates and returns a deep copy of the global elevator queue, ensuring modifications do not affect the original.

Prerequisites: None.

Returns: A deep copy of the global queue array.
*/
func f_CopyGlobalQueue(globalQueue []T_GlobalQueueEntry) []T_GlobalQueueEntry {
	deepCopyGlobalQueue := make([]T_GlobalQueueEntry, len(globalQueue))
	for i, entry := range globalQueue {
		deepCopyGlobalQueue[i] = entry
	}
	return deepCopyGlobalQueue
}

/*
Generates a list of all possible elevator requests based on the configured number of floors and available call types.

Prerequisites: Initialization of the system with a defined number of floors.

Returns: An array of potential elevator requests, covering all floors and directions.
*/
func f_FindPossibleRequests() []elevator.T_Request {
	possibleCalls := []elevator.T_CallType{elevator.CALLTYPE_CAB, elevator.CALLTYPE_HALL}
	possibleFloors := []int8{}
	for i := 0; i < int(FLOORS); i++ {
		possibleFloors = append(possibleFloors, int8(i))
	}
	possibleDirections := []elevator.T_ElevatorDirection{-1, 1}
	possibleRequests := make([]elevator.T_Request, 0)
	for _, floor := range possibleFloors {
		for _, call := range possibleCalls {
			if call == elevator.CALLTYPE_HALL {
				for _, direction := range possibleDirections {
					if !(floor == FLOORS-1 && direction == elevator.ELEVATORDIRECTION_UP) || !(floor == 0 && direction == elevator.ELEVATORDIRECTION_DOWN) {
						possibleRequests = append(possibleRequests, elevator.T_Request{Id: 0, State: 0, Calltype: call, Floor: floor, Direction: direction})
					}
				}
			} else if call == elevator.CALLTYPE_CAB {
				possibleRequests = append(possibleRequests, elevator.T_Request{Id: 0, State: 0, Calltype: call, Floor: floor, Direction: elevator.ELEVATORDIRECTION_NONE})
			}
		}
	}
	return possibleRequests
}

/*
Identifies elevator requests that are possible based on the system's configuration but not currently present in the global queue.

Prerequisites: Initialized global queue and a list of all possible requests.

Returns: An array of elevator requests not currently in the global queue.
*/
func f_FindNotPresentRequests(globalQueue []T_GlobalQueueEntry, possibleRequests []elevator.T_Request) []elevator.T_Request {
	notPresentRequests := make([]elevator.T_Request, 0)
	if len(globalQueue) == 0 {
		return possibleRequests
	}
	for _, request := range possibleRequests {
		found := false
		for _, entry := range globalQueue {
			if request.Floor == entry.Request.Floor {
				if request.Calltype == elevator.CALLTYPE_HALL && entry.Request.Calltype == elevator.CALLTYPE_HALL && request.Direction == entry.Request.Direction {
					found = true
					break
				} else if request.Calltype == elevator.CALLTYPE_CAB && entry.Request.Calltype == elevator.CALLTYPE_CAB && entry.AssignedNode == f_GetNodeInfo().PRIORITY {
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

/*
Determines the closest elevator node to a given floor from a list of available nodes, based on their current floor positions.

Prerequisites: A list of connected and available elevator nodes with known floor positions.

Returns: The priority identifier of the closest elevator node to the specified floor.
*/

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

/*
Creates a global queue entry from an elevator request, assigning it based on the current system state and the specifics of the request.

Prerequisites: The request must specify its state, and the system must have current node and global queue information.

Returns: A global queue entry structured from the given request, ready for queue insertion or update.
*/
func f_AssembleEntryFromRequest(receivedRequest elevator.T_Request, thisNodeInfo T_NodeInfo, assignedEntry T_GlobalQueueEntry) T_GlobalQueueEntry {
	returnEntry := T_GlobalQueueEntry{}
	if receivedRequest.State == elevator.REQUESTSTATE_DONE {
		F_WriteLog("Node: | " + strconv.Itoa(int(thisNodeInfo.PRIORITY)) + " | request resent DONE")
		returnEntry = T_GlobalQueueEntry{
			Request:           receivedRequest,
			RequestedNode:     assignedEntry.RequestedNode,
			AssignedNode:      assignedEntry.AssignedNode,
			TimeUntilReassign: 0,
		}
	} else if receivedRequest.State == elevator.REQUESTSTATE_ACTIVE {
		F_WriteLog("Node: | " + strconv.Itoa(int(thisNodeInfo.PRIORITY)) + " | request resent ACTIVE")
		returnEntry = T_GlobalQueueEntry{
			Request:           receivedRequest,
			RequestedNode:     assignedEntry.RequestedNode,
			AssignedNode:      assignedEntry.AssignedNode,
			TimeUntilReassign: REASSIGN_PERIOD,
		}
	} else if receivedRequest.State == elevator.REQUESTSTATE_UNASSIGNED {
		returnEntry = T_GlobalQueueEntry{
			Request:           receivedRequest,
			RequestedNode:     thisNodeInfo.PRIORITY,
			AssignedNode:      0,
			TimeUntilReassign: REASSIGN_PERIOD,
		}
	} else {
		F_WriteLog("Error: Received Assigned request from elevator")
	}
	return returnEntry
}

/*
Selects an unassigned request from the global queue and assigns it to the most suitable elevator node based on current system conditions.

Prerequisites: A list of available elevator nodes and an initialized global queue with at least one unassigned request.

Returns: The updated global queue entry with an assigned elevator node and its index in the queue.
*/
func f_AssignNewEntry(globalQueue []T_GlobalQueueEntry, avalibaleNodes []T_NodeInfo) (T_GlobalQueueEntry, int) {
	assignedEntry := T_GlobalQueueEntry{}
	assignedEntryIndex := -1
	for i, entry := range globalQueue {
		chosenNode := uint8(0)
		switch entry.Request.Calltype {
		case elevator.CALLTYPE_HALL:
			if (entry.Request.State == elevator.REQUESTSTATE_UNASSIGNED) && len(avalibaleNodes) > 0 {
				chosenNode = f_ClosestElevatorNode(entry.Request.Floor, avalibaleNodes)
			}
		case elevator.CALLTYPE_CAB:
			for _, avalibaleNode := range avalibaleNodes {
				if (entry.Request.State == elevator.REQUESTSTATE_UNASSIGNED) && (avalibaleNode.PRIORITY == entry.RequestedNode) {
					chosenNode = entry.RequestedNode
				}
			}
		}
		if chosenNode != 0 {
			entry.Request.State = elevator.REQUESTSTATE_ASSIGNED
			assignedEntry = T_GlobalQueueEntry{
				Request:           entry.Request,
				RequestedNode:     entry.RequestedNode,
				AssignedNode:      chosenNode,
				TimeUntilReassign: REASSIGN_PERIOD,
			}
			assignedEntryIndex = i
			break
		}
	}
	return assignedEntry, assignedEntryIndex
}

/*
Searches the global queue for an entry assigned to the current node, indicating an active request that the node is responsible for servicing.

Prerequisites: An initialized global queue and current node information.

Returns: The found global queue entry assigned to the current node and its index, or an empty entry and -1 if none are found.
*/
func f_FindAssignedEntry(globalQueue []T_GlobalQueueEntry, thisNodeInfo T_NodeInfo) (T_GlobalQueueEntry, int) {
	for i, entry := range globalQueue {
		if entry.AssignedNode == thisNodeInfo.PRIORITY && entry.Request.State == elevator.REQUESTSTATE_ASSIGNED {
			F_WriteLog("Found assigned request with ID: " + strconv.Itoa(int(entry.Request.Id)) + " assigned to node " + strconv.Itoa(int(entry.AssignedNode)))
			return entry, i
		}
	}
	return T_GlobalQueueEntry{}, -1
}

/*
Locates a specific entry within the global queue based on a comparison of request ID and requested node.

Prerequisites: An initialized global queue containing the current state of elevator requests.

Returns: The global queue entry matching the search criteria, or an empty entry if not found.
*/
func f_FindEntry(entryToFind T_GlobalQueueEntry, globalQueue []T_GlobalQueueEntry) T_GlobalQueueEntry {
	for _, entry := range globalQueue {
		if f_EntriesAreEqual(entryToFind, entry) {
			return entry
		}
	}
	return T_GlobalQueueEntry{}
}

/*
Incorporates changes from a master message into the local global queue, potentially adding new entries or updating existing ones based on their state.

Prerequisites: An initialized global queue and a master message containing global queue updates.

Returns: Nothing, but updates the local global queue to reflect the received changes.
*/
func f_UpdateGlobalQueue(c_getSetGlobalQueueInterface chan T_GetSetGlobalQueueInterface, getSetGlobalQueueInterface T_GetSetGlobalQueueInterface, masterMessage T_MasterMessage) {
	entriesToRemove := []T_GlobalQueueEntry{}
	for _, remoteEntry := range masterMessage.GlobalQueue {
		f_AddEntryGlobalQueue(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, remoteEntry)
		if remoteEntry.Request.State == elevator.REQUESTSTATE_DONE {
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

/*
Removes specific entries from the global queue based on their request ID and requested node, usually after they have been completed or reassigned.

Prerequisites: An initialized global queue and a list of entries to be removed.

Returns: A modified global queue with the specified entries removed.
*/
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

/*
Processes and removes an entry marked as finished from the global queue, after ensuring it is acknowledged by all nodes.

Prerequisites: An initialized global queue, current node information, and a finished entry to be removed.

Returns: The global queue with the finished entry removed, following successful acknowledgment.
*/
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
		//f_TurnOffLight(finishedEntry)
		F_WriteLog("Removed entry: | " + strconv.Itoa(int(finishedEntry.Request.Id)) + " | " + strconv.Itoa(int(finishedEntry.RequestedNode)) + " | from global queue")
		globalQueue = append(globalQueue[:finishedEntryIndex], globalQueue[finishedEntryIndex+1:]...)
		return globalQueue
	case <-breakOutTimer.C:
		return globalQueue
	}
}

/*
Updates an entry in the global queue to be unassigned and ready for reassignment, typically used when an entry hasn't been completed in the expected timeframe.

Prerequisites: An initialized global queue and an entry identified as unfinished or needing reassignment.

Returns: The global queue with the specified entry updated for reassignment.
*/
func f_ReassignUnfinishedEntry(globalQueue []T_GlobalQueueEntry, unFinishedEntry T_GlobalQueueEntry, unFinishedEntryIndex int) []T_GlobalQueueEntry {
	unFinishedEntry.Request.State = elevator.REQUESTSTATE_UNASSIGNED
	entryToReassign := T_GlobalQueueEntry{
		Request:           unFinishedEntry.Request,
		RequestedNode:     unFinishedEntry.RequestedNode,
		AssignedNode:      0,
		TimeUntilReassign: REASSIGN_PERIOD,
	}
	globalQueue[unFinishedEntryIndex] = entryToReassign
	F_WriteLog("Reassigned entry: | " + strconv.Itoa(int(unFinishedEntry.Request.Id)) + " | " + strconv.Itoa(int(unFinishedEntry.RequestedNode)) + " | in global queue")
	return globalQueue
}

/*
Adds a new entry to the global queue or updates an existing entry if it matches based on request ID and requested node, ensuring the queue reflects the current system state.

Prerequisites: An initialized global queue and an entry to add or update.

Returns: Nothing, but updates the global queue with the new or modified entry.
*/
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
	if entryIsUnique && entryToAdd.Request.State != elevator.REQUESTSTATE_DONE {
		globalQueue = append(globalQueue, entryToAdd)

	} else if !entryIsUnique {
		if (entryToAdd.Request.State >= globalQueue[entryIndex].Request.State || entryToAdd.TimeUntilReassign < globalQueue[entryIndex].TimeUntilReassign) ||
			(entryToAdd.AssignedNode != globalQueue[entryIndex].AssignedNode && entryToAdd.AssignedNode != 0 && globalQueue[entryIndex].AssignedNode != 0) {
			globalQueue[entryIndex] = entryToAdd
		} else {
			F_WriteLog("Disallowed backward information")
		}
	}
	getSetGlobalQueueInterface.c_set <- globalQueue
}
