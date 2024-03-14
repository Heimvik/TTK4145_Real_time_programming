# the-elevator
Real-time program of the triple-elevator system | Group: Tuesday 12:15 - 16:00 13

## Code-specific navigation and linguistics

- "Elevators" refer to the physical piece of hardware that builds the elevator setup with motors, buttons, lights etc.
- "FSM" (Finite State Machine) refers to the organization that holds different states of the elevator.
- "Nodes" refer to the controller which controls the elevator, and communicates with other nodes.
- "Priority" refers to the identification-number of a node. The smaller priority-number, the higher prioritized node. 
- "Global Queue" refers to the synchronization element between nodes, where queue-elements known as "entries" are added *to* and distributed *from*
- "Requests" refers to the different requests one can demand from the elevator via buttons: Hall-Down-requests, Hall-Up-requests and Cab-requests.
- "Log" refers to a development tool used to track system activity real time.  

The "config"-folder holds the "default.json"-file used to configure the operation of an individual node. The only parameters one would need to adjust for expected functionality are "elevatorport" which must correspond to the desired elevator setup, and "priority" which should be unique to each individual node, and should not exceed the "nodes"-parameter. The "log"-folder holds the development tool "debug1.log", which tracks the system activity, i.e. message passing, node-statistics . The "network"-folder holds



## Prefixes
In general, capital prefixes symbolizes global function/channel/type/pointer, while a lowercase prefix symbolizes local. NOTE: No type or pointer is defined locally.

- f: Local function, *f_HandleStopEvent()*
- F: Global function, *F_RunNode()*

- c: Local channel, *c_getSetElevatorInterface*
- C: Global channel, *C_get, C_set*

- T: Global type, *T_Request*

- P: Global pointer, *P_info*
