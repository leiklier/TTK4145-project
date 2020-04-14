package main

import (
	"fmt"

	"../costfunction"
	"../elevators"
	"../store"
)

// --------- The hallcall to be assigned ---------//
var hcFloor = 2
var hcDir = elevators.DirectionIdle

// --------- Elevator states ---------//
var currFloor1 = 0
var prevFloor1 = 1 // Betyr ingenting, brukes ikke!
var currDir1 = elevators.DirectionIdle

var currFloor2 = 3
var prevFloor2 = 2 // Betyr ingenting, brukes ikke!
var currDir2 = elevators.DirectionIdle

// --------- Setting up hallcalls --------- //

var hc1_0 = elevators.HallCall_s{Floor: 0, Direction: elevators.DirectionIdle}
var hc1_1 = elevators.HallCall_s{Floor: 1, Direction: elevators.DirectionDown}
var hc1_2 = elevators.HallCall_s{Floor: 2, Direction: elevators.DirectionIdle}
var hc1_3 = elevators.HallCall_s{Floor: 3, Direction: elevators.DirectionIdle}

var hc2_0 = elevators.HallCall_s{Floor: 0, Direction: elevators.DirectionUp}
var hc2_1 = elevators.HallCall_s{Floor: 1, Direction: elevators.DirectionIdle}
var hc2_2 = elevators.HallCall_s{Floor: 2, Direction: elevators.DirectionDown}
var hc2_3 = elevators.HallCall_s{Floor: 3, Direction: elevators.DirectionIdle}

var hcs1 = [store.NumFloors]elevators.HallCall_s{hc1_0, hc1_1, hc1_2, hc1_3}
var hcs2 = [store.NumFloors]elevators.HallCall_s{hc2_0, hc2_1, hc2_2, hc2_3}

var cc = [store.NumFloors]bool{false, false, false, false}

// {"1", currFloor1, store.NumFloors, prevFloor1, currDir1, hcs1, cc, true}

//{"2", currFloor2, store.NumFloors, prevFloor2, currDir2, hcs2, cc, true}

//var allStates = []elevators.Elevator_s{elev1, elev2}

func main() {
	// --------- Initiating elevators --------- //
	elev1 := elevators.New("1", store.NumFloors, currFloor1)
	store.SetDirectionMoving("1", currDir1)
	for _, i := range hcs1 {
		store.AddHallCall("1", i)
	}
	elev2 := elevators.New("2", store.NumFloors, currFloor2)
	store.SetDirectionMoving("2", currDir2)
	for _, i := range hcs2 {
		store.AddHallCall("2", i)
	}

	var allStates = []elevators.Elevator_s{elev1, elev2}

	fmt.Println("The test scenario is described by: ")
	fmt.Println("Elevator 1 is at floor", currFloor1, "going", dirToText(currDir1))
	fmt.Println("Elevator 2 is at floor", currFloor2, "going", dirToText(currDir2))
	fmt.Println("The incomming hallcall is at floor:", hcFloor, "wanting to go:", dirToText(hcDir))
	fmt.Println()

	ip := costfunction.MostSuitedElevator(allStates, store.NumFloors, hcFloor, hcDir)
	fmt.Println("Most suited elevator for this scenario is elevavtor nr:", ip)
}

func dirToText(dir elevators.Direction_e) string {
	switch dir {
	case elevators.DirectionUp:
		return "up"
	case elevators.DirectionDown:
		return "down"
	case elevators.DirectionIdle:
		return "nowhere"
	default:
		return "both"
	}
}
