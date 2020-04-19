package nextfloor

import (
	"fmt"

	"../../network/peers"
	"../elevators"
	"../store"
)

// Forslag til searcboth: bruk både searchunderneath og searchabove og velg nærmeste nf. abs(curr -nf)
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
		// Denne kan egntlig håndere hele suppen hvis vi kun sjekker på idle
		nf := searchBoth(currFloor, hallCalls, cabCalls)
		// nfAbove := searchAbove(currFloor, hallCalls, cabCalls)
		// nfUnderneath := searchUnderneath(currFloor, hallCalls, cabCalls)

		if nf != -1 {
			return nf
		}

	default:
		fmt.Println("Get the Bible and pray!")
	}
	// Only rerun when store has changed:
	return -1
}

// Returns nextFloor. If there are no more orders, it returns -1
func searchAbove(currFloor int, hallCalls []elevators.HallCall_s, cabCalls []bool) int {
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
					fmt.Println("Nextfloor is hallcall at: ", nextHall)
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

// Returns nextFloor. If there are no more orders, it returns -1
func searchUnderneath(currFloor int, hallCalls []elevators.HallCall_s, cabCalls []bool) int {
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
					fmt.Println("Nextfloor is hallcall at: ", nextHall)
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

func searchBoth(currFloor int, hallCalls []elevators.HallCall_s, cabCalls []bool) int {
	for counter := 1; counter < store.NumFloors; counter++ {

		lowerSearchIndex := currFloor - counter  // Sjekk om er under 0
		higherSearchIndex := currFloor + counter // Sjekk om er over NumFloors-1

		if !(lowerSearchIndex < 0) {
			if cabCalls[lowerSearchIndex] {
				fmt.Println("Returning cabCall")
				return lowerSearchIndex
			}
			if hallCalls[lowerSearchIndex].Direction != elevators.DirectionIdle {
				fmt.Println("Nextfloor is hallcall at: ", lowerSearchIndex)
				return lowerSearchIndex
			}
		}
		if !(higherSearchIndex > store.NumFloors-1) {
			if cabCalls[higherSearchIndex] {
				fmt.Println("Returning cabCall")
				return higherSearchIndex
			}
			if hallCalls[higherSearchIndex].Direction != elevators.DirectionIdle {
				fmt.Println("Nextfloor is hallcall at: ", higherSearchIndex)
				return higherSearchIndex
			}
		}

	}
	// No more calls
	return -1
}
