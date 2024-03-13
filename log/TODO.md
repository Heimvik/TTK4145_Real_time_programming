# TODO list
- Constants with FIRSTWORD_SECONDWORD
- Move pricess pair to main
- Arbo/vi mekker testliste
- Prerequistes and description for all functions
- Navngivning av variabler, funsjoner og filer

# Triple Elevator System Test Cases

## Main Requirements

### Button Lights Service Guarantee

#### Hall Call Buttons
- [ ] Check if an elevator arrives at the floor once the light on a hall call button is turned on.
- [ ] Ensure lights on hall call buttons show the same status on all workspaces under normal circumstances.

#### Cab Call Buttons
- [ ] Verify that the elevator at a specific workspace takes the cab call order.
- [ ] Ensure cab button lights do not share between workspaces.
- [ ] Test if cab button lights turn on as soon as the button is pressed and off after the call is serviced.

### No Calls Lost
- [ ] Test for loss of network connection.
- [ ] Verify software crash scenarios.
- [ ] Ensure doors that won't close are handled.
- [ ] Check for power loss to elevator motor and controlling machine.
- [ ] Validate that losing network packet is not considered a failure.

### Elevator Functionality
- [ ] Test for elevator disconnect from the network while serving active calls.
- [ ] Verify the elevator software does not require manual restart after intermittent network or power loss.
- [ ] Ensure no stopping at every floor unnecessarily.
- [ ] Check if an elevator arriving at a floor clears the corresponding call button light.

### Door Functionality
- [ ] Verify the door open lamp behavior.
- [ ] Test if the door opens only when the elevator is stopped at a floor.
- [ ] Ensure the duration for keeping the door open is 3 seconds.
- [ ] Validate the obstruction switch behavior.

## Secondary Requirements

### Calls Distribution
- [ ] Test calls distribution across elevators to ensure efficient service.

## Permitted Assumptions

### Assumption 1
- [ ] Ensure at least one elevator is not in a failure state.
- [ ] Validate that no failure includes the door obstruction scenario.

### Assumption 2
- [ ] Verify cab call redundancy behavior with a single or disconnected elevator.

### Assumption 3
- [ ] Test for no network partitioning scenario.

## Unspecified Behavior

### Initialization
- [ ] Test behavior when the elevator cannot connect to the network during initialization.

### Hall Buttons
- [ ] Verify behavior of hall buttons when the elevator is disconnected from the network.

### Stop Button
- [ ] Test behavior of the stop button.



