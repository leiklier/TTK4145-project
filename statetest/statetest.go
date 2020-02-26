package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"../store"
) /*

type ElevatorState struct {
	ip            string
	current_floor int
	direction     Direction           // 1=up, 0=idle -1=down
	hall_calls    [numFloors]HallCall // 0=up, -1=idle 1=down. Index is floor
	cab_calls     [numFloors]bool     // index is floor
	door_open     bool
	isWorking     bool
}*/
const numFloors = 4

var hc1 = [numFloors]store.HallCall{store.HC_none, store.HC_up, store.HC_none, store.HC_none}
var hc2 = [numFloors]store.HallCall{store.HC_up, store.HC_up, store.HC_down, store.HC_none}
var hc3 = [numFloors]store.HallCall{store.HC_up, store.HC_down, store.HC_none, store.HC_down}

var cc = [numFloors]bool{false, false, false, false}

var elev1 = store.ElevatorState{"1", 0, store.DIR_idle, hc1, cc, false, true}
var elev2 = store.ElevatorState{"2", 1, store.DIR_idle, hc2, cc, false, true}
var elev3 = store.ElevatorState{"3", 2, store.DIR_idle, hc3, cc, false, true}

func main() {
	store.GAllElevatorStates[0] = elev1
	store.GAllElevatorStates[1] = elev2
	store.GAllElevatorStates[2] = elev3

	fmt.Println("Elevator 1 is at floor", elev1.Current_floor, "going", dirToText(elev1.GDirection))
	fmt.Println("Elevator 2 is at floor", elev2.Current_floor, "going", dirToText(elev2.GDirection))
	fmt.Println("Elevator 3 is at floor", elev3.Current_floor, "going", dirToText(elev3.GDirection))
	fmt.Println()

	var HC store.HallCall = store.HC_down
	HCFloor := 2

	fmt.Println("Hall call is at floor", HCFloor, "going", dirToText(store.HCDirToElevDir(HC)))
	iperino := store.MostSuitedElevator(HC, HCFloor)
	fmt.Println("Most suited elevator is elevator nr:", iperino)

	readFile, err := os.Open("testscenario.txt")

	if err != nil {
		log.Fatalf("failed to open file: %s", err)
	}

}
func dirToText(dir store.Direction) string {
	switch dir {
	case store.DIR_up:
		return "up"
	case store.DIR_idle:
		return "nowhere"
	case store.DIR_down:
		return "down"
	default:
		return "both"
	}
}

func getLines(filename string) []string {
	readFile, err := os.Open(filename)

	if err != nil {
		log.Fatalf("failed to open file: %s", err)
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	var fileTextLines []string

	for fileScanner.Scan() {
		fileTextLines = append(fileTextLines, fileScanner.Text())
	}

	readFile.Close()

	return fileTextLines
}

func doTest(fileTextLines []string) {
	for i, line := range fileTextLines {
		if i == 0 {
			continue
		}
		floor := line[1]
		dir := stringToDir(string(line[2]))
		hc_list := hc1
		cc_list := cc
		door := string(line[3])
	}
}

// Can only take U,D,I or B
func stringToDir(s string) store.Direction {
	switch s {
	case "U":
		return store.DIR_up
	case "D":
		return store.DIR_down
	case "I":
		return store.DIR_idle
	case "B":
		return store.DIR_both
	}
}
