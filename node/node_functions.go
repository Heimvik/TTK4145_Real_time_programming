package node

import (
	"the-elevator/node/elevator"
)

/*
Initializes a node with default elevator and system states based on configuration settings, setting up its role, priority, and elevator operational parameters.

Prerequisites: A valid configuration must be provided, including priority and operational settings for the node and its elevator.

Returns: A fully initialized node structure ready for integration into the system's operational flow.
*/
func F_InitNode(config T_Config) T_Node {
	thisElevatorInfo := elevator.T_ElevatorInfo{
		Direction:  elevator.ELEVATORDIRECTION_NONE,
		Floor:      -1,
		State:      elevator.ELEVATORSTATE_IDLE,
		Obstructed: false,
	}
	thisNodeInfo := T_NodeInfo{
		PRIORITY:     config.Priority,
		ElevatorInfo: thisElevatorInfo,
		MSRole:       MSROLE_MASTER,
	}

	thisElevator := elevator.T_Elevator{
		P_info:       &thisElevatorInfo,
		ServeRequest: elevator.T_Request{},
		CurrentID:    0,
		StopButton:   false,
	}
	thisNode := T_Node{
		NodeInfo: thisNodeInfo,
		PBRole:   PBROLE_BACKUP,
		Elevator: thisElevator,
	}
	return thisNode
}

/*
Creates and returns a deep copy of a T_NodeInfo slice, ensuring modifications do not affect the original.

Prerequisites: None.

Returns: A deep copy of the slice.
*/
func f_CopyConnectedNodes(connectedNodes []T_NodeInfo) []T_NodeInfo {
	deepCopyConnectedNodes := make([]T_NodeInfo, len(connectedNodes))
	for i, node := range connectedNodes {
		deepCopyConnectedNodes[i] = node
	}
	return deepCopyConnectedNodes
}

/*
Determines and assigns a new role to the current node based on the priorities of connected nodes, ensuring proper master-slave hierarchy within the network.

Prerequisites: A list of currently connected nodes and their priorities.

Returns: An updated node information structure with the new role assigned to the current node.
*/
func f_AssignNewRole(thisNodeInfo T_NodeInfo, connectedNodes []T_NodeInfo) T_NodeInfo {
	var returnRole T_MSNodeRole = MSROLE_MASTER
	for _, remoteNodeInfo := range connectedNodes {
		if remoteNodeInfo.PRIORITY < thisNodeInfo.PRIORITY {
			returnRole = MSROLE_SLAVE
		}
	}
	newNodeInfo := T_NodeInfo{
		PRIORITY:            thisNodeInfo.PRIORITY,
		MSRole:              returnRole,
		TimeUntilDisconnect: thisNodeInfo.TimeUntilDisconnect,
		ElevatorInfo:        thisNodeInfo.ElevatorInfo,
	}
	return newNodeInfo
}

/*
Searches for and returns the information of a node within the list of connected nodes based on its priority identifier.

Prerequisites: A list of connected nodes.

Returns: The information of the specified node if found; otherwise, returns an empty node information structure.
*/
func f_FindNodeInfo(nodePriority uint8, connectedNodes []T_NodeInfo) T_NodeInfo {
	for _, nodeInfo := range connectedNodes {
		if nodePriority == nodeInfo.PRIORITY {
			return nodeInfo
		}
	}
	return T_NodeInfo{}
}

/*
Filters and returns a list of connected nodes that are available and idle, ready to be assigned new elevator requests.

Prerequisites: A list of currently connected nodes with updated elevator states.

Returns: A list of node information structures for nodes that are idle and available for assignments.
*/
func f_GetAvalibaleNodes(connectedNodes []T_NodeInfo) []T_NodeInfo {
	var avalibaleNodes []T_NodeInfo
	for i, nodeInfo := range connectedNodes {
		if (nodeInfo != T_NodeInfo{} && nodeInfo.ElevatorInfo.State == elevator.ELEVATORSTATE_IDLE) {
			avalibaleNodes = append(avalibaleNodes, connectedNodes[i])
		}
	}
	return avalibaleNodes
}

/*
Updates the list of connected nodes with the latest information of a node, adding it if new or updating its status if already present.

Prerequisites: An initialized list of connected nodes and updated information for the node to be added or updated.

Returns: Nothing, but modifies the global state of connected nodes based on the provided node information.
*/
func f_UpdateConnectedNodes(c_getSetConnectedNodesInterface chan T_GetSetConnectedNodesInterface, getSetConnectedNodesInterface T_GetSetConnectedNodesInterface, currentNode T_NodeInfo) {
	c_getSetConnectedNodesInterface <- getSetConnectedNodesInterface
	oldConnectedNodes := <-getSetConnectedNodesInterface.c_get

	nodeIsUnique := true
	nodeIndex := 0
	for i, oldConnectedNode := range oldConnectedNodes {
		if currentNode.PRIORITY == oldConnectedNode.PRIORITY {
			nodeIsUnique = false
			nodeIndex = i
			break
		}
	}

	if nodeIsUnique {
		currentNode.TimeUntilDisconnect = CONNECTION_PERIOD
		connectedNodes := append(oldConnectedNodes, currentNode)
		getSetConnectedNodesInterface.c_set <- connectedNodes
	} else {
		currentNode.TimeUntilDisconnect = CONNECTION_PERIOD
		oldConnectedNodes[nodeIndex] = currentNode
		getSetConnectedNodesInterface.c_set <- oldConnectedNodes
	}
}
