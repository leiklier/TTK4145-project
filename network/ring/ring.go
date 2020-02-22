package ring

import (
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
