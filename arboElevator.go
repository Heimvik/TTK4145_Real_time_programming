package main

import (
	"fmt"
	"sync"
	"time"
)

// ElevatorDirection represents the direction the elevator is moving.
type ElevatorDirection int

const (
	Up   ElevatorDirection = 1
	Down ElevatorDirection = -1
	Stop ElevatorDirection = 0
)

// Request represents a floor request for the elevator.
type Request struct {
	Floor int
}

// Elevator represents the state of the elevator.
type Elevator struct {
	sync.Mutex
	CurrentFloor int
	Direction    ElevatorDirection
	Requests     []Request
}

// Building represents the entire building with a single elevator.
type Building struct {
	Elevator Elevator
}

// NewBuilding creates a new building with a single elevator.
func NewBuilding() Building {
	return Building{
		Elevator: Elevator{
			CurrentFloor: 1,
			Direction:    Stop,
		},
	}
}

// RequestElevator sends a floor request to the elevator.
func (b *Building) RequestElevator(floor int) {
	request := Request{
		Floor: floor,
	}

	b.Elevator.Lock()
	defer b.Elevator.Unlock()

	b.Elevator.Requests = append(b.Elevator.Requests, request)
}

// MoveElevator simulates the movement of the elevator.
func (b *Building) MoveElevator() {
	for {
		time.Sleep(time.Second) // Simulate time passing

		b.Elevator.Lock()

		if len(b.Elevator.Requests) > 0 {
			nextFloor := b.Elevator.Requests[0].Floor
			if nextFloor > b.Elevator.CurrentFloor {
				b.Elevator.Direction = Up
				b.Elevator.CurrentFloor++
			} else if nextFloor < b.Elevator.CurrentFloor {
				b.Elevator.Direction = Down
				b.Elevator.CurrentFloor--
			} else {
				b.Elevator.Direction = Stop
				b.Elevator.Requests = b.Elevator.Requests[1:]
			}

			fmt.Printf("Elevator is at floor %d, moving %v to floor %d\n",
				b.Elevator.CurrentFloor, b.Elevator.Direction, nextFloor)
		}

		b.Elevator.Unlock()
	}
}

// func main() {
// 	building := NewBuilding()

// 	// Start a goroutine to simulate elevator movement
// 	go building.MoveElevator()

// 	// Request the elevator to different floors
// 	building.RequestElevator(5)
// 	building.RequestElevator(2)
// 	building.RequestElevator(8)
// 	building.RequestElevator(1)

// 	// Keep the program running to observe the elevator simulation
// 	select {}
// }
