
- [1. Known issues:](#1-known-issues)
  - [1.1. Disallow backward information in reassign](#11-disallow-backward-information-in-reassign)
    - [1.1.1. Context](#111-context)
    - [1.1.2. Problem](#112-problem)
    - [1.1.3. Solution](#113-solution)
  - [1.2. DONE entries not removed when two masters present](#12-done-entries-not-removed-when-two-masters-present)
    - [1.2.1. Context:](#121-context)
    - [1.2.2. Problem:](#122-problem)
    - [1.2.3. Solution](#123-solution)

# 1. Known issues:
## 1.1. Disallow backward information in reassign
### 1.1.1. Context
### 1.1.2. Problem
Upon having both nodes connected, disconnecting master, connecting master, we get the log:  
```
Node: | 1 | MASTER | has GQ:  
Entry: | 3 | State: | ACTIVE | Calltype: HALL | Floor: 1 | Direction: UP | Reassigned in: 2.00 | Requested node: | 2 | Assigned node: | 1 |   
Entry: | 4 | State: | ACTIVE | Calltype: HALL | Floor: 2 | Direction: UP | Reassigned in: 8.00 | Requested node: | 2 | Assigned node: | 1 |  
Entry: | 5 | State: | ACTIVE | Calltype: HALL | Floor: 3 | Direction: UP | Reassigned in: 10.00 | Requested node: | 2 | Assigned node: | 2 |   
Node: | 2 | SLAVE | received MM from | 1 | MASTER | GlobalQueue: [Request ID: | 3 | State: | ACTIVE | Calltype: HALL | Floor: 1 | Direction: UP | Reassigned in: 1 | Requested node: | 2 | Assigned node: | 1 |, Request ID: | 4 | State: | ACTIVE | Calltype: HALL | Floor: 2 | Direction: UP | Reassigned in: 7 | Requested node: | 2 | Assigned node: | 1 |, Request ID: | 5 | State: | ACTIVE | Calltype: HALL | Floor: 3 | Direction: UP | Reassigned in: 9 | Requested node: | 2 | Assigned node: | 2 |]  
Node: | 1 | MASTER | received MM from | 1 | MASTER | GlobalQueue: [Request ID: | 3 | State: | ACTIVE | Calltype: HALL | Floor: 1 | Direction: UP | Reassigned in: 1 | Requested node: | 2 | Assigned node: | 1 |, Request ID: | 4 | State: | ACTIVE | Calltype: HALL | Floor: 2 | Direction: UP | Reassigned in: 7 | Requested node: | 2 | Assigned node: | 1 |, Request ID: | 5 | State: | ACTIVE | Calltype: HALL | Floor: 3 | Direction: UP | Reassigned in: 9 | Requested node: | 2 | Assigned node: | 2 |]  
Node: | 1 | MASTER | has connected nodes | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) | 2 (Role: SLAVE, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 3) |   
Node: | 2 | SLAVE | has GQ:  
Entry: | 3 | State: | ACTIVE | Calltype: HALL | Floor: 1 | Direction: UP | Reassigned in: 1.00 | Requested node: | 2 | Assigned node: | 1 |   
Entry: | 4 | State: | ACTIVE | Calltype: HALL | Floor: 2 | Direction: UP | Reassigned in: 7.00 | Requested node: | 2 | Assigned node: | 1 |   
Entry: | 5 | State: | ACTIVE | Calltype: HALL | Floor: 3 | Direction: UP | Reassigned in: 9.00 | Requested node: | 2 | Assigned node: | 2 |   
Node: | 1 | MASTER | received SM from | 2 | Request ID: | 0 | State: | UNASSIGNED | Calltype: NONE | Floor: 0 | Direction: NONE |  
Node: | 1 | MASTER | has connected nodes | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) | 2 (Role: SLAVE, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) |   
Node: | 2 | SLAVE | received SM from | 2 | Request ID: | 0 | State: | UNASSIGNED | Calltype: NONE | Floor: 0 | Direction: NONE |  
Reassigned entry: | 3 | 2 | in global queue  
Node: | 2 | SLAVE | received MM from | 1 | MASTER | GlobalQueue: [Request ID: | 3 | State: | UNASSIGNED | Calltype: HALL | Floor: 1 | Direction: UP | Reassigned in: 15 | Requested node: | 2 | Assigned node: | 0 |, Request ID: | 4 | State: | ACTIVE | Calltype: HALL | Floor: 2 | Direction: UP | Reassigned in: 6 | Requested node: | 2 | Assigned node: | 1 |, Request ID: | 5 | State: | ACTIVE | Calltype: HALL | Floor: 3 | Direction: UP | Reassigned in: 8 | Requested node: | 2 | Assigned node: | 2 |]  
Disallowed backward information  
Node: | 1 | MASTER | received MM from | 1 | MASTER | GlobalQueue: [Request ID: | 3 | State: | UNASSIGNED | Calltype: HALL | Floor: 1 | Direction: UP | Reassigned in: 15 | Requested node: | 2 | Assigned node: | 0 |, Request ID: | 4 | State: | ACTIVE | Calltype: HALL | Floor: 2 | Direction: UP | Reassigned in: 6 | Requested node: | 2 | Assigned node: | 1 |, Request ID: | 5 | State: | ACTIVE | Calltype: HALL | Floor: 3 | Direction: UP | Reassigned in: 8 | Requested node: | 2 | Assigned node: | 2 |]  
Node: | 1 | MASTER | has connected nodes | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) | 2 (Role: SLAVE, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 3) |   
Node: | 1 | MASTER | received SM from | 2 | Request ID: | 0 | State: | UNASSIGNED | Calltype: NONE | Floor: 0 | Direction: NONE |  
Node: | 1 | MASTER | has connected nodes | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) | 2 (Role: SLAVE, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) |   
Node: | 2 | SLAVE | received SM from | 2 | Request ID: | 0 | State: | UNASSIGNED | Calltype: NONE | Floor: 0 | Direction: NONE |  
Node: | 1 | MASTER | has GQ:  
Entry: | 3 | State: | UNASSIGNED | Calltype: HALL | Floor: 1 | Direction: UP | Reassigned in: 15.00 | Requested node: | 2 | Assigned node: | 0 |   
Entry: | 4 | State: | ACTIVE | Calltype: HALL | Floor: 2 | Direction: UP | Reassigned in: 6.00 | Requested node: | 2 | Assigned node: | 1 |   
Entry: | 5 | State: | ACTIVE | Calltype: HALL | Floor: 3 | Direction: UP | Reassigned in: 8.00 | Requested node: | 2 | Assigned node: | 2 |   
Node: | 2 | SLAVE | received MM from | 1 | MASTER | GlobalQueue: [Request ID: | 3 | State: | UNASSIGNED | Calltype: HALL | Floor: 1 | Direction: UP | Reassigned in: 15 | Requested node: | 2 | Assigned node: | 0 |, Request ID: | 4 | State: | ACTIVE | Calltype: HALL | Floor: 2 | Direction: UP | Reassigned in: 5 | Requested node: | 2 | Assigned node: | 1 |, Request ID: | 5 | State: | ACTIVE | Calltype: HALL | Floor: 3 | Direction: UP | Reassigned in: 7 | Requested node: | 2 | Assigned node: | 2 |]  
Node: | 1 | MASTER | received MM from | 1 | MASTER | GlobalQueue: [Request ID: | 3 | State: | UNASSIGNED | Calltype: HALL | Floor: 1 | Direction: UP | Reassigned in: 15 | Requested node: | 2 | Assigned node: | 0 |, Request ID: | 4 | State: | ACTIVE | Calltype: HALL | Floor: 2 | Direction: UP | Reassigned in: 5 | Requested node: | 2 | Assigned node: | 1 |, Request ID: | 5 | State: | ACTIVE | Calltype: HALL | Floor: 3 | Direction: UP | Reassigned in: 7 | Requested node: | 2 | Assigned node: | 2 |]  
Node: | 1 | MASTER | has connected nodes | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) | 2 (Role: SLAVE, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 3) |   
Disallowed backward information  
Node: | 2 | SLAVE | has GQ:  
Entry: | 3 | State: | ACTIVE | Calltype: HALL | Floor: 1 | Direction: UP | Reassigned in: 1.00 | Requested node: | 2 | Assigned node: | 1 |   
Entry: | 4 | State: | ACTIVE | Calltype: HALL | Floor: 2 | Direction: UP | Reassigned in: 5.00 | Requested node: | 2 | Assigned node: | 1 |   
Entry: | 5 | State: | ACTIVE | Calltype: HALL | Floor: 3 | Direction: UP | Reassigned in: 7.00 | Requested node: | 2 | Assigned node: | 2 |   
Node: | 1 | MASTER | received SM from | 2 | Request ID: | 0 | State: | UNASSIGNED | Calltype: NONE | Floor: 0 | Direction: NONE |  
Node: | 1 | MASTER | has connected nodes | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) | 2 (Role: SLAVE, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) |   
Node: | 2 | SLAVE | received SM from | 2 | Request ID: | 0 | State: | UNASSIGNED | Calltype: NONE | Floor: 0 | Direction: NONE |  
Node: | 1 | request resent DONE  
Assigned request with ID: 3 assigned to node 1
```
### 1.1.3. Solution
Aka we disallow backward information when a entry is reassigned. This fixes itself.

## 1.2. DONE entries not removed when two masters present

### 1.2.1. Context:  
Node 1 is run, role is set to master. Node 2 is run, role is set to slave. Entries get added to GQ, Node 1 is terminated, Node 2 is set to master.
After disconnect time runs out Node 1 is revived. While both Node 1 and Node 2 are master, Node 1 assigns and finishes entry 5, and sends a MM to the slaves to remove entry 5 from GQ.  
### 1.2.2. Problem:
As Node 2 is still master when entry 5 should be removed, Node 2 will keep entry 5 as DONE. The following log describes the current problem:   
```
Started all master routines
Node: | 1 | MASTER | received MM from | 2 | MASTER | GlobalQueue: [Request ID: | 1 | State: | ACTIVE | Calltype: CAB | Floor: 1 | Direction: NONE | Reassigned in: 6 | Requested Node: | 1 | Assigned node: | 1 |, Request ID: | 4 | State: | ACTIVE | Calltype: CAB | Floor: 3 | Direction: NONE | Reassigned in: 12 | Requested node: | 1 | Assigned node: | 2 |, Request ID: | 5 | State: | UNASSIGNED | Calltype: HALL | Floor: 1 | Direction: NONE | Reassigned in: 15 | Requested node: | 1 | Assigned node: | 0 |, Request ID: | 6 | State: | UNASSIGNED | Calltype: HALL | Floor: 2 | Direction: NONE | Reassigned in: 15 | Requested node: | 1 | Assigned node: | 0 |]  
Node: | 1 | MASTER | has connected nodes | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:0}, TimeUntilDisconnect: 4) | 2 (Role: MASTER, ElevatorInfo: {Direction:1 Floor:1 State:2}, TimeUntilDisconnect: 4) |  
Node: | 2 | MASTER | received MM from | 2 | MASTER | GlobalQueue: [Request ID: | 1 | State: | ACTIVE | Calltype: CAB | Floor: 1 | Direction: NONE | Reassigned in: 6 | Requested node: | 1 | Assigned node: | 1 |, Request ID: | 4 | State: | ACTIVE | Calltype: CAB | Floor: 3 | Direction: NONE | Reassigned in: 12 | Requested node: | 1 | Assigned node: | 2 |, Request ID: | 5 | State: | UNASSIGNED | Calltype: HALL | Floor: 1 | Direction: NONE | Reassigned in: 15 | Requested node: | 1 | Assigned node: | 0 |, Request ID: | 6 | State: | UNASSIGNED | Calltype: HALL | Floor: 2 | Direction: NONE | Reassigned in: 15 | Requested node: | 1 | Assigned node: | 0 |]  
Node: | 2 | MASTER | has connected nodes | 2 (Role: MASTER, ElevatorInfo: {Direction:1 Floor:1 State:2}, TimeUntilDisconnect: 4) |  
Assigned request with ID: 5 assigned to node 1  
Getting ack from last assinged...  
Found assigned request with ID: 5 assigned to node 1  
Found assigned entry!  
Node: | 1 | request resent ACTIVE  
Node: | 1 | request resent DONE  
Node: | 1 | MASTER | updated GQ entry:  
Entry: | 5 | State: | ACTIVE | Calltype: HALL | Floor: 1 | Direction: NONE | Reassigned in: 15.00 | Requested node: | 1 | Assigned node: | 1 |  
Node: | 1 | MASTER | updated GQ entry:  
Entry: | 5 | State: | DONE | Calltype: HALL | Floor: 1 | Direction: NONE | Reassigned in: 0.00 | Requested node: | 1 | Assigned node: | 1 |   
MASTER found done entry waiting for sending to slave before removing  
Removed entry: | 5 | 1 | from global queue  
Node: | 1 | MASTER | received MM from | 1 | MASTER | GlobalQueue: [Request ID: | 1 | State: | ACTIVE | Calltype: CAB | Floor: 1 | Direction: NONE | Reassigned in: 6 | Requested node: | 1 | Assigned node: | 1 |, Request ID: | 4 | State: | ACTIVE | Calltype: CAB | Floor: 3 | Direction: NONE | Reassigned in: 12 | Requested node: | 1 | Assigned node: | 2 |, Request ID: | 5 | State: | DONE | Calltype: HALL | Floor: 1 | Direction: NONE | Reassigned in: 0 | Requested node: | 1 | Assigned node: | 1 |, Request ID: | 6 | State: | UNASSIGNED | Calltype: HALL | Floor: 2 | Direction: NONE | Reassigned in: 15 | Requested node: | 1 | Assigned node: | 0 |]  
Node: | 1 | MASTER | has connected nodes | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:1}, TimeUntilDisconnect: 4) | 2 (Role: MASTER, ElevatorInfo: {Direction:1 Floor:1 State:2}, TimeUntilDisconnect: 4) |   
Node: | 2 | MASTER | received MM from | 1 | MASTER | GlobalQueue: [Request ID: | 1 | State: | ACTIVE | Calltype: CAB | Floor: 1 | Direction: NONE | Reassigned in: 6 | Requested node: | 1 | Assigned node: | 1 |, Request ID: | 4 | State: | ACTIVE | Calltype: CAB | Floor: 3 | Direction: NONE | Reassigned in: 12 | Requested node: | 1 | Assigned node: | 2 |, Request ID: | 5 | State: | DONE | Calltype: HALL | Floor: 1 | Direction: NONE | Reassigned in: 0 | Requested node: | 1 | Assigned node: | 1 |, Request ID: | 6 | State: | UNASSIGNED | Calltype: HALL | Floor: 2 | Direction: NONE | Reassigned in: 15 | Requested node: | 1 | Assigned node: | 0 |]  
Node: | 2 | MASTER | has connected nodes | 2 (Role: MASTER, ElevatorInfo: {Direction:1 Floor:1 State:2}, TimeUntilDisconnect: 4) | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:1}, TimeUntilDisconnect: 4) |   
Closed: f_CheckAssignedNodeState  
Closed: f_CheckIfShouldAssign  
Closed: f_CheckGlobalQueueEntryStatus  
Node: | 2 | SLAVE | received MM from | 1 | MASTER | GlobalQueue: [Request ID: | 1 | State: | ACTIVE | Calltype: CAB | Floor: 1 | Direction: NONE | Reassigned in: 6 | Requested node: | 1 | Assigned node: | 1 |, Request ID: | 4 | State: | ACTIVE | Calltype: CAB | Floor: 3 | Direction: NONE | Reassigned in: 12 | Requested node: | 1 | Assigned node: | 2 |, Request ID: | 6 | State: | UNASSIGNED | Calltype: HALL | Floor: 2 | Direction: NONE | Reassigned in: 15 | Requested node: | 1 | Assigned node: | 0 |]  
Node: | 1 | MASTER | received MM from | 1 | MASTER | GlobalQueue: [Request ID: | 1 | State: | ACTIVE | Calltype: CAB | Floor: 1 | Direction: NONE | Reassigned in: 6 | Requested node: | 1 | Assigned node: | 1 |, Request ID: | 4 | State: | ACTIVE | Calltype: CAB | Floor: 3 | Direction: NONE | Reassigned in: 12 | Requested node: | 1 | Assigned node: | 2 |, Request ID: | 6 | State: | UNASSIGNED | Calltype: HALL | Floor: 2 | Direction: NONE | Reassigned in: 15 | Requested node: | 1 | Assigned node: | 0 |]  
Node: | 1 | MASTER | has connected nodes | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:1}, TimeUntilDisconnect: 4) | 2 (Role: MASTER, ElevatorInfo: {Direction:1 Floor:1 State:2}, TimeUntilDisconnect: 3) |   
Node: | 2 | SLAVE | has GQ:  
Entry: | 1 | State: | ACTIVE | Calltype: CAB | Floor: 1 | Direction: NONE | Reassigned in: 6.00 | Requested node: | 1 | Assigned node: | 1 |   
Entry: | 4 | State: | ACTIVE | Calltype: CAB | Floor: 3 | Direction: NONE | Reassigned in: 12.00 | Requested node: | 1 | Assigned node: | 2 |   
Entry: | 5 | State: | DONE | Calltype: HALL | Floor: 1 | Direction: NONE | Reassigned in: 0.00 | Requested node: | 1 | Assigned node: | 1 |   
Entry: | 6 | State: | UNASSIGNED | Calltype: HALL | Floor: 2 | Direction: NONE | Reassigned in: 15.00 | Requested node: | 1 | Assigned node: | 0 |     
Closed: f_DecrementTimeUntilReassign  
Node: | 1 | MASTER | received SM from | 2 | Request ID: | 0 | State: | UNASSIGNED | Calltype: NONE | Floor: 0 | Direction: NONE |  
Node: | 1 | MASTER | has connected nodes | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:1}, TimeUntilDisconnect: 4) | 2 (Role: SLAVE, ElevatorInfo: {Direction:1 Floor:1 State:2}, TimeUntilDisconnect: 4) |   
Node: | 2 | SLAVE | received SM from | 2 | Request ID: | 0 | State: | UNASSIGNED | Calltype: NONE | Floor: 0 | Direction: NONE |  
Node: | 1 | MASTER | has GQ:  
Entry: | 1 | State: | ACTIVE | Calltype: CAB | Floor: 1 | Direction: NONE | Reassigned in: 5.00 | Requested node: | 1 | Assigned node: | 1 |   
Entry: | 4 | State: | ACTIVE | Calltype: CAB | Floor: 3 | Direction: NONE | Reassigned in: 11.00 | Requested node: | 1 | Assigned node: | 2 |  
```
### 1.2.3. Solution
As the problem will likely be solved either whent Node 1 is once again terminated, and Node 2 is set to master, or when Node 2 is terminated, this problem is not critical.  
The proposed solution however is changing the functions,

```go  

func f_UpdateGlobalQueueMaster(c_getSetGlobalQueueInterface chan T_GetSetGlobalQueueInterface, getSetGlobalQueueInterface T_GetSetGlobalQueueInterface, masterMessage T_MasterMessage) {
	for _, remoteEntry := range masterMessage.GlobalQueue {
		f_AddEntryGlobalQueue(c_getSetGlobalQueueInterface, getSetGlobalQueueInterface, remoteEntry)
	}
}
func f_UpdateGlobalQueueSlave(c_getSetGlobalQueueInterface chan T_GetSetGlobalQueueInterface, getSetGlobalQueueInterface T_GetSetGlobalQueueInterface, masterMessage T_MasterMessage) {
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
```
and combining them into a singular function:

```go  

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
```

## Wont find ack if order in the same etasje -> fails to assign any further
Assigned request with ID: 1 assigned to node 1
2024/03/10 18:17:00 log_functions.go:15: Getting ack from last assinged...
2024/03/10 18:17:00 log_functions.go:15: Found assigned request with ID: 1 assigned to node 1
2024/03/10 18:17:00 log_functions.go:15: Found assigned entry!
2024/03/10 18:17:00 log_functions.go:15: Node: | 1 | request resent ACTIVE
2024/03/10 18:17:00 log_functions.go:15: Node: | 1 | request resent DONE
2024/03/10 18:17:00 log_functions.go:15: Node: | 1 | MASTER | updated GQ entry:

### Solution
Adding a channel for sending ack upon receiving messages as well (in addition to just polling data from CN)