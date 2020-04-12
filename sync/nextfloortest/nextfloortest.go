package main

import (
	"fmt"

	"../elevators"
	"../store"
)

//////////       My elevator       //////////

var hc0 = elevators.HallCall_s{Floor: 0, Direction: elevators.DirectionIdle}
var hc1 = elevators.HallCall_s{Floor: 1, Direction: elevators.DirectionIdle}
var hc2 = elevators.HallCall_s{Floor: 2, Direction: elevators.DirectionIdle}
var hc3 = elevators.HallCall_s{Floor: 3, Direction: elevators.DirectionIdle}

var hallCalls = [store.NumFloors]elevators.HallCall_s{hc0, hc1, hc2, hc3}
var cabCalls = [store.NumFloors]bool{true, false, false, false}

//var currDir = elevators.DirectionDown
var currFloor = 3

func main() {
	fmt.Println("Welcome to some serious testing!")
	fmt.Println()

	fmt.Println("Testing the downwards loops:")
	displayScenario(elevators.DirectionDown, currFloor)
	downresult := downWardsLoops(currFloor, hallCalls, cabCalls)
	fmt.Println("I got: ", downresult)
	fmt.Println("*-------------------------------------------------------*")

	fmt.Println("Testing the upwards loops: ")
	displayScenario(elevators.DirectionUp, 2)
	upresult := upwardsLoops(2, hallCalls, cabCalls)
	fmt.Println("I got:", upresult)
	fmt.Println("*-------------------------------------------------------*")

	fmt.Println("Testing the Idle loops:")
	displayScenario(elevators.DirectionIdle, 2)
	idleresult := bothLoop(2, hallCalls, cabCalls)
	fmt.Println("I got:", idleresult)

}

func backwardsFoorLoop(n int) {
	for i := n - 1; i >= 0; i-- {
		fmt.Println(i)
	}
}

func displayScenario(dirr elevators.Direction_e, currFloor int) {
	fmt.Println("Location:", currFloor)
	switch dirr {
	case elevators.DirectionDown:
		fmt.Println("Heading: Downwards")
	case elevators.DirectionUp:
		fmt.Println("Heading: Upwards")
	case elevators.DirectionIdle:
		fmt.Println("Heading: Standing still")
	default:
		fmt.Println("Get the Bible")
	}
	fmt.Println("HC:", hallCalls)
	fmt.Println("CC: ", cabCalls)

}

func upwardsLoops(currFloor int, hallCalls [store.NumFloors]elevators.HallCall_s, cabCalls [store.NumFloors]bool) int {
	nextCab := store.NumFloors + 1 // Way too high
	// Looking for hot single moms upstairs...
	for i, cab := range cabCalls {
		if i <= currFloor {
			continue // Vi sjekker kun oppover¨
		} else if cab {
			nextCab = i
			break
		}
	}
	for i, hc := range hallCalls {
		if i <= currFloor {
			continue // Vi sjekker kun oppover
		} else {
			if hc.Direction != elevators.DirectionIdle {
				nextHall := i
				if nextHall < nextCab {
					fmt.Println("Returning hallCall")
					return nextHall
				}
				break
			}
		}
	}
	if nextCab == store.NumFloors+1 {
		// No calls
		return -1
	}
	fmt.Println("Returning cabCall")
	return nextCab

}

func downWardsLoops(currFloor int, hallCalls [store.NumFloors]elevators.HallCall_s, cabCalls [store.NumFloors]bool) int {
	nextCab := -1 // Below surface
	// Looking for hot single moms downstairs ...
	for i := store.NumFloors - 1; i >= 0; i-- {
		if i >= currFloor {
			continue // Sjekker kun nedover
		} else {
			if cabCalls[i] {
				nextCab = i
				break
			}
		}
	}
	for i := store.NumFloors - 1; i >= 0; i-- {
		if i >= currFloor {
			continue // Sjekker kun nedover
		} else {
			if hallCalls[i].Direction != elevators.DirectionIdle {
				nextHall := i
				if nextHall > nextCab {
					fmt.Println("Returning hallCall")
					return nextHall
				}
				break
			}
		}
	}
	if nextCab == -1 {
		// No calls
		return -1
	}
	fmt.Println("Returning cabCall")
	return nextCab
}

func bothLoop(currFloor int, hallCalls [store.NumFloors]elevators.HallCall_s, cabCalls [store.NumFloors]bool) int {
	for counter := 1; counter < store.NumFloors; counter++ {
		lowerSearchIndex := currFloor - counter  // Sjekk om er under 0
		higherSearchIndex := currFloor + counter // Sjekk om er over NumFloors-1
		fmt.Println()
		fmt.Println("Iteration nr", counter)
		fmt.Println("lowerIndex:", lowerSearchIndex)
		fmt.Println("higheIndex", higherSearchIndex)

		if !(lowerSearchIndex < 0) {
			if cabCalls[lowerSearchIndex] {
				fmt.Println("Returning cabCall")
				return lowerSearchIndex
			}
			if hallCalls[lowerSearchIndex].Direction != elevators.DirectionIdle {
				fmt.Println("Returning hallCall")
				return lowerSearchIndex
			}
		}
		if !(higherSearchIndex > store.NumFloors-1) {
			if cabCalls[higherSearchIndex] {
				fmt.Println("Returning cabCall")
				return higherSearchIndex
			}
			if hallCalls[higherSearchIndex].Direction != elevators.DirectionIdle {
				fmt.Println("Returning hallCall")
				return higherSearchIndex
			}
		}

	}
	// No more calls
	fmt.Println("No more to take")
	return -1
}
