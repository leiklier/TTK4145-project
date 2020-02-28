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

type HallCall struct {
	GDirection Direction
	GFloor     int
}

type HallCallDir int

const (
	HC_none HallCallDir = -1 // Nothing to do
	HC_up               = 0
	HC_down             = 1
	HC_both             = 2
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
	elev.Hall_calls = [numFloors]HallCall{{DIR_idle, 0}, {DIR_idle, 0}, {DIR_idle, 0}, {DIR_idle, 0}}
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
