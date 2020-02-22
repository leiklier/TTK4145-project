package main

import (
	"fmt"
	"time"

	"./network/messages"
)

func main() {
	messages.Start()
	go receiveMessages()
	time.Sleep(4 * time.Second)
	messages.ConnectTo("10.100.23.203")
	for {
		messages.SendMessage("blabla", []byte("heyhey"))
		time.Sleep(2 * time.Second)
	}
}

func receiveMessages() {
	for {
		newMessage := messages.Receive("blabla")
		fmt.Printf("Received:", string(newMessage))
	}
}
