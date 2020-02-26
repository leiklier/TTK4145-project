package store

import (
	"fmt"
	"sync"
	"time"

	"../elevio"
)

const numElevators = 3 // OBS OBS gAllValues må håndteres ved skalering

type ClearVariant int

// Clear Variants, trenger det litt senere
const (
	CV_All    ClearVariant = 0
	CV_InDirn              = 1
)

type HallCall int

const (
	HC_none HallCall = -1 // Nothing to do
	HC_up            = 0
	HC_down          = 1
	HC_both          = 2
)

type Direction int

const (
	DIR_up   Direction = 1
	DIR_down           = -1
	DIR_idle           = 0
	DIR_both           = 2 // Needed for conversion with HC_both
)

type ElevatorState struct {
	ip            string
	current_floor int
	direction     Direction           // 1=up, 0=idle -1=down
	hall_calls    [numFloors]HallCall // 0=up, -1=idle 1=down. Index is floor
	cab_calls     [numFloors]bool     // index is floor
	door_open     bool
	isWorking     bool
}

type Command struct {
	CurFloor int
	DstFloor int
}

const numFloors = 4

const _pollRate = 20 * time.Millisecond

var gAllElevatorStates = make([]ElevatorState, numElevators) // TODO fiks dynamisk shit

var localElevator = initElevatorState()
var mutex = &sync.Mutex{}

var RecieveElevState = make(chan ElevatorState)
var SendElevState = make(chan ElevatorState)

func getOtherElevatorStates() {
	for {
		change := <-RecieveElevState
		updateElevatorStates(change)
	}
}
func sendOwnState() {
	for {
		SendElevState <- localElevator
	}
}

// Updates the list of all elevators
func updateElevatorStates(elev ElevatorState) {

	for i := 0; i < len(gAllElevatorStates); i++ {
		if gAllElevatorStates[i].ip == elev.ip {
			gAllElevatorStates[i] = elev
		}
	}
}

// Initiates elevator state, default at 1st (0) floor
// Event_handler makes sure that the elevator indeed is on the 1st floor on initiation.
func initElevatorState() ElevatorState {
	elev := ElevatorState{}
	elev.ip = "" // Kanskje legg inn at den henter egen IP
	elev.current_floor = 0
	elev.direction = 0
	elev.hall_calls = [numFloors]HallCall{HC_none, HC_none, HC_none, HC_none}
	elev.cab_calls = [numFloors]bool{false, false, false, false}
	elev.door_open = false
	elev.isWorking = true // Assuming that it works... Mby add some test is possible?
	return elev
}

func UpdateFloorState(floor int) {
	if floor <= numFloors || floor >= 0 {
		mutex.Lock()
		defer mutex.Unlock()
		localElevator.current_floor = floor
		localElevator.cab_calls[floor] = false
		localElevator.hall_calls[floor] = HC_none
	}
}

func UpdateDirectionState(direction Direction) {
	if direction <= DIR_up || direction >= DIR_down {
		mutex.Lock()
		defer mutex.Unlock()
		localElevator.direction = direction
	}
}

func UpdateCalls(floor int, btn_type elevio.ButtonType) bool {
	if floor == localElevator.current_floor && localElevator.direction == 0 {
		fmt.Println("Same floor")
		return false // If elevator is standing still and at floor, dont accept
	}
	if btn_type == elevio.BT_Cab {
		if floor <= numFloors || floor >= 0 { // For cab calls
			mutex.Lock()
			defer mutex.Unlock()
			if localElevator.cab_calls[floor] {
				localElevator.cab_calls[floor] = false
			} else {
				localElevator.cab_calls[floor] = true
			}
		}
	} else {

		current_call := localElevator.hall_calls[floor] // For hall calls

		if current_call == HC_down && btn_type == HC_up {
			localElevator.hall_calls[floor] = HC_both
		} else if current_call == HC_up && btn_type == HC_down {
			localElevator.hall_calls[floor] = HC_both
		} else {
			localElevator.hall_calls[floor] = HallCall(btn_type)
		}
	}
	return true
}

func OpenDoor(door_state bool) {
	localElevator.door_open = door_state

}

func GetFloorAndDir() (int, Direction) {
	mutex.Lock()
	defer mutex.Unlock()
	return localElevator.current_floor, localElevator.direction
}

// This turned out to be a very stupid elevator :/
func GetDestination(dst chan<- Command) { // make cab calls have priority over hall calls
	var up int
	var down int
	var dest int
	var assigned bool

	for {
		time.Sleep(_pollRate)

		// mutex.Lock()  // Got error with the mutex things
		if localElevator.direction == DIR_idle && localElevator.door_open == false {

			for i := 0; i < numFloors; i++ { // Using BFSish to find first floor to go to
				up = localElevator.current_floor + i // Keep going in that direction
				down = localElevator.current_floor - i
				if up <= numFloors-1 {
					if localElevator.cab_calls[up] {
						dest = up
						assigned = true
					}
				}
				if down >= 0 && (!assigned) {
					if localElevator.cab_calls[down] {
						dest = down
						assigned = true
					}
				}

				if !assigned {
					for i := 0; i < numFloors; i++ {
						if localElevator.hall_calls[i] != HC_none && localElevator.current_floor != i {
							dest = i
							assigned = true
							break
						}
					}
				}
				// mutex.Unlock()
			}
			if assigned {
				assigned = false
				dst <- Command{localElevator.current_floor, dest}

			}
		}
	}
}

// GO has no built in absolute value function for integers, so must create my own
// Abs returns the absolute value of x.
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Convert from HC_dir to Elevator Direction
func HCDirToElevDir(hc HallCall) Direction {
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

// Returns IP address of most suited elevator to handle hallcal
func mostSuitedElevator(hc HallCall, originFloor int) string {
	// Steg 1: Gi ordren til heis uten calls, som er nærmest
	// Håndterer om det er idle, og gir til idle

	isClear := true
	var candidates []ElevatorState
	for _, elev := range gAllElevatorStates {
		hc := elev.hall_calls

		for _, v := range hc {
			if v != (HC_none) {
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
		// IP of closest elevator to origin floor. Default with err msg
		currCand := "Something went wrong"
		for _, elev := range candidates {
			floorDiff := Abs(elev.current_floor - originFloor)
			if floorDiff < currMaxDiff {
				currMaxDiff = floorDiff
				currCand = elev.ip
			}
		}
		return currCand
	} else {

		// Steg 2
		currMaxFS := 0
		var currentMax string // IP of elevator with highest FSvalue

		for _, elev := range gAllElevatorStates {
			// Extract elevator information
			currFloor := elev.current_floor
			elevDir := elev.direction

			hcDir := HCDirToElevDir(hc)

		

			var sameDir bool
			if hcDir == elevDir {
				sameDir = true
			} else {
				sameDir = false
			}
			floorDiff := Abs(currFloor - originFloor)

			var goingTowards bool
			if (currFloor-originFloor) > 0 && elevDir == DIR_down {
				goingTowards = true
			} else if (currFloor-originFloor) < 0 && elevDir == DIR_down {
				goingTowards = false
			} else if (currFloor-originFloor) > 0 && elevDir == DIR_up {
				goingTowards = false
			} else if (currFloor-originFloor) < 0 && elevDir == DIR_up {
				goingTowards = true
			} else if (currFloor - originFloor) == 0 {
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
			if FS > currMaxFS {
				currMaxFS = FS
				currentMax = elev.ip
			}
		}
		return currentMax
	}
}
