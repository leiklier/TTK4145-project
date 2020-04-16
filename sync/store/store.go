package store

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"../../network/peers"
	"../costfunction"
	"../elevators"
)

var gState []elevators.Elevator_s
var gStateMutex sync.Mutex

const NumFloors = 4
const BACKUPNAME = "backup.txt" //Blir dette et problem når vi tester på 1 pc?

var ShouldRecalculateNextFloorChannel = make(chan bool, 100)
var ShouldRecalculateHCLightsChannel = make(chan bool, 100)

func WriteCCBackup(cabCalls []bool, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}

	t := time.Now()
	tstring := t.String()
	date := strings.Split(tstring, " ")[0]
	time := strings.Split(tstring, " ")[1]

	ccString := "("
	for i, v := range cabCalls {
		if i != (len(cabCalls) - 1) {
			if v {
				ccString = ccString + "true,"
			} else {
				ccString = ccString + "false,"
			}
		} else {
			if v {
				ccString = ccString + "true"
			} else {
				ccString = ccString + "false"
			}
		}

	}
	ccString = ccString + ")"
	file.WriteString(ccString + ";" + date + " " + time)
	file.Close()
}

func ReadCCBackup(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	stream, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	s := string(stream)
	file.Close()

	return s
}

func Init() {
	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	gState = make([]elevators.Elevator_s, 1)
	localHostname := peers.GetRelativeTo(peers.Self, 0)

	selfInitialFloor := 0
	gState[0] = elevators.New(localHostname, NumFloors, selfInitialFloor)
}

func IsExistingHallCall(hallCall elevators.HallCall_s) bool {
	allElevators := GetAll()
	for _, elevator := range allElevators {
		currHallCall := elevator.GetAllHallCalls()[hallCall.Floor]
		if hallCall.Direction == elevators.DirectionUp &&
			(currHallCall.Direction == elevators.DirectionUp || currHallCall.Direction == elevators.DirectionBoth) {
			return true
		} else if hallCall.Direction == elevators.DirectionDown &&
			(currHallCall.Direction == elevators.DirectionDown || currHallCall.Direction == elevators.DirectionBoth) {
			return true
		} else if hallCall.Direction == elevators.DirectionBoth &&
			currHallCall.Direction != elevators.DirectionIdle {
			return true
		}

	}

	return false
}

func Add(newElevator elevators.Elevator_s) error {
	// gStateMutex.Lock()
	// defer gStateMutex.Unlock()

	for _, elevatorInStore := range gState {
		if newElevator.GetHostname() == elevatorInStore.GetHostname() {
			return errors.New("ERR_ALREADY_EXISTS")
		}
	}

	gState = append(gState, newElevator)
	return nil
}

func Remove(HostnameToRemove string) {
	// gStateMutex.Lock()
	// defer gStateMutex.Unlock()

	for i, currentElevator := range gState {
		if currentElevator.GetHostname() == HostnameToRemove {
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

func GetElevator(elevatorHostname string) (elevators.Elevator_s, error) {
	gStateMutex.Lock()
	defer gStateMutex.Unlock()
	for _, elevatorInStore := range gState {
		if elevatorInStore.GetHostname() == elevatorHostname {
			return elevatorInStore, nil
		}
	}
	return elevators.Elevator_s{}, errors.New("ERR_ELEVATOR_DOES_NOT_EXIST")
}

func GetCurrentFloor(elevatorHostname string) (int, error) {
	elevator, err := GetElevator(elevatorHostname)
	if err != nil {
		return 0, err
	}

	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	return elevator.GetCurrentFloor(), nil
}

func SetCurrentFloor(elevatorHostname string, currentFloor int) error {
	elevator, err := GetElevator(elevatorHostname)
	if err != nil {
		return err
	}

	gStateMutex.Lock()
	defer gStateMutex.Unlock()
	elevator.SetCurrentFloor(currentFloor)
	UpdateState(elevator)
	return nil
}

func GetDirectionMoving(elevatorHostname string) (elevators.Direction_e, error) {
	elevator, err := GetElevator(elevatorHostname)
	if err != nil {
		return 0, err
	}

	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	return elevator.GetDirectionMoving(), nil
}

func GetPreviousFloor(elevatorHostname string) (int, error) {
	elevator, err := GetElevator(elevatorHostname)
	if err != nil {
		return 0, err
	}

	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	return elevator.GetPreviousFloor(), nil
}

func SetDirectionMoving(elevatorHostname string, newDirection elevators.Direction_e) error {
	elevator, err := GetElevator(elevatorHostname)
	if err != nil {
		return err
	}

	gStateMutex.Lock()
	defer gStateMutex.Unlock()
	elevator.SetDirectionMoving(newDirection)
	UpdateState(elevator)
	return nil
}

func AddHallCall(elevatorHostname string, hallCall elevators.HallCall_s) error {
	elevator, err := GetElevator(elevatorHostname)
	if err != nil {
		return err
	}

	gStateMutex.Lock()
	defer gStateMutex.Unlock()
	elevator.AddHallCall(hallCall)
	UpdateState(elevator)
	return nil
}

func RemoveHallCalls(elevatorHostname string, floor int) error {
	elevator, err := GetElevator(elevatorHostname)
	if err != nil {
		return err
	}

	gStateMutex.Lock()
	defer gStateMutex.Unlock()
	elevator.RemoveHallCalls(floor)
	UpdateState(elevator)
	return nil
}

func GetAllHallCalls(elevatorHostname string) ([]elevators.HallCall_s, error) {
	elevator, err := GetElevator(elevatorHostname)
	if err != nil {
		return []elevators.HallCall_s{}, err
	}

	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	return elevator.GetAllHallCalls(), nil
}

func AddCabCall(elevatorHostname string, floor int) error {
	elevator, err := GetElevator(elevatorHostname)
	if err != nil {
		return err
	}

	gStateMutex.Lock()
	defer gStateMutex.Unlock()
	elevator.AddCabCall(floor)
	UpdateState(elevator)
	return nil
}

func RemoveCabCall(elevatorHostname string, floor int) error {
	elevator, err := GetElevator(elevatorHostname)
	if err != nil {
		return err
	}

	gStateMutex.Lock()
	defer gStateMutex.Unlock()
	elevator.RemoveCabCall(floor)
	UpdateState(elevator)
	return nil

}

func GetAllCabCalls(elevatorHostname string) ([]bool, error) {
	elevator, err := GetElevator(elevatorHostname)
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
	Remove(elevator.GetHostname())
	Add(elevator)

	select {
	case ShouldRecalculateNextFloorChannel <- true: // Only add to channel if not full
	default:
	}

	select {
	case ShouldRecalculateHCLightsChannel <- true:
	default:
	}
}

func PrintStateAll() {
	for {
		select {
		case <-time.After(5 * time.Second):
			all := GetAll()
			for index, item := range all {
				fmt.Println("---------------")
				fmt.Printf("index : %d\n", index)
				fmt.Printf("%+v\n", item.GetHostname())
				fmt.Printf("Current floor: %+v\n", item.GetCurrentFloor())
				fmt.Printf("Cab calls: %+v\n", item.GetAllCabCalls())
				fmt.Printf("Hall calls: %+v\n", item.GetAllHallCalls())
				fmt.Printf("Direction: %+v\n", item.GetDirectionMoving())
			}

		}
	}
}
