package store

import (
	"fmt"
	"sync"
	"time"

	"../elevio"
)

const numElevators = 3 // OBS OBS gAllValues må håndteres ved skalering
var GOLANGSUGER int = 3

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
	Ip            string
	Current_floor int
	GDirection    Direction           // 1=up, 0=idle -1=down
	Hall_calls    [numFloors]HallCall // 0=up, -1=idle 1=down. Index is floor
	Cab_calls     [numFloors]bool     // index is floor
	Door_open     bool
	IsWorking     bool
}

type Command struct {
	CurFloor int
	DstFloor int
}

const numFloors = 4

const _pollRate = 20 * time.Millisecond

var GAllElevatorStates = make([]ElevatorState, numElevators) // TODO fiks dynamisk shit

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

	for i := 0; i < len(GAllElevatorStates); i++ {
		if GAllElevatorStates[i].Ip == elev.Ip {
			GAllElevatorStates[i] = elev
		}
	}
}

// Initiates elevator state, default at 1st (0) floor
// Event_handler makes sure that the elevator indeed is on the 1st floor on initiation.
func initElevatorState() ElevatorState {
	elev := ElevatorState{}
	elev.Ip = "" // Kanskje legg inn at den henter egen Ip
	elev.Current_floor = 0
	elev.GDirection = 0
	elev.Hall_calls = [numFloors]HallCall{HC_none, HC_none, HC_none, HC_none}
	elev.Cab_calls = [numFloors]bool{false, false, false, false}
	elev.Door_open = false
	elev.IsWorking = true // Assuming that it works... Mby add some test is possible?
	return elev
}

func UpdateFloorState(floor int) {
	if floor <= numFloors || floor >= 0 {
		mutex.Lock()
		defer mutex.Unlock()
		localElevator.Current_floor = floor
		localElevator.Cab_calls[floor] = false
		localElevator.Hall_calls[floor] = HC_none
	}
}

func UpdateDirectionState(direction Direction) {
	if direction <= DIR_up || direction >= DIR_down {
		mutex.Lock()
		defer mutex.Unlock()
		localElevator.GDirection = direction
	}
}

func UpdateCalls(floor int, btn_type elevio.ButtonType) bool {
	if floor == localElevator.Current_floor && localElevator.GDirection == 0 {
		fmt.Println("Same floor")
		return false // If elevator is standing still and at floor, dont accept
	}
	if btn_type == elevio.BT_Cab {
		if floor <= numFloors || floor >= 0 { // For cab calls
			mutex.Lock()
			defer mutex.Unlock()
			if localElevator.Cab_calls[floor] {
				localElevator.Cab_calls[floor] = false
			} else {
				localElevator.Cab_calls[floor] = true
			}
		}
	} else {

		current_call := localElevator.Hall_calls[floor] // For hall calls

		if current_call == HC_down && btn_type == HC_up {
			localElevator.Hall_calls[floor] = HC_both
		} else if current_call == HC_up && btn_type == HC_down {
			localElevator.Hall_calls[floor] = HC_both
		} else {
			localElevator.Hall_calls[floor] = HallCall(btn_type)
		}
	}
	return true
}

func OpenDoor(door_state bool) {
	localElevator.Door_open = door_state

}

func GetFloorAndDir() (int, Direction) {
	mutex.Lock()
	defer mutex.Unlock()
	return localElevator.Current_floor, localElevator.GDirection
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
		if localElevator.GDirection == DIR_idle && localElevator.Door_open == false {

			for i := 0; i < numFloors; i++ { // Using BFSish to find first floor to go to
				up = localElevator.Current_floor + i // Keep going in that Direction
				down = localElevator.Current_floor - i
				if up <= numFloors-1 {
					if localElevator.Cab_calls[up] {
						dest = up
						assigned = true
					}
				}
				if down >= 0 && (!assigned) {
					if localElevator.Cab_calls[down] {
						dest = down
						assigned = true
					}
				}

				if !assigned {
					for i := 0; i < numFloors; i++ {
						if localElevator.Hall_calls[i] != HC_none && localElevator.Current_floor != i {
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
				dst <- Command{localElevator.Current_floor, dest}

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

// Returns Ip address of most suited elevator to handle hallcal
func MostSuitedElevator(hc HallCall, originFloor int) string {
	// Steg 1: Gi ordren til heis uten calls, som er nærmest
	// Håndterer om det er idle, og gir til idle

	isClear := true
	var candidates []ElevatorState
	for _, elev := range GAllElevatorStates {
		hc := elev.Hall_calls
		dir := elev.GDirection
		if dir == DIR_idle {
			return elev.Ip
		}

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
		// Ip of closest elevator to origin floor. Default with err msg
		currCand := "Something went wrong"
		for _, elev := range candidates {
			floorDiff := Abs(elev.Current_floor - originFloor)
			if floorDiff < currMaxDiff {
				currMaxDiff = floorDiff
				currCand = elev.Ip
				fmt.Println("Kom inn i isClear")
			}
		}
		return currCand
	} else {
		fmt.Println("Kom inn i steg 2")

		// Steg 2
		currMaxFS := 0
		var currentMax string // Ip of elevator with highest FSvalue

		for _, elev := range GAllElevatorStates {
			// Extract elevator information
			currFloor := elev.Current_floor
			elevDir := elev.GDirection

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
				currentMax = elev.Ip
			}
		}
		return currentMax
	}
}
