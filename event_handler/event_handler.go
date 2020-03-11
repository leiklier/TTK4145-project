package event_handler

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	"../elevio"
	"../network/peers"
	"../order_distributor"
	"../sync/elevators"
	"../sync/store"
)

const numFloors = 4

// RunElevator Her skjer det
func RunElevator() {
	// Warning for windows users
	if runtime.GOOS == "windows" {
		fmt.Println("Can't Execute this on a windows machine")
		os.Exit(3)
	}
	// First we start the server
	fmt.Println("Starting elevator server ...")
	err := (exec.Command("gnome-terminal", "-x", "/home/student/ElevatorServer")).Run()
	if err != nil {
		fmt.Println("Something went wrong!")
		log.Fatal(err)
	}

	// one goroutine to update store from driver
	// one goroutine to run elevator based on store

	elevio.Init("localhost:15657", numFloors)
	selfIP := peers.GetRelativeTo(peers.Self, 0)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	// dst := make(chan store.Command)
	update := make(chan bool)
	nextFloor := make(chan int)

	go elevio.PollButtons(drv_buttons) // Etasje og hvilken type knapp som blir trykket
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	var d elevators.Direction_e

	time.Sleep(time.Duration(2 * time.Second))
	fmt.Println("Elevator server is running")

	// Initialize all elevators at the bottom when the program is first run.
	goToFloor(0, numFloors, drv_floors)

	for {
		select {
		case a := <-drv_buttons: // Just sets the button lamp, need to translate into calls
			fmt.Println("Noen trykket på en knapp, oh lø!")

			// Setter på lyset
			// light := store.DetermineLight(a.Floor, a.Button)
			// elevio.SetButtonLamp(a.Button, a.Floor, light)

			// Håndtere callen
			if a.Button == elevio.BT_Cab {
				store.AddCabCall(selfIP, a.Floor)
			} else {
				elevDir := BtnDirToElevDir(a.Button)
				mostSuitedIP := store.MostSuitedElevator(a.Floor, elevDir)

				// Create and send HallCall
				hc := elevators.HallCall_s{Floor: a.Floor, Direction: elevDir}
				order_distributor.SendHallCall(mostSuitedIP, hc)
			}

		// case a := <-drv_obstr: // Looks for obstruction and stops if true
		// 	fmt.Printf("%+v\n", a)
		// 	if a {
		// 		elevio.SetMotorDirection(elevio.MD_Stop)
		// 	} else {
		// 		elevio.SetMotorDirection(d)
		// 	}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}

		case a := <-update:
			if a == false {
				// Do nothing
			} else {

			}
		case a := <-nextFloor:
			go goToFloor(a.DstFloor, a.CurFloor, drv_floors) // Må endre parametre
		}
	}
}

func goToFloor(dest_floor int, current_floor int, drv_floors <-chan int) { // Probably add a timeout'

	elevio.SetDoorOpenLamp(false)

	if current_floor < dest_floor {
		elevio.SetMotorDirection(elevators.Elevator_s)
		store.UpdateDirectionState(store.Direction(elevators.Direction_e))
		for {
			select {
			case a := <-drv_floors: // Wait for elevator to reach floor
				elevio.SetFloorIndicator(a)
				if a == dest_floor {
					store.UpdateFloorState(a)
					elevio.SetMotorDirection(elevio.MD_Stop) // Stop elevator and set lamps and stuff
					store.UpdateDirectionState(store.Direction(elevio.MD_Stop))
					updateFromStore()

					elevio.SetDoorOpenLamp(true)
					store.OpenDoor(true)
					time.Sleep(3 * time.Second)
					store.OpenDoor(false)
					elevio.SetDoorOpenLamp(false)
					return
				}
			}
		}
	} else if current_floor > dest_floor {
		elevio.SetMotorDirection(elevio.MD_Down)
		store.UpdateDirectionState(store.Direction(elevio.MD_Down))
		for {
			select {
			case a := <-drv_floors: // Wait for elevator to reach floor
				elevio.SetFloorIndicator(a)
				if a == dest_floor {
					store.UpdateFloorState(a)
					elevio.SetMotorDirection(elevio.MD_Stop) // Stop elevator and set lamps and stuff
					store.UpdateDirectionState(store.Direction(elevio.MD_Stop))
					updateFromStore()

					elevio.SetDoorOpenLamp(true)
					store.OpenDoor(true)
					time.Sleep(3 * time.Second)
					store.OpenDoor(false)
					elevio.SetDoorOpenLamp(false)
					return
				}
			}
		}
	}
}

func updateFromStore() {
	floor, dir := store.GetFloorAndDir()
	elevio.SetFloorIndicator(floor)
	elevio.SetButtonLamp(elevio.BT_HallUp, floor, false)
	elevio.SetButtonLamp(elevio.BT_HallDown, floor, false)
	elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
	if dir == 0 {
		elevio.SetDoorOpenLamp(true)
	}
}
func BtnDirToElevDir(btn elevio.ButtonType) elevators.Direction_e {
	switch btn {
	case elevio.BT_HallDown:
		return elevators.DirectionDown
	case elevio.BT_HallUp:
		return elevators.DirectionUp
	default:
		fmt.Println("Invalid use, must be either up or down")
		return elevators.DirectionIdle
	}
}
