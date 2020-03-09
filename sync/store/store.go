package store

import(
	"../../network/peers"
	"../elevators"
	"../costfunction"
	"sync"
	"errors"
)

var gState = []elevators.Elevator_s
var gStateMutex sync.Mutex

var gIsInitialized = false
const NumFloors = 4

func initialize() {
	if gIsInitialized {
		return
	}

	gStateMutex.Lock()
	defer gStateMutex.Unlock()
	gIsInitialized = true

	gState = make([]elevators.Elevator_s, 1)

	localIP := peers.GetRelativeTo(peers.Self, 0)
	selfInitialFloor := 0
	gState[0] = elevators.New(localIP, NumFloors, selfInitialFloor)


}

func Add(newElevator elevators.Elevator_s) error {
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

func Remove(ipToRemove string) {
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

func getElevator(elevatorIP string) (elevators.Elevator_s, error) {
	for i, elevatorInStore := range gState {
		if elevatorInStore.GetIP() == elevatorIP {
			return elevatorInStore, nil
		}
	}
	return elevators.Elevator_s{}, errors.New("ERR_ELEVATOR_DOES_NOT_EXIST")
}

func GetCurrentFloor(elevatorIP string)  (int, error) {
	initialize()

	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	elevator, err := getElevator(elevatorIP)
	if err != nil {
		return 0, err
	}

	return elevator.GetCurrentFloor(), nil
}

func SetCurrentFloor(elevatorIP string, currentFloor int) error {
	initialize()

	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	elevator, err := getElevator(elevatorIP)
	if err != nil {
		return err
	}

	elevator.SetCurrentFloor(currentFloor)
	return nil
}

func GetDirectionMoving(elevatorIP string)  (elevators.Direction_e, error) {
	initialize()

	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	elevator, err := getElevator(elevatorIP)
	if err != nil {
		return 0, err
	}

	return elevator.GetDirectionMoving(), nil
}

func SetDirectionMoving(elevatorIP string, newDirection elevators.Direction_e) error {
	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	elevator, err := getElevator(elevatorIP)
	if err != nil {
		return err
	}

	elevator.SetDirectionMoving(newDirection)
	return nil
}

func AddHallCall(elevatorIP string, hallCall elevators.HallCall_s) error {
	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	elevator, err := getElevator(elevatorIP)
	if err != nil {
		return err
	}


	elevator.AddHallCall(hallCall)

	return nil
}

func RemoveHallCalls(elevatorIP string, floor int) error {
	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	elevator, err := getElevator(elevatorIP)
	if err != nil {
		return err
	}

	elevator.RemoveHallCalls(floor)

	return nil
}

func AddCabCall(elevatorIP string, floor int) error {
	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	elevator, err := getElevator(elevatorIP)
	if err != nil {
		return err
	}

	elevator.AddCabCall(floor)

	return nil

}

func RemoveCabCall(elevatorIP string, floor int) error {
	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	elevator, err := getElevator(elevatorIP)
	if err != nil {
		return err
	}

	elevator.RemoveCabCall(floor)

}

func MostSuitedElevator(hcFloor int, hcDirection elevators.HCDirection_e) string {
	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	return costfunction.MostSuitedElevator(allElevators, NumFloors, hcFloor, hcDirection)
}