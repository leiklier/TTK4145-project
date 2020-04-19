package costfunction

import (
	"fmt"

	"../elevators"
)

// Returns hostname of most suited elevator to handle hallcal
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

		if directionMoving == elevators.DirectionIdle {
			areThereIdles = true
			idleCandidates = append(idleCandidates, elevator)
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
		// Hostname of closest elevator to origin floor. Default with err msg
		currCand := "Something went wrong"
		for _, elevator := range idleCandidates {
			floorDiff := abs(elevator.GetCurrentFloor() - hcFloor)
			if floorDiff < currMaxDiff {
				currMaxDiff = floorDiff
				currCand = elevator.GetHostname()
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
			}
		}
		return currCand
	} else {
		// There are HC in at least one elevator that must be taken into consideration

		currMaxFS := 0
		var currentMax string // Hostname of elevator with highest FSvalue

		for _, elevator := range allElevators {
			currFloor := elevator.GetCurrentFloor()
			elevDir := elevator.GetDirectionMoving()

			var sameDir bool
			if hcDirection == elevDir {
				sameDir = true
			} else {
				sameDir = false
			}
			floorDiff := abs(currFloor - hcFloor)

			goingTowards := false
			if (currFloor-hcFloor) > 0 && elevDir == elevators.DirectionDown {
				goingTowards = true
			} else if (currFloor-hcFloor) < 0 && elevDir == elevators.DirectionDown {
				goingTowards = false
			} else if (currFloor-hcFloor) > 0 && elevDir == elevators.DirectionUp {
				goingTowards = false
			} else if (currFloor-hcFloor) < 0 && elevDir == elevators.DirectionUp {
				goingTowards = true
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

// Go has no built in absolute value function for integers
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
