package ring

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"syscall"

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

	ln, err := net.ListenPacket("udp", "255.255.255.255:"+strconv.Itoa(gPORT))

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

// send msg to 255
func sendJoinMSG() {
	// Connect til
	gselfIP := peers.GetRelativeTo(peers.Self, 0)
	conn := dialBroadcastUDP(gPORT)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", gBroadcastIP, gPORT))
	conn.WriteTo([]byte(gselfIP), addr)
}

// Tar inn port, returnerer en udpconn til porten.
func dialBroadcastUDP(port int) net.PacketConn {
	s, _ := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
	syscall.Bind(s, &syscall.SockaddrInet4{Port: port})

	f := os.NewFile(uintptr(s), "")
	conn, _ := net.FilePacketConn(f)
	f.Close()

	return conn
}
