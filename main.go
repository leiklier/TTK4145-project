package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"bufio"
	"strconv"

	"./event_handler"

	"./network/ring"
	"./order_distributor"
	"./sync/store"
)


func main() {
	inport := os.Args[1]
	outport := os.Args[2]
	elevNumberStr := os.Args[3]
	elevNumber, _ := strconv.Atoi(elevNumberStr)

	err := ring.Init(inport, outport)
	store.Init()
	// if elevNumber == 0 {
	// 	 spawnElevators()

	// }
	if err != nil {
		fmt.Println(err)
	}

	order_distributor.Init()
	// go store.sPrintStateAll()
	event_handler.Init(elevNumber)

	bufio.NewReader(os.Stdin).ReadBytes('\n')

}

func spawnElevators() {

	err := (exec.Command("gnome-terminal", "-e", "./spawnElevators.sh")).Run()
	if err != nil {
		fmt.Println("Something went wrong!")
		log.Fatal(err)
	}

}
