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

}
