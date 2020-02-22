package ring

import (
	"fmt"
	"net"
	"os"
	"syscall"
	"time"

	"../messages"
	"../peers"
)

// type DataType struct {
// 	type
// 	data []byte
// }

// type Message struct {
// 	Type Broadcast Ping PingAck
// 	Purpose: Custom string
// 	Data
// }

// messages.ReceiveBroadcast("Custom string")
// messages.ReceivePingAck

const ( // Ugh
	NodeChange = "NodeChange"
	CallList   = "CallList"
)

const gPORT = 1567
const gBroadcastIP = "255.255.255.255"

var gHEAD = false // Has to be changed
const gConnectAttempts = 5
const gBCASTPING = "RING EXISTS"
const gTIMEOUT = 2


order_chn := make(chan store.ElevatorOrder)

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
	} else {
		gHEAD = true
		go Listenjoin()
		peers.GetAll() // Sets also the head and tail
	}

	go maintainRing()
	go listenCalls()
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
			msg := messages.Message{messages.Broadcast, AddNode, []byte(peersList)}
			messages.Send(msg)
			if err != nil {
				fmt.Println("Failed to broadcast")
				fmt.Println(err)
			}
		}
	}
}

func maintainRing() {
	for {
		msg := messages.Recieve(NodeChange)
		peers.Set(string(msg.Data))
		messages.Send(msg)
	}
}


func listenCallsIn() {
	for {
		msg := messages.Receive(CallList)
		store.reciveElevState <- msg.Data
	}
}


func listenCallsOut() {
	for{
		 states <-store.SendElevState:
			msg := messages.Message{CallList, messages.Broadcast, states}
			messages.Send(msg)
			break
		}
	}
}

// send msg to 255
func sendJoinMSG() {
	// Connect til
	selfIP := peers.GetRelativeTo(peers.Self, 0)
	conn := dialBroadcastUDP(gPORT)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", gBroadcastIP, gPORT))
	conn.WriteTo([]byte(selfIP), addr)
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
