package ring

import (
	"fmt"
	"net"

	"../peers"
)

const PORT = "1567"

var HEAD = true

func Listenjoin() {
	if !HEAD { // Hmmm
		return
	}
	buffer := make([]byte, 1024) // What happens if packet > buffer

	ln, err := net.ListenPacket("udp", "255.255.255.255:"+PORT)

	if err != nil {
		fmt.Println("Unable to listen to udp broadcast")
		fmt.Println(err)
	}

	defer ln.Close()

	for {
		n_bytes, addr, err := ln.ReadFrom(buffer)
		peers.AddTail(string(buffer[:n]))
		err := ring.Broadcast(peers.GetAll)
		if err != nil {
			fmt.Println("Failed to broadcast")
			fmt.Println(err)
		}

	}
	"../p2p"
)

var gPort int

// Ring module. Handles

// Listens to broadcast messages comming from the other computers in the ring.
// Connecs to previous node in ring, listens for a msg and send that msg to

func listenBroadcast() {
	//prevIP := peers.GetRelativeTo(peers.Self, -1)
	// Setup server
	p2p.Start(gPort)

}
