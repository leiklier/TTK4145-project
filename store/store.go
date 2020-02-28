package store

import(
	"../network/peers"
	"./elevator"
	"sync"
	"errors"
)

type controlSignal_s struct {
	Command controlCommand_t,
	Payload []
}

var gState = []elevator.Elevator_s
var gStateMutex sync.Mutex

var gIsInitialized = false
const gNumFloors = 4

func initialize() {
	if gIsInitialized {
		return
	}
	func stateServer() {
		state := make([]Elevator_s, 1)
	
		localIP := peers.GetRelativeTo(peers.Self, 0)
		selfInitialFloor := 0
		state := elevator.New(localIP, gNumFloors, selfInitialFloor)
	
		for {
	
		}
	
	}
	
	defer gStateMutex.Unlock()
	
	gState = make([]elevator.Elevator_s, 1)

	localIP := peers.GetRelativeTo(peers.Self, 0)
	selfInitialFloor := 0
	gState[0] = elevator.New(localIP, gNumFloors, selfInitialFloor)


}

func Add(elevator.Elevator_s newElevator) error {
	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	for _, elevatorInStore := range gState {
		if newElevator.GetIP() == elevatorInStore.GetIP() {
			return errors.New("ERR_ALREADY_EXISTS")
		}
	}

	gState = gState
}