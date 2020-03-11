package main

import (
	"./event_handler"
	"./network/ring"
)

func main() {
	ring.Init()
	event_handler.RunElevator()

}
