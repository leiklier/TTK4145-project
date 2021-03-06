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

	"../network/peers"
	"./elevators"
	"./costfunction"
)

var gState []elevators.Elevator_s
var gStateMutex sync.Mutex

const NumFloors = 4

var BACKUPNAME string // Blir dette et problem når vi tester på 1 pc?

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

	var ccString string
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
	file.WriteString(ccString + ";" + date + " " + time)
	file.Close()
}

func ReadCCBackup(filename string) []bool {
	// Check if the file exists
	_, err := os.Stat(filename)

	if err != nil {
		fmt.Println("File does not exist, initializing", filename)
		InitCCBackupFile(filename)
	}

	// File now 100% exists and we can proceed
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

	onlyCC := strings.Split(s, ";")[0]
	fmt.Println(onlyCC)
	ccs := strings.Split(onlyCC, ",")
	var newcc []bool
	for _, i := range ccs {
		if i == "true" {
			newcc = append(newcc, true)
		} else {
			newcc = append(newcc, false)
		}
	}
	return newcc
}

// Initializes with no CC
func InitCCBackupFile(filename string) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Failed to create backup file!")
		log.Fatal(err)
	}
	file.Close()
	var cc = []bool{true, false, false, false}
	WriteCCBackup(cc, filename)
}

func Init() {
	gStateMutex.Lock()

	gState = make([]elevators.Elevator_s, 1)
	localHostname := peers.GetRelativeTo(peers.Self, 0)
	BACKUPNAME = "backup_" + localHostname + ".txt"

	selfInitialFloor := 0
	gState[0] = elevators.New(localHostname, NumFloors, selfInitialFloor)
	gStateMutex.Unlock()
	// Load CC From file
	ccFromFile := ReadCCBackup(BACKUPNAME)

	fmt.Println("From the backup file I've got:", ccFromFile)
	for i, e := range ccFromFile {
		if e {
			AddCabCall(localHostname, i)
		}
	}

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
	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	for _, elevatorInStore := range gState {
		if newElevator.GetHostname() == elevatorInStore.GetHostname() {
			return errors.New("ERR_ALREADY_EXISTS")
		}
	}

	gState = append(gState, newElevator)
	return nil
}

func Remove(HostnameToRemove string) {
	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	for i, currentElevator := range gState {
		if currentElevator.GetHostname() == HostnameToRemove {
			copy(gState[i:], gState[i+1:])                 // Shift gState[i+1:] left one index.
			gState[len(gState)-1] = elevators.Elevator_s{} // Erase last element (write nil value).
			gState = gState[:len(gState)-1]                // Truncate slice.
			break
		}
	}
}

func Replace(elevator elevators.Elevator_s) {
	gStateMutex.Lock()
	defer gStateMutex.Unlock()

	// Remove it...
	for i, currentElevator := range gState {
		if currentElevator.GetHostname() == elevator.GetHostname() {
			copy(gState[i:], gState[i+1:])                 // Shift gState[i+1:] left one index.
			gState[len(gState)-1] = elevators.Elevator_s{} // Erase last element (write nil value).
			gState = gState[:len(gState)-1]                // Truncate slice.
			break
		}
	}

	// ...and then add it again
	gState = append(gState, elevator)

	select {
	case ShouldRecalculateNextFloorChannel <- true: // Only add to channel if not full
	default:
	}

	select {
	case ShouldRecalculateHCLightsChannel <- true:
	default:
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

	elevator.SetCurrentFloor(currentFloor)
	Replace(elevator)
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

func SetDirectionMoving(elevatorHostname string, newDirection elevators.Direction_e) error {
	elevator, err := GetElevator(elevatorHostname)
	if err != nil {
		return err
	}
	elevator.SetDirectionMoving(newDirection)
	Replace(elevator)
	return nil
}

func AddHallCall(elevatorHostname string, hallCall elevators.HallCall_s) error {
	elevator, err := GetElevator(elevatorHostname)
	if err != nil {
		return err
	}

	elevator.AddHallCall(hallCall)
	Replace(elevator)
	return nil
}

func RemoveHallCalls(elevatorHostname string, floor int) error {
	elevator, err := GetElevator(elevatorHostname)
	if err != nil {
		return err
	}

	elevator.RemoveHallCalls(floor)
	Replace(elevator)
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

	elevator.AddCabCall(floor)
	Replace(elevator)

	cc, _ := GetAllCabCalls(peers.GetRelativeTo(peers.Self, 0))
	fmt.Println("I've updated, current cc state:,", cc)
	WriteCCBackup(cc, BACKUPNAME)

	return nil
}

func RemoveCabCall(elevatorHostname string, floor int) error {
	elevator, err := GetElevator(elevatorHostname)
	if err != nil {
		return err
	}

	elevator.RemoveCabCall(floor)
	Replace(elevator)

	cc, _ := GetAllCabCalls(peers.GetRelativeTo(peers.Self, 0))
	WriteCCBackup(cc, BACKUPNAME)

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
