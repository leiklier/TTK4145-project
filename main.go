package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"./event_handler"

	"./network"
	"./order_distributor"
	"./store"
)

func main() {
	inport := os.Args[1]
	elevNumberStr := os.Args[2]
	elevNumber, _ := strconv.Atoi(elevNumberStr)

	// Establishes the ring network
	err := network.Init(inport)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	// Holds information about all elevators
	store.Init()

	// Passes orders between elevators
	order_distributor.Init()

	// Handles button events, motor direction and lights
	event_handler.Init(elevNumber)

	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
