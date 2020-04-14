package costfunction

import (
	"fmt"

	"../elevators"
)

// Returns IP address of most suited elevator to handle hallcal
func MostSuitedElevator(allElevators []elevators.Elevator_s, numFloors int, hcFloor int, hcDirection elevators.Direction_e) string {
	// Checking all elevators for any idle elevators
	areThereIdles := false
	var idleCandidates []elevators.Elevator_s

	isClear := true
	var clearCandidates []elevators.Elevator_s

	// Checking for IDLE elevators and elevators with NO hallcalls
	for _, elevator := range allElevators {
		hallCalls := elevator.GetAllHallCalls()
		directionMoving := elevator.GetDirectionMoving()

		// The first elevator it finds that is IDLE is assgined the HC
		// This must be fixed
		if directionMoving == elevators.DirectionIdle {
			areThereIdles = true
			idleCandidates = append(idleCandidates, elevator)
			//return elevator.GetHostname()
		}

		// We cycle hallCalls to see if there are any HC
		for _, hallCall := range hallCalls {
			if hallCall.Direction != elevators.DirectionIdle {
				isClear = false
				break
			}
		}
		// If the elevator is clear, it is added to clearCand
		if isClear {
			clearCandidates = append(clearCandidates, elevator)
		}
	}
	if areThereIdles {
		// This means there are IDLE elevators. We find the closest.
		currMaxDiff := numFloors + 1
		// Ip of closest elevator to origin floor. Default with err msg
		currCand := "Something went wrong"
		for _, elevator := range idleCandidates {
			floorDiff := abs(elevator.GetCurrentFloor() - hcFloor)
			if floorDiff < currMaxDiff {
				currMaxDiff = floorDiff
				currCand = elevator.GetHostname()
				//fmt.Println("Kom inn i isClear")
			}
		}
		return currCand
	}

	if isClear {
		// This means ALL elevators have NO hallCalls, and are all moving
		currMaxDiff := numFloors + 1
		// Ip of closest elevator to origin floor. Default with err msg
		currCand := "Something went wrong"
		for _, elevator := range clearCandidates {
			floorDiff := abs(elevator.GetCurrentFloor() - hcFloor)
			if floorDiff < currMaxDiff {
				currMaxDiff = floorDiff
				currCand = elevator.GetHostname()
				//fmt.Println("Kom inn i isClear")
			}
		}
		return currCand
	} else {
		// There are HC in at least one elevator that must be taken into consideration
		//fmt.Println("Kom inn i steg 2")

		// Steg 2
		currMaxFS := 0
		var currentMax string // Ip of elevator with highest FSvalue

		for _, elevator := range allElevators {
			// Extract elevator information
			currFloor := elevator.GetCurrentFloor()
			elevDir := elevator.GetDirectionMoving()

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
				// And is moving. Extremely unlikely...
				// In hindsight I've come to realise that this is not at all that unlikely
				// It should probably be fixed
				// I think setting false if the correction option here
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
			fmt.Println("FS Score of elevator", elevator.GetHostname(), "is:", FS)
			if FS > currMaxFS {
				currMaxFS = FS
				currentMax = elevator.GetHostname()
			}
		}

		return currentMax
	}
}

// Jakob todo: skriv ferdig kostfunksjonen, hvis ip == self.ip, kjør heis.

/////////////////////////////////////////////////
/// Helper functions, not to be exported.
/////////////////////////////////////////////////

// TIL TESTING: KJØR EN HEIS FRA 0 - 3. Legg inn HC/CC i 2 mens den kjører. Hva skjer?
// Possible bug: If elevator has many HC and is idle JUST as the incomming HC is pressed, it will be assigned if it is the closest
// Even if there are other moving elevators that may be a better option...

// GO has no built in absolute value function for integers, so must create my own
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
