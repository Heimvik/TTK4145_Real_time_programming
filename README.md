# the-elevator
Real-time program of the triple-elevator system | Group: Tuesday 12:15 - 16:00 13

## Code-specific navigation and linguistics

### The list below includes frequently used terms, provided with brief explanations of how these terms should be understood within the code
- "Elevators" refer to the physical piece of hardware that builds the elevator setup with motors, buttons, lights etc.
- "FSM" (Finite State Machine) refers to the organization that holds different states of the elevator.
- "Nodes" refer to the controller which controls the elevator, and communicates with other nodes.
- "Slave" and "Master" refers to the relationship between nodes. One master-node controls several slave-nodes, i.e. each node holds only one role, which is either slave or master.
- "Priority" refers to the identification-number of a node, and is used to distribute slave -and master-roles. The smaller priority-number, the higher prioritized node.
- "Primary" and "Backup" refers to the two routines running in one single node. The primary routine communicates with the other primary routines of other nodes, while the backup-routine only communicates with the primary routine within the same node. Each node should run their own primary and backup routine.
- "Global Queue" refers to the synchronization element between nodes, where queue-elements known as "entries" are added *to* and distributed *from* the global queue by the master.
- "Requests" refers to the different requests one can demand from the elevator via buttons: Hall-Down-requests, Hall-Up-requests and Cab-requests.
- "Log" refers to a development tool used to track system activity real time.  

### To simplify navigation within the project file, this brief description of the project structure should give insight in the various folders  
The "config"-folder holds the "default.json"-file used to configure the operation of an individual node. The only parameters one would need to adjust for expected functionality are "elevatorport" which must correspond to the desired elevator setup, and "priority" which should be unique to each individual node, and should not exceed the "nodes"-parameter. Also, make sure "slaveport" and "masterport" is not used by other users of the elevatorserver. The "log"-folder holds the development tool "debug1.log", which tracks the system activity, i.e. message passing, node-statistics etc.. The "network"-folder holds various handout-code. The "node"-folder holds node-specific files, e.g. "node-functions", which holds global node-specific functions, or "run-node" which holds the local variables and is called on to run a single node. Within the "node-folder" lies the "elevator"-folder which similarily is sorted with e.g. a run-file, functions-file and other elevator-specific code. Outside of the directories lies the "main"-function, which initializes and run the entire program when called upon. The process pair also requires the terminal program "gnu-terminal" installed to work properly.

### An update with an executable file is expected

## Prefixes
### Prefixes are used to abbreviate variable names, and are constant throughout the project-file.
In general, capital prefixes symbolizes global function/channel/type/pointer, while a lowercase prefix symbolizes local. NOTE: No type or pointer is defined locally.

- f: Local function, *f_HandleStopEvent()*
- F: Global function, *F_RunNode()*

- c: Local channel, *c_getSetElevatorInterface*
- C: Global channel, *C_get, C_set*

- T: Global type, *T_Request*

- P: Global pointer, *P_info*
