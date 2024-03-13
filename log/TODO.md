# TODO list
## Generelt
- [ ] Dobbeltsjekk at testlista dekker alt
- [ ] Simulate packet loss as described on [this site](https://medium.com/@adilrk/network-tools-19a12519737b).
## Node
- [x] Constants with FIRSTWORD_SECONDWORD
- [x] Move pricess pair to main
- [x] Prerequistes and description for all functions
- [x] Navngivning av variabler, funsjoner og filer
## Elevator
- [x] Constants with FIRSTWORD_SECONDWORD
- [x] Navngivning av enums.
- [x] Navngivning av funksjoner.
- [x] Navngivning av variabler.
- [x] Navngivning av filer.
- [x] Funksjonsstruktur
- [x] Prerequistes and description for all functions
- [ ] Legge til kilde der jeg bruker elevio.go
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
- [ ] Test behavior of the stop button (NOT!!).

# Navngivning

## Naming different types of variables
- If you modify a name with a qualifier like Total, Sum, Average, Max, Min, Record, String, or Pointer, , put the modifier at the end of the name

- Use opposites precisely. Using naming conventions for opposites helps consistency, which helps readability. Pairs like begin/end are easy to understand and remember

- Short loops: i, j, k. Longer loops i.e. nested loops: find better name

- Placeholders: avoid altogether, no names like "temp", instead: give them descriptive names to communicate task-understanding.

- Boolean variables: Should imply true or false. done, error, found, success, ok. if (isFound) less readable than if (found). Avoid negative variable names: notFound, notDone

- Enums: Can represent types, classes or constants, use naming convention accordingly. I.e. Color_Red works for a type, but not for a class: Color.Color_Red. Enums in plural: T_TypeNames 

## Naming Conventions
- Prefixes (det vi bruker allerede)

## Abreviations

- Remove all nonleading vowels. (computer becomes cmptr, and screen becomes
scrn)

- Use every significant word in the name, up to a maximum of three words.

- Remove useless suffixes—ing, ed, and so on

- Document all abbreviations in a project-level “Standard Abbreviations” document

- Avoid misleading names or abbreviations

- Avoid names with similar meanings

- Avoid variables with different meanings but similar names

- Avoid names that sound similar, such as wrap and rap

- Avoid numerals in names

- Avoid misspelled words in names: Higlight to Hilite (avoid this)

- Avoid words that are commonly misspelled in English

- Don’t differentiate variable names solely by capitalization

- Avoid multiple natural languages

- Avoid the names of standard types, variables, and routines (if, and, then ...)

## Checklist Navngivning

### General Naming Considerations
- Does the name fully and accurately describe what the variable represents?
- Does the name refer to the real-world problem rather than to the programming-language solution?
- Is the name long enough that you don’t have to puzzle it out?
- Are computed-value qualifiers, if any, at the end of the name?
- Does the name use Count or Index instead of Num?
### Naming Specific Kinds of Data
- Are loop index names meaningful (something other than i, j, or k if the
loop is more than one or two lines long or is nested)?
- Have all “temporary” variables been renamed to something more meaningful?
- Are boolean variables named so that their meanings when they’re true are
clear?
- Do enumerated-type names include a prefix or suffix that indicates the category—for example, Color_ for Color_Red, Color_Green, Color_Blue, and so
on?
- Are named constants named for the abstract entities they represent rather
than the numbers they refer to?
Naming Conventions
- Does the convention distinguish among local, class, and global data?
- Does the convention distinguish among type names, named constants,
enumerated types, and variables?
- Does the convention identify input-only parameters to routines in languages that don’t enforce them?
- Is the convention as compatible as possible with standard conventions for
the language?
- Are names formatted for readability?
## Short Names
- Does the code use long names (unless it’s necessary to use short ones)?
- Does the code avoid abbreviations that save only one character?
- Are all words abbreviated consistently?


