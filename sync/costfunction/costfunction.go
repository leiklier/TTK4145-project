package costfunction

import (
	"fmt"

	"../elevators"
)

// Returns IP address of most suited elevator to handle hallcal
func MostSuitedElevator(allElevators []elevators.Elevator_s, numFloors int, hcFloor int, hcDirection elevators.HCDirection_e) string {
	// Steg 1: Gi ordren til heis uten calls, som er nærmest
	// Håndterer om det er idle, og gir til idle

	isClear := true
	var candidates []elevators.Elevator_s
	for _, elevator := range allElevators {
		hallCalls := elevator.GetHallCalls()
		directionMoving := elevator.GetDirectionMoving()

		if directionMoving == elevators.DirectionIdle {
			return elevator.GetIP()
		}

		for _, hallCall := range hallCalls {
			if hallCall.Direction != elevators.DirectionIdle {
				isClear = false
				break
			}
		}
		if isClear {
			candidates = append(candidates, elevator)
		}
	}
	if isClear {
		currMaxDiff := numFloors + 1
		// Ip of closest elevator to origin floor. Default with err msg
		currCand := "Something went wrong"
		for _, elevator := range candidates {
			floorDiff := abs(elevator.GetCurrentFloor() - hcFloor)
			if floorDiff < currMaxDiff {
				currMaxDiff = floorDiff
				currCand = elevator.GetIP()
				//fmt.Println("Kom inn i isClear")
			}
		}
		return currCand
	} else {
		//fmt.Println("Kom inn i steg 2")

		// Steg 2
		currMaxFS := 0
		var currentMax string // Ip of elevator with highest FSvalue

		for _, elevator := range allElevators {
			// Extract elevator information
			currFloor := elevator.GetCurrentFloor()
			elevDir := elevator.GetDirectionMoving()

			// hcDir := HCDirToElevDir(hc)

			var sameDir bool
			if hcDirection == elevDir {
				sameDir = true
			} else {
				sameDir = false
			}
			floorDiff := abs(currFloor - hcFloor)

			var goingTowards bool
			if (currFloor-hcFloor) > 0 && elevDir == elevators.DirectionDown {
				goingTowards = true
			} else if (currFloor-hcFloor) < 0 && elevDir == elevators.DirectionDown {
				goingTowards = false
			} else if (currFloor-hcFloor) > 0 && elevDir == elevators.DirectionUp {
				goingTowards = false
			} else if (currFloor-hcFloor) < 0 && elevDir == elevators.DirectionUp {
				goingTowards = true
			} else if (currFloor - hcFloor) == 0 {
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
			fmt.Println("FS Score of elevator", elevator.GetIP(), "is:", FS)
			if FS > currMaxFS {
				currMaxFS = FS
				currentMax = elevator.GetIP()
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
