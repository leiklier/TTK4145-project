package ring

import (
	"fmt"
	"net"
	"os"
	"syscall"

	"../peers"
)

const gPORT = "1567"
const gbroadcastIP = "255.255.255.255"

var HEAD = true

var gselfIP string

func Listenjoin() {
	if !HEAD { // Hmmm
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
		n_bytes, addr, err := ln.ReadFrom(buffer)
		peers.AddTail(string(buffer[:n]))
		err = ring.Broadcast(peers.GetAll)
		if err != nil {
			fmt.Println("Failed to broadcast")
			fmt.Println(err)
		}

	}
}

// send msg to 255, if head du får svar
func sendJoinMSG() {
	// Connect til
	gselfIP := peers.GetRelativeTo(peers.Self, 0)
	conn := dialBroadcastUDP(gPort)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", _broadcastIP, port))
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
