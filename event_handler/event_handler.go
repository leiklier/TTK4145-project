package event_handler

import (
	"fmt"
	"strconv"
	"time"

	"../elevio"
	"../network/peers"
	"../order_distributor"
	"../sync/elevators"
	"../sync/nextfloor"
	"../sync/store"
)

var selfIP string

var drv_buttons = make(chan elevio.ButtonEvent)
var drv_floors = make(chan int)
var drv_obstr = make(chan bool)
var drv_stop = make(chan bool)
var nextFloor = make(chan int)

// Init Her skjer det
func Init(elevNumber int) {

	// First we start the server
	fmt.Println("Starting elevator server ...")
	selfIP = peers.GetRelativeTo(peers.Self, 0)
	connPort := strconv.Itoa(15657 + elevNumber)
	time.Sleep(time.Duration(1 * time.Second)) // To avoid crash due to not started sim
	elevio.Init("localhost:"+connPort, store.NumFloors)

	go elevio.PollButtons(drv_buttons) // Etasje og hvilken type knapp som blir trykket
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	store.SetCurrentFloor(selfIP, store.NumFloors)

	go nextfloor.SubscribeToDestinationUpdates(nextFloor)
	go elevatorDriver(nextFloor)
	go buttonHandler()

}

func buttonHandler() {
	// Reset all Cab call lamps:
	//cabCalls, _ := store.GetAllCabCalls(selfIP)
	//for cabCall, floor

	for {
		buttonEvent := <-drv_buttons
		fmt.Println(buttonEvent)
		// Håndtere callen
		if buttonEvent.Button == elevio.BT_Cab {
			// Cab call
			elevio.SetButtonLamp(elevio.BT_Cab, buttonEvent.Floor, true)
			store.AddCabCall(selfIP, buttonEvent.Floor)
		} else {
			// Hall call
			elevDir := btnDirToElevDir(buttonEvent.Button)
			mostSuitedIP := store.MostSuitedElevator(buttonEvent.Floor, elevDir)

			// Create and send HallCall
			hCall := elevators.HallCall_s{Floor: buttonEvent.Floor, Direction: elevDir}
			if mostSuitedIP == selfIP {
				store.AddHallCall(selfIP, hCall)
			}
			order_distributor.SendHallCall(mostSuitedIP, hCall)
		}
		// Send update/state
	}
}

func elevatorDriver(nextFloorChan chan int) {
	goToFloor(0)

	for {
		nextFloor := <-nextFloorChan
		fmt.Printf("nextFloor: %d\n", nextFloor)
		goToFloor(nextFloor)
		// Send update/state
	}
}

func goToFloor(destinationFloor int) {

	direction := elevators.DirectionIdle
	currentFloor, _ := store.GetCurrentFloor(selfIP)
	if currentFloor < destinationFloor {
		direction = elevators.DirectionUp
	} else if currentFloor > destinationFloor {
		direction = elevators.DirectionDown
	} else {
		// WE DONT HAVE TO MOVE SINCE WE ARE ALREADY HERE
		store.RemoveCabCall(selfIP, currentFloor)
		store.RemoveHallCalls(selfIP, currentFloor)

		return
	}

	elevio.SetMotorDirection(direction)
	store.SetDirectionMoving(selfIP, direction)
	for {
		select {
		case floor := <-drv_floors: // Wait for elevator to reach floor
			elevio.SetFloorIndicator(floor)
			// CLear everything onn this floor
			store.SetCurrentFloor(selfIP, floor)
			store.RemoveCabCall(selfIP, floor)
			store.RemoveHallCalls(selfIP, floor)


			if floor == destinationFloor {
				elevio.SetMotorDirection(elevators.DirectionIdle) // Stop elevator and set lamps and stuff
				store.SetDirectionMoving(selfIP, elevators.DirectionIdle)

				openAndCloseDoors(floor)
				return
			}
			break
		case <-time.After(10 * time.Second):
			fmt.Println("Didn't reach floor in time!")
			elevio.SetMotorDirection(elevators.DirectionIdle)
			//Do some shit
			return
			// break
		}
	}

}

func openAndCloseDoors(floor int) {
	elevio.SetFloorIndicator(floor)
	elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
	elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
	elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
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

func DetermineLight(floor int, button elevio.ButtonType) bool {
	localElevator, _ := store.GetElevator(selfIP)
	if floor == localElevator.GetCurrentFloor() && localElevator.GetDirectionMoving() == 0 {
		return false // If elevator is standing still and at floor, dont accept
	}
	return true
}
