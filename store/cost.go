package store

import (
	"fmt"
)

func DistributeHallCall(hc HallCall, ip string) {

}

/*

Måter vi kan distrubuere ordren på

1. Send rundt til alle



*/

// Returns Ip address of most suited elevator to handle hallcal
func MostSuitedElevator(hc HallCall, flflf int) string {
	// Steg 1: Gi ordren til heis uten calls, som er nærmest
	// Håndterer om det er idle, og gir til idle

	isClear := true
	var candidates []ElevatorState
	for _, elev := range GAllElevatorStates {
		hc_list := elev.Hall_calls
		dir := elev.GDirection
		if dir == DIR_idle {
			return elev.Ip
		}

		for _, hcElement := range hc_list {
			if hcElement.GDirection != (DIR_idle) {
				isClear = false
				break
			}
		}
		if isClear {
			candidates = append(candidates, elev)
		}
	}
	if isClear {
		currMaxDiff := numFloors + 1
		// Ip of closest elevator to origin floor. Default with err msg
		currCand := "Something went wrong"
		for _, elev := range candidates {
			floorDiff := abs(elev.Current_floor - hc.GFloor)
			if floorDiff < currMaxDiff {
				currMaxDiff = floorDiff
				currCand = elev.Ip
				//fmt.Println("Kom inn i isClear")
			}
		}
		return currCand
	} else {
		//fmt.Println("Kom inn i steg 2")

		// Steg 2
		currMaxFS := 0
		var currentMax string // Ip of elevator with highest FSvalue

		for _, elev := range GAllElevatorStates {
			// Extract elevator information
			currFloor := elev.Current_floor
			elevDir := elev.GDirection

			// hcDir := HCDirToElevDir(hc)

			var sameDir bool
			if hc.GDirection == elevDir {
				sameDir = true
			} else {
				sameDir = false
			}
			floorDiff := abs(currFloor - hc.GFloor)

			var goingTowards bool
			if (currFloor-hc.GFloor) > 0 && elevDir == DIR_down {
				goingTowards = true
			} else if (currFloor-hc.GFloor) < 0 && elevDir == DIR_down {
				goingTowards = false
			} else if (currFloor-hc.GFloor) > 0 && elevDir == DIR_up {
				goingTowards = false
			} else if (currFloor-hc.GFloor) < 0 && elevDir == DIR_up {
				goingTowards = true
			} else if (currFloor - hc.GFloor) == 0 {
				// Hmmmm, this means that it is at the same floor when button is pressed.
				// Extremely unlikely...
			} else {
				// Mby add default case to make sure that goingTowards has a value...
				// If for some reason the above expressions should fail,
				// we could just assume the worst and set goingTowards = false
			}

			var FS int
			// Computing FS Values based upon cases:
			if goingTowards && sameDir {
				FS = (numFloors - 1) + 2 - floorDiff
			} else if goingTowards && !sameDir {
				FS = (numFloors - 1) + 1 - floorDiff
			} else if !goingTowards {
				FS = 1
			}
			fmt.Println("FS Score of elevator", elev.Ip, "is:", FS)
			if FS > currMaxFS {
				currMaxFS = FS
				currentMax = elev.Ip
			}
		}

		return currentMax
	}
}

// Jakob todo: skriv ferdig kostfunksjonen, hvis ip == self.ip, kjør heis.

/////////////////////////////////////////////////
/// Helper functions, not to be exported.
/////////////////////////////////////////////////

// GO has no built in absolute value function for integers, so must create my own
// abs returns the absolute value of x.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Convert from HC Direction to Elevator Direction
func HCDirToElevDir(hc HallCallDir) Direction {
	switch hc {
	case HC_up:
		return DIR_up
	case HC_down:
		return DIR_down
	case HC_both:
		return DIR_both
	default: // Assumes HC_none
		return DIR_idle
	}
}
