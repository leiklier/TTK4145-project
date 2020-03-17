package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	"./event_handler"

	"./network/ring"
	"./sync/store"
)

func main() {
	innport := os.Args[1]
	outport := os.Args[2]
	elevNumberStr := os.Args[3]
	elevNumber, _ := strconv.Atoi(elevNumberStr)

	err := ring.Init(innport, outport)
	store.Init()
	// if elevNumber == 0 {
	// 	 spawnElevators()

	// }
	if err != nil {
		fmt.Println(err)
	}
	// for {
	// 	select {
	// 	case <-time.After(100 * time.Second):
	// 		break
	// 	}
	// }

	event_handler.RunElevator(elevNumber)

}

func spawnElevators() {

	err := (exec.Command("gnome-terminal", "-e", "./spawnElevators.sh")).Run()
	if err != nil {
		fmt.Println("Something went wrong!")
		log.Fatal(err)
	}

}
