package ring

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"
	"time"

	"../messages"
	"../peers"
)

const gBCASTPORT = 6971
const gBroadcastIP = "255.255.255.255"
const gConnectAttempts = 5
const gTIMEOUT = 2
const gJOINMESSAGE = "JOIN"
const NodeChange = "NodeChange"

var NewNeighbourNode = make(chan string)

// Initializes the network if it's present. Establishes a new network if not
func Init() {
	messages.Start()
	go ringWatcher()
	go handleJoin()
}

//////////////////////////////////////////////
/// Exposed functions for sending and reciving
//////////////////////////////////////////////

func BroadcastMessage(purpose string, data []byte) bool {
	return messages.SendMessage(purpose, data)
}

func GetReceiver(purpose string) chan<- []byte {
	return messages.GetReceiver(purpose)
}

func SendToPeer(purpose string, ip string, data []byte) bool {
	dataMap := make(map[string][]byte)
	dataMap[ip] = data
	dataMapbytes, _ := json.Marshal(dataMap)
	return messages.SendMessage(purpose, dataMapbytes)
}

//////////////////////////	///////////////////////
/// Functions for setting up and maintaining ring
/////////////////////////////////////////////////

// Only runs if you are HEAD, listen for new machines broadcasting
// on the network using UDP. The new machine is added to the list of
// known machines. That list is propagted trpough the ring to update the ring
func handleJoin() {
	readChn := make(chan string)
	go nonBlockingRead(readChn)
	for {
		select {
		case tail := <-readChn:
			if !peers.IsHead() {
				break
			}

			peers.AddTail(tail)
			if peers.IsNextTail() {
				messages.ConnectTo(tail)
			}
			nodes := peers.GetAll()
			nodesBytes, _ := json.Marshal(nodes)
			messages.SendMessage(NodeChange, nodesBytes)
			break

		case <-time.After(10 * time.Second): // Listens for new elevators on the network
			if peers.IsAlone() {
				sendJoinMSG()
			}
			break
		}
	}
}

// Handles ring growth and shrinking
// Detects if the node infront of you disconnects, alerts rest of ring
// That node becomes the master
func ringWatcher() {
	var nodesList []string
	var disconnectedIP string

	nodeChangeReciver := messages.GetReceiver(NodeChange)

	for {
		select {
		case disconnectedIP = <-messages.DisconnectedFromServerChannel:
			fmt.Printf("Disconnect : %s\n", disconnectedIP)

			peers.Remove(disconnectedIP)
			peers.BecomeHead()
			nextNode := peers.GetNextPeer()
			messages.ConnectTo(nextNode)

			nodeList := peers.GetAll()
			nodeBytes, _ := json.Marshal(nodeList)
			messages.SendMessage(NodeChange, nodeBytes)
			break

		case nodeBytes := <-nodeChangeReciver:
			json.Unmarshal(nodeBytes, &nodesList)
			if !peers.IsEqualTo(nodesList) {
				peers.Set(nodesList)
				nextNode := peers.GetNextPeer()
				messages.ConnectTo(nextNode)
				messages.SendMessage(NodeChange, nodeBytes)
			}
			break
		case addedNode := <-peers.AddedNextChannel:
			NewNeighbourNode <- addedNode
			break
		}
	}
}

// Uses UDP broadcast to notify any existing ring about its presens
func sendJoinMSG() {
	connWrite := dialBroadcastUDP(gBCASTPORT)
	defer connWrite.Close()

	for i := 0; i < gConnectAttempts; i++ {
		selfIP := peers.GetRelativeTo(peers.Self, 0)
		addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", gBroadcastIP, gBCASTPORT))
		connWrite.WriteTo([]byte(gJOINMESSAGE+":"+selfIP), addr)
		time.Sleep(gTIMEOUT * time.Second) // wait for response
		if !peers.IsAlone() {
			return
		}
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

// Makes it possible to have timeout on udp read
func nonBlockingRead(readChn chan<- string) { // This is iffy, was a quick fix
	buffer := make([]byte, 100)
	connRead := dialBroadcastUDP(gBCASTPORT)

	defer connRead.Close()
	for {
		nBytes, _, _ := connRead.ReadFrom(buffer[0:])
		msg := string(buffer[:nBytes])
		splittedMsg := strings.SplitN(msg, ":", 2)
		selfIp := peers.GetRelativeTo(peers.Self, 0)
		receivedJoin := splittedMsg[0]
		receivedIP := splittedMsg[1]

		if receivedJoin == gJOINMESSAGE && receivedIP != selfIp { // Hmmmmmm
			readChn <- receivedIP
		}
	}
}
