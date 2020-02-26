package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"strconv"

	"../store"
) 
const numFloors = 4

var hc1 = [numFloors]store.HallCall{store.HC_none, store.HC_up, store.HC_none, store.HC_none}
var hc2 = [numFloors]store.HallCall{store.HC_up, store.HC_up, store.HC_down, store.HC_none}
var hc3 = [numFloors]store.HallCall{store.HC_up, store.HC_down, store.HC_none, store.HC_down}

var cc = [numFloors]bool{false, false, false, false}

func main() {
	f := getLines("testscenario.txt")
	testAmount := (len(f)-1)/5
	fmt.Println("testamount is",testAmount,"\n")
	for i:=0; i<testAmount;i++{
		doTest(f,i)
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
		fileTextLines = append(fileTextLines, string(fileScanner.Text()))
	}

	readFile.Close()

	return fileTextLines
}
// En test tar 5 linjer 
// TestNr er nullindeksert
func doTest(fileTextLines []string, testNr int) {
	var elev1 store.ElevatorState
	var elev2 store.ElevatorState
	var elev3 store.ElevatorState
	var HCFloor int
	var hc_dir store.HallCall
	
	j := 0
	for i, line := range fileTextLines {
		if (i) == 0 && testNr < 1{ 
			continue
		}
		if i < ((testNr*5) +1) {
			continue
		}
		splitLine := strings.Split(line,",")
	
		if splitLine[0] == "-" {
			fmt.Println("Results from scenario",testNr,"are: \n")
			fmt.Println("Elevator 1 is at floor", elev1.Current_floor, "going", dirToText(elev1.GDirection))
			fmt.Println("Elevator 2 is at floor", elev2.Current_floor, "going", dirToText(elev2.GDirection))
			fmt.Println("Elevator 3 is at floor", elev3.Current_floor, "going", dirToText(elev3.GDirection))
			fmt.Println()
			fmt.Println("Hall call is at floor", HCFloor, "going", dirToText(store.HCDirToElevDir(hc_dir)),"\n")
			iperino := store.MostSuitedElevator(hc_dir, HCFloor)
			fmt.Println("Most suited elevator is elevator nr:", iperino)
			fmt.Println("------------------------------------------------")
			break
		}

		if splitLine[0] == "HC"{

			HCFloor,_ = strconv.Atoi(splitLine[1])
			hc_dir = stringToHC(splitLine[2])
		}

		floor := splitLine[1]
		dir := stringToDir(string(splitLine[2]))
		hc_list := hc1
		cc_list := cc
		door,_ := strconv.ParseBool(splitLine[3])
		// Assign to correct elevator 
		if j == 0 {
			elev1.Ip = "1"
			elev1.Current_floor,_ = strconv.Atoi(floor)
			elev1.GDirection = dir
			elev1.Hall_calls = hc_list
			elev1.Cab_calls = cc_list
			elev1.Door_open = door
			elev1.IsWorking = true
			store.GAllElevatorStates[0] = elev1
		}
		if j == 1 {
			elev2.Ip = "2"
			elev2.Current_floor,_ = strconv.Atoi(floor)
			elev2.GDirection = dir
			elev2.Hall_calls = hc_list
			elev2.Cab_calls = cc_list
			elev2.Door_open = door
			elev2.IsWorking = true
			store.GAllElevatorStates[1] = elev2

		}
		if j == 2 {
			elev3.Ip = "3"
			elev3.Current_floor,_ = strconv.Atoi(floor)
			elev3.GDirection = dir
			elev3.Hall_calls = hc_list
			elev3.Cab_calls = cc_list
			elev3.Door_open = door
			elev3.IsWorking = true
			store.GAllElevatorStates[2] = elev3
		}
		j = j+1
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
	default:
		return store.DIR_both
	}
}
// Can only take U,D,I or B
func stringToHC(s string) store.HallCall {
	switch s {
	case "U":
		return store.HC_up
	case "D":
		return store.HC_down
	case "I":
		return store.HC_none
	default:
		return store.HC_both
	}
}


