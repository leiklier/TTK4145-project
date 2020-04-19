package nextfloor

import (
	"fmt"

	"../../network/peers"
	"../elevators"
	"../store"
)

// GetNextFloor returns the nextfloor the elevator should travel to.
// Returns -1 if there are no more orders to take
func GetNextFloor() int {
	cabCalls, _ := store.GetAllCabCalls(peers.GetRelativeTo(peers.Self, 0))
	hallCalls, _ := store.GetAllHallCalls(peers.GetRelativeTo(peers.Self, 0))
	currentDirection, _ := store.GetDirectionMoving(peers.GetRelativeTo(peers.Self, 0))
	currFloor, _ := store.GetCurrentFloor(peers.GetRelativeTo(peers.Self, 0))

	switch currentDirection {
	case elevators.DirectionDown:
		nf := searchUnderneath(currFloor, hallCalls, cabCalls)
		if nf != -1 {
			return nf
		}

	case elevators.DirectionUp:
		nf := searchAbove(currFloor, hallCalls, cabCalls)
		if nf != -1 {
			return nf
		}

	case elevators.DirectionIdle:
		nfAbove := searchAbove(currFloor, hallCalls, cabCalls)
		nfUnderneath := searchUnderneath(currFloor, hallCalls, cabCalls)

		if nfAbove != -1 || nfUnderneath != -1 {
			if abs(currFloor-nfUnderneath) < abs(currFloor-nfAbove) {
				return nfUnderneath
			} else {
				return nfAbove
			}
		}

	default:
		fmt.Println("Get the Bible and pray!")
	}
	return -1
}

// Returns nextFloor. If there are no more orders, it returns -1
func searchAbove(currFloor int, hallCalls []elevators.HallCall_s, cabCalls []bool) int {
	nextCab := store.NumFloors + 1 // Init with too high
	for i, cab := range cabCalls {
		if i <= currFloor {
			continue // Check only above
		} else if cab {
			nextCab = i
			break
		}
	}
	for i, hc := range hallCalls {
		if i <= currFloor {
			continue // Check only above
		} else {
			if hc.Direction != elevators.DirectionIdle {
				nextHall := i
				if nextHall < nextCab {
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
	return nextCab

}

// Returns nextFloor. If there are no more orders, it returns -1
func searchUnderneath(currFloor int, hallCalls []elevators.HallCall_s, cabCalls []bool) int {
	nextCab := -1 // Init with below surface
	for i := store.NumFloors - 1; i >= 0; i-- {
		if i >= currFloor {
			continue // Check only downwards
		} else {
			if cabCalls[i] {
				nextCab = i
				break
			}
		}
	}
	for i := store.NumFloors - 1; i >= 0; i-- {
		if i >= currFloor {
			continue // Check only downwards
		} else {
			if hallCalls[i].Direction != elevators.DirectionIdle {
				nextHall := i
				if nextHall > nextCab {
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
	return nextCab
}

// Go has no built in absolute value function for integers
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
