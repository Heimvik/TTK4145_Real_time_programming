# Known issues:
## Disallow backward information in reassign
Upon having both nodes connected, disconnecting master, connecting master, we get the log: 
    Node: | 1 | MASTER | has GQ:
    2024/03/09 14:56:36 log_functions.go:15: Entry: | 3 | State: | ACTIVE | Calltype: HALL | Floor: 1 | Direction: UP | Reassigned in: 2.00 | Requested node: | 2 | Assigned node: | 1 | 
    2024/03/09 14:56:36 log_functions.go:15: Entry: | 4 | State: | ACTIVE | Calltype: HALL | Floor: 2 | Direction: UP | Reassigned in: 8.00 | Requested node: | 2 | Assigned node: | 1 | 
    2024/03/09 14:56:36 log_functions.go:15: Entry: | 5 | State: | ACTIVE | Calltype: HALL | Floor: 3 | Direction: UP | Reassigned in: 10.00 | Requested node: | 2 | Assigned node: | 2 | 
    2024/03/09 14:56:36 log_functions.go:15: Node: | 2 | SLAVE | received MM from | 1 | MASTER | GlobalQueue: [Request ID: | 3 | State: | ACTIVE | Calltype: HALL | Floor: 1 | Direction: UP | Reassigned in: 1 | Requested node: | 2 | Assigned node: | 1 |, Request ID: | 4 | State: | ACTIVE | Calltype: HALL | Floor: 2 | Direction: UP | Reassigned in: 7 | Requested node: | 2 | Assigned node: | 1 |, Request ID: | 5 | State: | ACTIVE | Calltype: HALL | Floor: 3 | Direction: UP | Reassigned in: 9 | Requested node: | 2 | Assigned node: | 2 |]
    2024/03/09 14:56:36 log_functions.go:15: Node: | 1 | MASTER | received MM from | 1 | MASTER | GlobalQueue: [Request ID: | 3 | State: | ACTIVE | Calltype: HALL | Floor: 1 | Direction: UP | Reassigned in: 1 | Requested node: | 2 | Assigned node: | 1 |, Request ID: | 4 | State: | ACTIVE | Calltype: HALL | Floor: 2 | Direction: UP | Reassigned in: 7 | Requested node: | 2 | Assigned node: | 1 |, Request ID: | 5 | State: | ACTIVE | Calltype: HALL | Floor: 3 | Direction: UP | Reassigned in: 9 | Requested node: | 2 | Assigned node: | 2 |]
    2024/03/09 14:56:36 log_functions.go:15: Node: | 1 | MASTER | has connected nodes | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) | 2 (Role: SLAVE, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 3) | 
    2024/03/09 14:56:36 log_functions.go:15: Node: | 2 | SLAVE | has GQ:
    2024/03/09 14:56:36 log_functions.go:15: Entry: | 3 | State: | ACTIVE | Calltype: HALL | Floor: 1 | Direction: UP | Reassigned in: 1.00 | Requested node: | 2 | Assigned node: | 1 | 
    2024/03/09 14:56:36 log_functions.go:15: Entry: | 4 | State: | ACTIVE | Calltype: HALL | Floor: 2 | Direction: UP | Reassigned in: 7.00 | Requested node: | 2 | Assigned node: | 1 | 
    2024/03/09 14:56:36 log_functions.go:15: Entry: | 5 | State: | ACTIVE | Calltype: HALL | Floor: 3 | Direction: UP | Reassigned in: 9.00 | Requested node: | 2 | Assigned node: | 2 | 
    2024/03/09 14:56:36 log_functions.go:15: Node: | 1 | MASTER | received SM from | 2 | Request ID: | 0 | State: | UNASSIGNED | Calltype: NONE | Floor: 0 | Direction: NONE |
    2024/03/09 14:56:36 log_functions.go:15: Node: | 1 | MASTER | has connected nodes | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) | 2 (Role: SLAVE, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) | 
    2024/03/09 14:56:36 log_functions.go:15: Node: | 2 | SLAVE | received SM from | 2 | Request ID: | 0 | State: | UNASSIGNED | Calltype: NONE | Floor: 0 | Direction: NONE |
    2024/03/09 14:56:37 log_functions.go:15: Reassigned entry: | 3 | 2 | in global queue
    2024/03/09 14:56:37 log_functions.go:15: Node: | 2 | SLAVE | received MM from | 1 | MASTER | GlobalQueue: [Request ID: | 3 | State: | UNASSIGNED | Calltype: HALL | Floor: 1 | Direction: UP | Reassigned in: 15 | Requested node: | 2 | Assigned node: | 0 |, Request ID: | 4 | State: | ACTIVE | Calltype: HALL | Floor: 2 | Direction: UP | Reassigned in: 6 | Requested node: | 2 | Assigned node: | 1 |, Request ID: | 5 | State: | ACTIVE | Calltype: HALL | Floor: 3 | Direction: UP | Reassigned in: 8 | Requested node: | 2 | Assigned node: | 2 |]
    2024/03/09 14:56:37 log_functions.go:15: Disallowed backward information
    2024/03/09 14:56:37 log_functions.go:15: Node: | 1 | MASTER | received MM from | 1 | MASTER | GlobalQueue: [Request ID: | 3 | State: | UNASSIGNED | Calltype: HALL | Floor: 1 | Direction: UP | Reassigned in: 15 | Requested node: | 2 | Assigned node: | 0 |, Request ID: | 4 | State: | ACTIVE | Calltype: HALL | Floor: 2 | Direction: UP | Reassigned in: 6 | Requested node: | 2 | Assigned node: | 1 |, Request ID: | 5 | State: | ACTIVE | Calltype: HALL | Floor: 3 | Direction: UP | Reassigned in: 8 | Requested node: | 2 | Assigned node: | 2 |]
    2024/03/09 14:56:37 log_functions.go:15: Node: | 1 | MASTER | has connected nodes | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) | 2 (Role: SLAVE, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 3) | 
    2024/03/09 14:56:37 log_functions.go:15: Node: | 1 | MASTER | received SM from | 2 | Request ID: | 0 | State: | UNASSIGNED | Calltype: NONE | Floor: 0 | Direction: NONE |
    2024/03/09 14:56:37 log_functions.go:15: Node: | 1 | MASTER | has connected nodes | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) | 2 (Role: SLAVE, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) | 
    2024/03/09 14:56:37 log_functions.go:15: Node: | 2 | SLAVE | received SM from | 2 | Request ID: | 0 | State: | UNASSIGNED | Calltype: NONE | Floor: 0 | Direction: NONE |
    2024/03/09 14:56:38 log_functions.go:15: Node: | 1 | MASTER | has GQ:
    2024/03/09 14:56:38 log_functions.go:15: Entry: | 3 | State: | UNASSIGNED | Calltype: HALL | Floor: 1 | Direction: UP | Reassigned in: 15.00 | Requested node: | 2 | Assigned node: | 0 | 
    2024/03/09 14:56:38 log_functions.go:15: Entry: | 4 | State: | ACTIVE | Calltype: HALL | Floor: 2 | Direction: UP | Reassigned in: 6.00 | Requested node: | 2 | Assigned node: | 1 | 
    2024/03/09 14:56:38 log_functions.go:15: Entry: | 5 | State: | ACTIVE | Calltype: HALL | Floor: 3 | Direction: UP | Reassigned in: 8.00 | Requested node: | 2 | Assigned node: | 2 | 
    2024/03/09 14:56:38 log_functions.go:15: Node: | 2 | SLAVE | received MM from | 1 | MASTER | GlobalQueue: [Request ID: | 3 | State: | UNASSIGNED | Calltype: HALL | Floor: 1 | Direction: UP | Reassigned in: 15 | Requested node: | 2 | Assigned node: | 0 |, Request ID: | 4 | State: | ACTIVE | Calltype: HALL | Floor: 2 | Direction: UP | Reassigned in: 5 | Requested node: | 2 | Assigned node: | 1 |, Request ID: | 5 | State: | ACTIVE | Calltype: HALL | Floor: 3 | Direction: UP | Reassigned in: 7 | Requested node: | 2 | Assigned node: | 2 |]
    2024/03/09 14:56:38 log_functions.go:15: Node: | 1 | MASTER | received MM from | 1 | MASTER | GlobalQueue: [Request ID: | 3 | State: | UNASSIGNED | Calltype: HALL | Floor: 1 | Direction: UP | Reassigned in: 15 | Requested node: | 2 | Assigned node: | 0 |, Request ID: | 4 | State: | ACTIVE | Calltype: HALL | Floor: 2 | Direction: UP | Reassigned in: 5 | Requested node: | 2 | Assigned node: | 1 |, Request ID: | 5 | State: | ACTIVE | Calltype: HALL | Floor: 3 | Direction: UP | Reassigned in: 7 | Requested node: | 2 | Assigned node: | 2 |]
    2024/03/09 14:56:38 log_functions.go:15: Node: | 1 | MASTER | has connected nodes | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) | 2 (Role: SLAVE, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 3) | 
    2024/03/09 14:56:38 log_functions.go:15: Disallowed backward information
    2024/03/09 14:56:38 log_functions.go:15: Node: | 2 | SLAVE | has GQ:
    2024/03/09 14:56:38 log_functions.go:15: Entry: | 3 | State: | ACTIVE | Calltype: HALL | Floor: 1 | Direction: UP | Reassigned in: 1.00 | Requested node: | 2 | Assigned node: | 1 | 
    2024/03/09 14:56:38 log_functions.go:15: Entry: | 4 | State: | ACTIVE | Calltype: HALL | Floor: 2 | Direction: UP | Reassigned in: 5.00 | Requested node: | 2 | Assigned node: | 1 | 
    2024/03/09 14:56:38 log_functions.go:15: Entry: | 5 | State: | ACTIVE | Calltype: HALL | Floor: 3 | Direction: UP | Reassigned in: 7.00 | Requested node: | 2 | Assigned node: | 2 | 
    2024/03/09 14:56:38 log_functions.go:15: Node: | 1 | MASTER | received SM from | 2 | Request ID: | 0 | State: | UNASSIGNED | Calltype: NONE | Floor: 0 | Direction: NONE |
    2024/03/09 14:56:38 log_functions.go:15: Node: | 1 | MASTER | has connected nodes | 1 (Role: MASTER, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) | 2 (Role: SLAVE, ElevatorInfo: {Direction:0 Floor:1 State:2}, TimeUntilDisconnect: 4) | 
    2024/03/09 14:56:38 log_functions.go:15: Node: | 2 | SLAVE | received SM from | 2 | Request ID: | 0 | State: | UNASSIGNED | Calltype: NONE | Floor: 0 | Direction: NONE |
    2024/03/09 14:56:38 log_functions.go:15: Node: | 1 | request resent DONE
    2024/03/09 14:56:38 log_functions.go:15: Assigned request with ID: 3 assigned to node 1

Aka we disallow backward information when a entry is reassigned. This fixes itself.


## Slave keeps spawning terminals when in slave

Slave keeps spawning terminals when in slave. Not sure of what causes, but problem lies in f_BackupManager.