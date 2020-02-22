package ring

import (
	"fmt"
	"net"
	"os"
	"syscall"
	"time"

	"../peers"
)

const gPORT = 1567
const gBroadcastIP = "255.255.255.255"

var gHEAD = false // Has to be changed
const gConnectAttempts = 5
const gBCASTPING = "RING EXISTS"
const gTIMEOUT = 2

// Initializes the network if it's present. Establishes a new network if not
func Init() {
	var ringExists = false
	buffer := make([]byte, 1024) // What happens if packet > buffer
	conn := dialBroadcastUDP(gPORT)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", gBroadcastIP, gPORT))
	conn.SetReadDeadline(time.Now().Add(gTIMEOUT)) // Timeout after 2 sec

	defer conn.Close()

	for i := 0; i < gConnectAttempts; i++ {
		nBytes, addr, err := conn.ReadFrom(buffer)
		if nBytes > 0 {
			if string(buffer[:nBytes]) == gBCASTPING {
				ringExists = true
				break
			}
		}
	}
	if ringExists {
		sendJoinMSG()
		// Start tcp ring server?
	} else {
		gHEAD = true
		go Listenjoin()
	}
	// go peers.Server()

}

// Only runs if you are HEAD, listen for new machines broadcasting
// on the network using UDP. The new machine is added to the list of
// known machines. That list is propagted trpough the ring to update the ring
func Listenjoin() {
	if !gHEAD { // Hmmm, has to be changed
		return
	}
	buffer := make([]byte, 1024) // What happens if packet > buffer

	conn := dialBroadcastUDP(gPORT)
	conn.SetReadDeadline(time.Now().Add(gTIMEOUT)) // Timeout after 2 sec

	defer conn.Close()

	for {
		nBytes, addr, err := conn.ReadFrom(buffer)
		if nBytes > 0 {
			peers.AddTail(string(buffer[:nBytes]))
			err = ring.addNode(peers.GetAll())
			if err != nil {
				fmt.Println("Failed to broadcast")
				fmt.Println(err)
			}
		}

	}
}

func addNode(peersList []string) {
	// peers.Send(msg)
}

func listenBroadcast() {

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
