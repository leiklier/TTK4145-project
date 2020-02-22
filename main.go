package main

import (
	"bufio"
	"fmt"
	"os"

	"./network/peers"
)

func main() {
	go printChanges()
	peers.AddTail("Node 1")
	peers.AddTail("Node 2")
	peers.AddTail("Node 3")
	peers.Remove("Node 2")
	peers.Set([]string{"Node 5", "Node 6", "Node 7"})
	fmt.Println(peers.GetAll())
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func printChanges() {
	for {
		changeEvent := peers.PollUpdate()
		fmt.Printf("A node with IP %s was %s\n", changeEvent.Peer, changeEvent.Event)
		fmt.Printf("\n")
	}
}
