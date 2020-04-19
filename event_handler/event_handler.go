package event_handler

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"./next_floor"

	"../elevio"
	"../network/peers"
	"../order_distributor"
	"../store"
	"../store/elevators"
)

var selfHostname string

var drv_buttons = make(chan elevio.ButtonEvent)
var drv_floors = make(chan int)

func Init(elevNumber int) {
	// First we start the server
	fmt.Println("Starting elevator server ...")
	selfHostname = peers.GetRelativeTo(peers.Self, 0)
	connPort := strconv.Itoa(15657 + elevNumber)
	time.Sleep(time.Duration(1 * time.Second)) // To avoid crash due to not started sim

	elevio.Init("localhost:"+connPort, store.NumFloors)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)

	store.SetCurrentFloor(selfHostname, store.NumFloors-1)

	go elevatorDriver()
	go hallCallLightDriver()
	go buttonHandler()

}

// Handles action related to a button getting pressed
func buttonHandler() {
	// Reset all cab call button lamps:
	cabCalls, _ := store.GetAllCabCalls(selfHostname)
	for floor, cabCall := range cabCalls {
		elevio.SetButtonLamp(elevio.BT_Cab, floor, cabCall)
	}

	for {
		buttonEvent := <-drv_buttons
		currentFloor, _ := store.GetCurrentFloor(selfHostname)

		if buttonEvent.Floor == currentFloor && buttonEvent.Button != elevio.BT_Cab {
			openAndCloseDoors(currentFloor)
			continue
		}

		if buttonEvent.Button == elevio.BT_Cab {
			// Handle cab call
			if buttonEvent.Floor != currentFloor {
				elevio.SetButtonLamp(elevio.BT_Cab, buttonEvent.Floor, true)
				store.AddCabCall(selfHostname, buttonEvent.Floor)
			}
		} else {
			// Handle hall call
			elevDir := btnDirToElevDir(buttonEvent.Button)
			hCall := elevators.HallCall_s{Floor: buttonEvent.Floor, Direction: elevDir}
			if store.IsExistingHallCall(hCall) {
				continue
			}
			mostSuitedHostname := store.MostSuitedElevator(buttonEvent.Floor, elevDir)
			fmt.Printf("Most suited ip: %s\n", mostSuitedHostname)

			store.AddHallCall(mostSuitedHostname, hCall)
			order_distributor.SendHallCall(mostSuitedHostname, hCall)
		}
	}
}

// TODO: Leik fiks comment.
func elevatorDriver() {
	goToFloor(0)

	for {
		currentFloor, _ := store.GetCurrentFloor(selfHostname)
		if store.IsExistingHallCall(elevators.HallCall_s{Floor: currentFloor, Direction: elevators.DirectionBoth}) {
			openAndCloseDoors(currentFloor)

			store.RemoveHallCalls(selfHostname, currentFloor)
			order_distributor.ShouldSendStateUpdate <- true
		}

		nextFloor := next_floor.GetNextFloor()
		if nextFloor != next_floor.NoNextFloor {
			goToFloor(nextFloor)
		}
		<-store.ShouldRecalculateNextFloorChannel
	}
}

// Handles lights for hallcall buttons
func hallCallLightDriver() {
	for {
		var lights [store.NumFloors][2]bool
		allElevators := store.GetAll()

		for _, elevator := range allElevators {
			for _, hallCall := range elevator.GetAllHallCalls() {
				if hallCall.Direction == elevators.DirectionUp {
					lights[hallCall.Floor][1] = true
				} else if hallCall.Direction == elevators.DirectionDown {
					lights[hallCall.Floor][0] = true
				} else if hallCall.Direction == elevators.DirectionBoth {
					lights[hallCall.Floor][1] = true
					lights[hallCall.Floor][0] = true
				}
			}
		}

		for floor, value := range lights {
			elevio.SetButtonLamp(elevio.BT_HallDown, floor, value[0])
			elevio.SetButtonLamp(elevio.BT_HallUp, floor, value[1])
		}

		<-store.ShouldRecalculateHCLightsChannel
	}
}

func goToFloor(destinationFloor int) {

	direction := elevators.DirectionIdle
	currentFloor, _ := store.GetCurrentFloor(selfHostname)
	if currentFloor < destinationFloor {
		direction = elevators.DirectionUp
	} else if currentFloor > destinationFloor {
		direction = elevators.DirectionDown
	} else {
		// We dont have to move since we are already on floor
		store.RemoveCabCall(selfHostname, currentFloor)
		store.RemoveHallCalls(selfHostname, currentFloor)

		return
	}

	elevio.SetMotorDirection(direction)
	store.SetDirectionMoving(selfHostname, direction)
	for {
		select {
		case floor := <-drv_floors: // Wait for elevator to reach floor
			elevio.SetFloorIndicator(floor)
			// Clear everything on this floor
			store.SetCurrentFloor(selfHostname, floor)

			order_distributor.ShouldSendStateUpdate <- true

			if floor == destinationFloor {
				store.RemoveCabCall(selfHostname, floor)
				store.RemoveHallCalls(selfHostname, floor)
				elevio.SetMotorDirection(elevators.DirectionIdle)
				store.SetDirectionMoving(selfHostname, elevators.DirectionIdle)
				elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
				openAndCloseDoors(floor)
				order_distributor.ShouldSendStateUpdate <- true
				return
			}
		case <-time.After(10 * time.Second):
			elevio.SetMotorDirection(elevators.DirectionIdle)
			// From the specification it is ok to kill the program
			log.Fatal("Didn't reach floor in time!\n")
		}
	}
}

func openAndCloseDoors(floor int) {
	elevio.SetDoorOpenLamp(true)
	time.Sleep(2 * time.Second)
	elevio.SetDoorOpenLamp(false)
}

func btnDirToElevDir(btn elevio.ButtonType) elevators.Direction_e {
	switch btn {
	case elevio.BT_HallDown:
		return elevators.DirectionDown
	case elevio.BT_HallUp:
		return elevators.DirectionUp
	default:
		return elevators.DirectionIdle
	}
}
