package ring

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"
	"time"

	"../../store"
	"../messages"
	"../peers"
)

const ( // Ugh
	NodeChange = "NodeChange"
	CallList   = "CallList"
	Maintain   = "Maintain" //ping ish?
)

const gBCASTPORT = 6971
const gRINGPORT = 6972
const gBroadcastIP = "255.255.255.255"
const gConnectAttempts = 1
const gBCASTPING = "RING EXISTS"
const gTIMEOUT = 2
const gJOINMESSAGE = "JOIN"

// Initializes the network if it's present. Establishes a new network if not
func Init() {

	go listenCallsIn()
	go listenCallsOut()
	go maintainRing()
	go neighbourWatcher()
	go handleRingChange()

	sendJoinMSG()
	listenjoin()

}

// send msg to 255
func sendJoinMSG() bool { //TODO: make goroutine
	// Connect til
	connWrite := dialBroadcastUDP(gBCASTPORT)
	defer connWrite.Close()

	for i := 0; i < gConnectAttempts; i++ {

		selfIP := peers.GetRelativeTo(peers.Self, 0)
		addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", gBroadcastIP, gBCASTPORT))
		connWrite.WriteTo([]byte(gJOINMESSAGE+":"+selfIP), addr)
		time.Sleep(gTIMEOUT * time.Second)
		fmt.Println("Sent")
		if peers.GetRelativeTo(peers.Head, 0) != peers.GetRelativeTo(peers.Self, 0) {
			return true
		}
	}
	return false
}

// Only runs if you are HEAD, listen for new machines broadcasting
// on the network using UDP. The new machine is added to the list of
// known machines. That list is propagted trpough the ring to update the ring
func listenjoin() {

	buffer := make([]byte, 100)
	connRead := dialBroadcastUDP(gBCASTPORT)

	defer connRead.Close()

	for {
		if peers.GetRelativeTo(peers.Head, 0) == peers.GetRelativeTo(peers.Self, 0) {
			n, _, _ := connRead.ReadFrom(buffer[0:])
			msg := string(buffer[:n])
			splittedMsg := strings.SplitN(msg, ":", 2)

			if splittedMsg[0] == gJOINMESSAGE {
				fmt.Println("New node on the network")
				nodes, _ := json.Marshal(peers.GetAll())
				peers.AddTail(splittedMsg[1])
				if peers.GetRelativeTo(peers.Tail, -1) == peers.GetRelativeTo(peers.Self, 0) {
					messages.ConnectTo(peers.GetRelativeTo(peers.Self, 1))
				} else {
					messages.SendMessage(NodeChange, nodes)
				}
			}
		}
	}
}

func handleRingChange() {
	var nodesList []string
	fmt.Println("Updating ring...")
	fmt.Println(peers.GetAll())
	nodes := messages.Receive(NodeChange)
	json.Unmarshal(nodes, &nodesList)
	peers.Set(nodesList)
	nextNode := peers.GetRelativeTo(peers.Self, 1)
	messages.ConnectTo(nextNode)
	messages.SendMessage(NodeChange, nodes)
}

func maintainRing() { // kind of ping??
	// var nodesList []string
	messages.Start()
	for {
		time.Sleep(gTIMEOUT * time.Second)
		// fmt.Println("Maintaining ring...")
		if peers.GetRelativeTo(peers.Head, 0) == peers.GetRelativeTo(peers.Self, 0) {
			messages.SendMessage(Maintain, []byte("Her kan vi sende noe lurt\000"))
		} else {
			nodes := messages.Receive(Maintain)
			fmt.Println(string(nodes))
			messages.SendMessage(Maintain, nodes)
		}
	}
}

// Detects if the node infront of you disconnects, alerts rest of ring
// That node becomes the master
func neighbourWatcher() {
	missingIP := messages.ServerDisconnected()
	peers.Remove(missingIP)
	peers.BecomeHead()
	nextNode := peers.GetRelativeTo(peers.Self, 1)
	messages.ConnectTo(nextNode)
	nodeList := peers.GetAll()
	nodes, _ := json.Marshal(nodeList)
	messages.SendMessage(NodeChange, nodes)
}

func listenCallsIn() {
	for {
		msg := messages.Receive(CallList)
		fmt.Println(string(msg))
		state := store.ElevatorState{}
		json.Unmarshal(msg, &state)
		store.RecieveElevState <- state
	}
}

func listenCallsOut() {
	for {
		states := <-store.SendElevState
		statesBytes, _ := json.Marshal(&states)
		messages.SendMessage(CallList, statesBytes)
	}
}

////////////////////////////////////////////////////
// Helper functions
////////////////////////////////////////////////////

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
