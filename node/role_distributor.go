package node

//should contain functionality for the master/slave logic
//should take in:
// - this node
// - connected nodes
// - last received message

//give out:
// - T_NodeRole variale in function F_IsMaster

//should also handle the establishment of new master on reconnection

func F_ChooseRole(thisNode *T_Node, connectedNodes []T_NodeInfo) T_NodeRole {
	var returnRole T_NodeRole
	for _, nodeInfo := range connectedNodes {
		if nodeInfo.PRIORITY > thisNode.P_info.PRIORITY {
			returnRole = SLAVE
		} else {
			returnRole = MASTER
		}
	}
	return returnRole
}
