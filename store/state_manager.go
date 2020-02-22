package store

import (
	"fmt"
	"sync"
	"time"

	"../elevio"
)

type ClearMethod int

const (
	CV_All    ClearMethod = 0
	CV_InDirn             = 1
)

type HallCall int

const (
	HC_none HallCall = -1
	HC_up            = 0
	HC_down          = 1
	HC_both          = 2
)

type Direction int

const (
	DIR_up   Direction = 1
	DIR_down           = -1
	DIR_idle           = 0
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

var gAllElevatorStates []ElevatorState

var localElevator = initElevatorState()
var mutex = &sync.Mutex{}

var recieveElevState = make(chan ElevatorState)
var sendElevState = make(chan ElevatorState)

func getOtherElevatorStates() {
	for {
		change := <-recieveElevState
		updateElevatorStates(change)
	}
}
func sendOwnState() {
	for {
		sendElevState <- localElevator
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

// Returns IP address of most suited elevator to handle hallcal
func mostSuitedElevator(hallCall HallCall) string {
	// Steg 1: Gi ordren til heis uten calls, som er nærmest

	//	type ElevatorState struct {
	/* 	ip            string
	   	current_floor int
	   	direction     Direction           // 1=up, 0=idle -1=down
	   	hall_calls    [numFloors]HallCall // 0=up, -1=idle 1=down. Index is floor
	   	cab_calls     [numFloors]bool     // index is floor
	   	door_open     bool
	   }*/
}
