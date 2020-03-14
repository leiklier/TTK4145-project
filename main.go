package main

import (
	"./event_handler"
	"./network/ring"
)

func main() {
	ring.Init()                 // Need to be called explicitly to establish connection before first call
	event_handler.RunElevator() // Yeet this whole function into main? Sverre says so, but its a bit weird imo

}
