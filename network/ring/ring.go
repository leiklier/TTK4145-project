package ring

import (
	"fmt"
	"net"

	"../p2p"
	"../peers"
)

const gPORT = 1567
const gBroadcastIP = "255.255.255.255"

var HEAD = true // Has to be changed
const gConnectAttempts = 5
const gBCASTPING = "RING EXISTS"

func Init() {
	var ringExists = false
	ln, err := net.ListenPacket("udp", "255.255.255.255:"+PORT) // Maybe have helper function returning ln
	buffer := make([]byte, 1024)                                // What happens if packet > buffer

	if err != nil {
		fmt.Println("Unable to listen to udp broadcast")
		fmt.Println(err)
	}

	defer ln.Close()

	for i := 0; i < gConnectAttempts; i++ {
		nBytes, addr, err := ln.ReadFrom(buffer)
		if nBytes > 0 {
			ringExists = true
		}
		
	}

	}

}


// Only runs if you are HEAD, listen for new machines broadcasting
// on the network using UDP. The new machine is added to the list of
// known machines. That list is propagted trpough the ring to update the ring
func Listenjoin() {
	if !HEAD { // Hmmm, has to be changed
		return
	}
	buffer := make([]byte, 1024) // What happens if packet > buffer

	ln, err := net.ListenPacket("udp", "255.255.255.255:"+gPORT)

	if err != nil {
		fmt.Println("Unable to listen to udp broadcast")
		fmt.Println(err)
	}

	defer ln.Close()

	for {
		nBytes, addr, err := ln.ReadFrom(buffer)
		if nBytes > 0 {
			peers.AddTail(string(buffer[:nBytes]))
			err = ring.Broadcast(peers.GetAll())
			if err != nil {
				fmt.Println("Failed to broadcast")
				fmt.Println(err)
			}
		}

	}
}

var gPort int

// Ring module. Handles

// Listens to broadcast messages comming from the other computers in the ring.
// Connecs to previous node in ring, listens for a msg and send that msg to

func listenBroadcast() {
	//prevIP := peers.GetRelativeTo(peers.Self, -1)
	// Setup server
	p2p.Start(gPort)

}
