package node

import (
	"the-elevator/node/elevator"
)

func f_InitNode(config T_Config) T_Node {
	thisElevatorInfo := elevator.T_ElevatorInfo{
		Direction: elevator.NONE,
		Floor:     1, //-1, 1 for test purposes only!
		State:     elevator.IDLE,
	}
	thisNodeInfo := T_NodeInfo{
		PRIORITY:     config.Priority,
		ElevatorInfo: thisElevatorInfo,
		MSRole:       MSROLE_MASTER,
	}

	thisElevator := elevator.T_Elevator{
		P_info:         &thisElevatorInfo,
		P_serveRequest: nil,
		CurrentID:      0,
		Obstructed:     false,
		StopButton:     false,
	}
	thisNode := T_Node{
		NodeInfo: thisNodeInfo,
		PBRole:   PBROLE_BACKUP,
		Elevator: thisElevator,
	}
	return thisNode
}

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

func f_FindNodeInfo(nodePriority uint8, connectedNodes []T_NodeInfo) T_NodeInfo {
	for _, nodeInfo := range connectedNodes {
		if nodePriority == nodeInfo.PRIORITY {
			return nodeInfo
		}
	}
	return T_NodeInfo{}
}
func f_GetAvalibaleNodes(connectedNodes []T_NodeInfo) []T_NodeInfo {
	var avalibaleNodes []T_NodeInfo
	for i, nodeInfo := range connectedNodes {
		if (nodeInfo != T_NodeInfo{} && nodeInfo.ElevatorInfo.State == elevator.IDLE) {
			avalibaleNodes = append(avalibaleNodes, connectedNodes[i])
		}
	}
	return avalibaleNodes
}

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
		currentNode.TimeUntilDisconnect = CONNECTIONTIME
		connectedNodes := append(oldConnectedNodes, currentNode)
		getSetConnectedNodesInterface.c_set <- connectedNodes
	} else {
		currentNode.TimeUntilDisconnect = CONNECTIONTIME
		oldConnectedNodes[nodeIndex] = currentNode
		getSetConnectedNodesInterface.c_set <- oldConnectedNodes
	}
}
