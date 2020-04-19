package ring

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"../bcast"
	"../messages"
	"../peers"
)

const gBCASTPORT = 6971
const gBroadcastIP = "255.255.255.255"
const gConnectAttempts = 5
const gTIMEOUT = 2
const gJOINMESSAGE = "JOIN"
const NodeChange = "NodeChange"

var DisconnectedPeer = make(chan string)

// Initializes the network if it's present. Establishes a new network
func Init(innPort string) error {
	peersError := peers.Init(innPort)
	fmt.Println("Started peers server")
	if peersError != nil {
		fmt.Println("Error starting peers server")
		fmt.Println(peersError)
		return peersError
	}
	messages.Init(innPort)
	fmt.Println("Started messages")
	go handleJoin(innPort)
	go ringWatcher()
	fmt.Println("Starting ring...")
	return nil
}

//////////////////////////////////////////////
/// Exposed functions for sending and reciving
//////////////////////////////////////////////

func BroadcastMessage(purpose string, data []byte) bool {
	return messages.SendMessage(purpose, data)
}

func GetReceiver(purpose string) chan []byte {
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
func handleJoin(innPort string) {
	readChn := make(chan string, 10)
	go nonBlockingRead(readChn)
	for {
		select {
		case tail := <-readChn:
			if !peers.IsHead() {
				break
			}
			if !peers.AddTail(tail) {
				break
			}
			if peers.IsNextTail() {
				messages.ConnectTo(tail)
			}
			nodes := peers.GetAll()
			nodesBytes, _ := json.Marshal(nodes)
			messages.SendMessage(NodeChange, nodesBytes)
			break

		case <-time.After(5 * time.Second): // Listens for new elevators on the network
			if peers.IsAlone() {
				sendJoinMSG(innPort)
			}
			break
		}
	}
}

// Uses UDP broadcast to notify any existing ring about its presens
func sendJoinMSG(innPort string) {
	connWrite := bcast.DialBroadcastUDP(gBCASTPORT)
	defer connWrite.Close()

	for i := 0; i < gConnectAttempts; i++ {
		selfIP := peers.GetRelativeTo(peers.Self, 0)
		addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", gBroadcastIP, gBCASTPORT))
		message := gJOINMESSAGE + "-" + selfIP
		connWrite.WriteTo([]byte(message), addr)
		time.Sleep(gTIMEOUT * time.Second) // wait for response
		if !peers.IsAlone() {
			return
		}
	}
}

// Handles ring growth and shrinking
// Detects if the node infront of you disconnects, alerts rest of ring
// That node becomes the master
func ringWatcher() {
	nodeChangeReciver := messages.GetReceiver(NodeChange)

	for {
		select {
		case disconnectedIP := <-messages.DisconnectedFromServerChannel: // Doesn't get triggered on connect

			peers.Remove(disconnectedIP)
			if peers.IsAlone() {
				break
			}
			peers.BecomeHead()
			nextNode := peers.GetRelativeTo(peers.Self, 1)
			messages.ConnectTo(nextNode)

			nodeList := peers.GetAll()
			nodeBytes, _ := json.Marshal(nodeList)
			messages.SendMessage(NodeChange, nodeBytes)
			DisconnectedPeer <- disconnectedIP
			break

		// Never more then one disconect or add per change
		case nodeBytes := <-nodeChangeReciver:
			nodeList := []string{""}
			json.Unmarshal(nodeBytes, &nodeList)
			if !peers.IsEqualTo(nodeList) {
				oldNodes := peers.GetAll()
				disconnectedIP, shouldRemove := difference(oldNodes, nodeList)
				if shouldRemove {
					DisconnectedPeer <- disconnectedIP
				}
				peers.Set(nodeList)
				nextNode := peers.GetRelativeTo(peers.Self, 1)
				messages.ConnectTo(nextNode)
				messages.SendMessage(NodeChange, nodeBytes)
			}
			break
		}
	}
}

// Makes it possible to have timeout on udp read
func nonBlockingRead(readChn chan<- string) { // This is iffy, was a quick fix
	buffer := make([]byte, 100)
	connRead := bcast.DialBroadcastUDP(gBCASTPORT)

	defer connRead.Close()
	for {
		nBytes, _, _ := connRead.ReadFrom(buffer[0:]) // Fuck errors
		msg := string(buffer[:nBytes])
		splittedMsg := strings.SplitN(msg, "-", 2)
		self := peers.GetRelativeTo(peers.Self, 0)
		receivedJoin := splittedMsg[0]
		receivedHost := splittedMsg[1]
		if receivedJoin == gJOINMESSAGE && receivedHost != self { // Hmmmmmm
			readChn <- receivedHost
		}
	}
}

// Slice1 is longer than slice2
func difference(slice1 []string, slice2 []string) (string, bool) {
	if len(slice1) <= len(slice2) {
		return "", false
	}
	for _, s1 := range slice1 {
		found := false
		for _, s2 := range slice2 {
			if s1 == s2 {
				found = true
				break
			}
		}
		// String not found.
		if !found {
			return s1, true
		}
	}
	return "", false
}
