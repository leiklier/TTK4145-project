package main

import (
	"fmt"

	"./network/peers"
)

func main() {
	peers.AddTail("Node 1")
	peers.AddTail("Node 2")
	peers.AddTail("Node 3")
	fmt.Printf("Self is %s\n", peers.GetRelativeTo(peers.Self, 0))
	fmt.Printf("Current HEAD is %s\n", peers.GetRelativeTo(peers.Head, -1))
	fmt.Printf("Current Tail is %s\n", peers.GetRelativeTo(peers.Tail, 2))
	fmt.Println(peers.GetAll())
}

func printChanges() {
	for {
		fmt.Printf("A node was %s")
	}
}
