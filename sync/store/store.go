package store

import (
	"errors"
	"sync"

	"../../network/peers"
	"../costfunction"
	"../elevators"
)

var gState []elevators.Elevator_s
var gStateMutex sync.Mutex

const NumFloors = 4

var StoreUpdate = make(chan bool)

func Init() {
	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	gState = make([]elevators.Elevator_s, 1)
	localIP := peers.GetRelativeTo(peers.Self, 0)

	selfInitialFloor := 0
	gState[0] = elevators.New(localIP, NumFloors, selfInitialFloor)
}

func Add(newElevator elevators.Elevator_s) error {
	// gStateMutex.Lock()
	// defer gStateMutex.Unlock()

	for _, elevatorInStore := range gState {
		if newElevator.GetIP() == elevatorInStore.GetIP() {
			return errors.New("ERR_ALREADY_EXISTS")
		}
	}

	gState = append(gState, newElevator)
	return nil
}

func Remove(ipToRemove string) {
	// gStateMutex.Lock()
	// defer gStateMutex.Unlock()

	for i, currentElevator := range gState {
		if currentElevator.GetIP() == ipToRemove {
			copy(gState[i:], gState[i+1:])                 // Shift peers[i+1:] left one index.
			gState[len(gState)-1] = elevators.Elevator_s{} // Erase last element (write nil value).
			gState = gState[:len(gState)-1]                // Truncate slice.
		}
	}
}

func GetAll() []elevators.Elevator_s {
	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	return gState
}

func GetElevator(elevatorIP string) (elevators.Elevator_s, error) {
	gStateMutex.Lock()
	defer gStateMutex.Unlock()
	for _, elevatorInStore := range gState {
		if elevatorInStore.GetIP() == elevatorIP {
			return elevatorInStore, nil
		}
	}
	return elevators.Elevator_s{}, errors.New("ERR_ELEVATOR_DOES_NOT_EXIST")
}

func GetCurrentFloor(elevatorIP string) (int, error) {
	elevator, err := GetElevator(elevatorIP)
	if err != nil {
		return 0, err
	}

	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	return elevator.GetCurrentFloor(), nil
}

func SetCurrentFloor(elevatorIP string, currentFloor int) error {
	elevator, err := GetElevator(elevatorIP)
	if err != nil {
		return err
	}

	gStateMutex.Lock()
	defer gStateMutex.Unlock()
	elevator.SetCurrentFloor(currentFloor)
	UpdateState(elevator)
	return nil
}

func GetDirectionMoving(elevatorIP string) (elevators.Direction_e, error) {
	elevator, err := GetElevator(elevatorIP)
	if err != nil {
		return 0, err
	}

	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	return elevator.GetDirectionMoving(), nil
}

func GetPreviousFloor(elevatorIP string) (int, error) {
	elevator, err := GetElevator(elevatorIP)
	if err != nil {
		return 0, err
	}

	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	return elevator.GetPreviousFloor(), nil
}

func SetDirectionMoving(elevatorIP string, newDirection elevators.Direction_e) error {
	elevator, err := GetElevator(elevatorIP)
	if err != nil {
		return err
	}

	gStateMutex.Lock()
	defer gStateMutex.Unlock()
	elevator.SetDirectionMoving(newDirection)
	UpdateState(elevator)
	return nil
}

func AddHallCall(elevatorIP string, hallCall elevators.HallCall_s) error {
	elevator, err := GetElevator(elevatorIP)
	if err != nil {
		return err
	}

	gStateMutex.Lock()
	defer gStateMutex.Unlock()
	elevator.AddHallCall(hallCall)
	UpdateState(elevator)
	return nil
}

func RemoveHallCalls(elevatorIP string, floor int) error {
	elevator, err := GetElevator(elevatorIP)
	if err != nil {
		return err
	}

	gStateMutex.Lock()
	defer gStateMutex.Unlock()
	elevator.RemoveHallCalls(floor)
	UpdateState(elevator)
	return nil
}

func GetAllHallCalls(elevatorIP string) ([]elevators.HallCall_s, error) {
	elevator, err := GetElevator(elevatorIP)
	if err != nil {
		return []elevators.HallCall_s{}, err
	}

	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	return elevator.GetAllHallCalls(), nil
}

func AddCabCall(elevatorIP string, floor int) error {
	elevator, err := GetElevator(elevatorIP)
	if err != nil {
		return err
	}

	gStateMutex.Lock()
	defer gStateMutex.Unlock()
	elevator.AddCabCall(floor)
	UpdateState(elevator)
	return nil
}

func RemoveCabCall(elevatorIP string, floor int) error {
	elevator, err := GetElevator(elevatorIP)
	if err != nil {
		return err
	}

	gStateMutex.Lock()
	defer gStateMutex.Unlock()
	elevator.RemoveCabCall(floor)
	UpdateState(elevator)
	return nil

}

func GetAllCabCalls(elevatorIP string) ([]bool, error) {
	elevator, err := GetElevator(elevatorIP)
	if err != nil {
		return []bool{}, err
	}

	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	return elevator.GetAllCabCalls(), nil
}

func MostSuitedElevator(hcFloor int, hcDirection elevators.Direction_e) string {
	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	return costfunction.MostSuitedElevator(gState, NumFloors, hcFloor, hcDirection)
}

func UpdateState(elevator elevators.Elevator_s) {
	Remove(elevator.GetIP())
	Add(elevator)
	// StoreUpdate <- true
}
