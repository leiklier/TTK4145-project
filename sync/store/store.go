package store

import(
	"../../network/peers"
	"../elevators"
	"sync"
	"errors"
)

var gState = []elevators.Elevator_s
var gStateMutex sync.Mutex

var gIsInitialized = false
const gNumFloors = 4

func initialize() {
	if gIsInitialized {
		return
	}

	gStateMutex.Lock()
	defer gStateMutex.Unlock()
	gIsInitialized = true

	gState = make([]elevator.Elevator_s, 1)

	localIP := peers.GetRelativeTo(peers.Self, 0)
	selfInitialFloor := 0
	gState[0] = elevators.New(localIP, gNumFloors, selfInitialFloor)


}

func Add(elevators.Elevator_s newElevator) error {
	initialize()

	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	for _, elevatorInStore := range gState {
		if newElevator.GetIP() == elevatorInStore.GetIP() {
			return errors.New("ERR_ALREADY_EXISTS")
		}
	}

	gState = append(gState, newElevator)
}

func Remove(ipToRemove string) elevators.Elevator_s {
	initialize()

	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	for i, currentElevator := range gState {
		if currentElevator.GetIP() == ipToRemove {
			copy(gState[i:], gState[i+1:]) // Shift peers[i+1:] left one index.
			gState[len(gState)-1] = nil     // Erase last element (write nil value).
			gState = gState[:len(gState)-1] // Truncate slice.
		}
	}
}

func Get(elevatorIP) elevators.Elevator_s {

}

func SetCurrentFloor(elevatorIP, currentFloor)