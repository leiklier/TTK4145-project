package event_handler

import (
	"fmt"
	"log"
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

// var nextFloor = make(chan int)

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

	store.SetCurrentFloor(selfIP, store.NumFloors-1)

	// go nextfloor.SubscribeToDestinationUpdates(nextFloor)
	go elevatorDriver()
	go hcLightDriver()
	go buttonHandler()

}

func buttonHandler() {
	// Reset all Cab call lamps:
	cabCalls, _ := store.GetAllCabCalls(selfIP)
	for floor, cabCall := range cabCalls {
		elevio.SetButtonLamp(elevio.BT_Cab,floor,cabCall)
	}

	for {
		buttonEvent := <-drv_buttons
		fmt.Println(buttonEvent)
		currentFloor, _ := store.GetCurrentFloor(selfIP)

		if buttonEvent.Floor == currentFloor && buttonEvent.Button != elevio.BT_Cab {
			openAndCloseDoors(currentFloor)
			continue
		}
		// Håndtere callen
		if buttonEvent.Button == elevio.BT_Cab {
			if buttonEvent.Floor != currentFloor {
				// Cab call
				elevio.SetButtonLamp(elevio.BT_Cab, buttonEvent.Floor, true)
				store.AddCabCall(selfIP, buttonEvent.Floor)
			}
		} else {
			// Hall call
			// Create HallCall
			elevDir := btnDirToElevDir(buttonEvent.Button)
			hCall := elevators.HallCall_s{Floor: buttonEvent.Floor, Direction: elevDir}
			if store.IsExistingHallCall(hCall) {
				continue
			}
			mostSuitedIP := store.MostSuitedElevator(buttonEvent.Floor, elevDir)
			fmt.Printf("Most suited ip: %s\n", mostSuitedIP)

			store.AddHallCall(mostSuitedIP, hCall)
			order_distributor.SendHallCall(mostSuitedIP, hCall)
		}
	}
}

func elevatorDriver() {
	goToFloor(0)

	for {
		// nextFloor := <-nextFloorChan
		currentFloor, _ := store.GetCurrentFloor(selfIP)
		if store.IsExistingHallCall(elevators.HallCall_s{Floor: currentFloor, Direction: elevators.DirectionBoth}) {
			openAndCloseDoors(currentFloor)
			// clear hall call

			store.RemoveHallCalls(selfIP, currentFloor)
			select {
			case order_distributor.SendStateUpdate <- true:
			default:
			}
		}

		nextFloor := nextfloor.SubscribeToDestinationUpdates()
		// fmt.Printf("nextFloor: %d\n", nextFloor)
		if nextFloor != -1 {
			goToFloor(nextFloor)
		}
		<-store.ShouldRecalculateNextFloorChannel
		// Send update/state
	}
}

func hcLightDriver() {
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
	currentFloor, _ := store.GetCurrentFloor(selfIP)
	if currentFloor < destinationFloor {
		direction = elevators.DirectionUp
	} else if currentFloor > destinationFloor {
		direction = elevators.DirectionDown
	} else {
		// WE DONT HAVE TO MOVE SINCE WE ARE ALREADY HERE
		// fmt.Println("Same floor idiot")
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

			select {
			case order_distributor.SendStateUpdate <- true:
			default:
			}

			if floor == destinationFloor {
				store.RemoveCabCall(selfIP, floor)
				store.RemoveHallCalls(selfIP, floor)
				elevio.SetMotorDirection(elevators.DirectionIdle) // Stop elevator and set lamps and stuff
				store.SetDirectionMoving(selfIP, elevators.DirectionIdle)

				openAndCloseDoors(floor)
				select {
				case order_distributor.SendStateUpdate <- true:
				default:
				}
				return
			}
			break
		case <-time.After(10 * time.Second):
			elevio.SetMotorDirection(elevators.DirectionIdle)
			//Do some shit
			log.Fatal("Didn't reach floor in time!\n")
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
