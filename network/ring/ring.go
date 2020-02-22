package ring

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"syscall"
	"time"
	"bufio"
	"encoding/json"

	"../../store"
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
	Maintain   = "Maintain"
)

const gBCASTPORT = 6971
const gRINGPORT = 6972 
const gBroadcastIP = "255.255.255.255"

var gHEAD = false // Has to be changed
const gConnectAttempts = 5
const gBCASTPING = "RING EXISTS"
const gTIMEOUT = 2
const _timeoutMs = 3500

// Initializes the network if it's present. Establishes a new network if not
func Init() {
	if  !sendJoinMSG() {
		go listenjoin()
	}
	go maintainRing()
	go listenCallsIn()
	go listenCallsOut()
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

}

// Only runs if you are HEAD, listen for new machines broadcasting
// on the network using UDP. The new machine is added to the list of
// known machines. That list is propagted trpough the ring to update the ring
func listenjoin() {

	fmt.Println("Listening for join msgs")

	var msg string
	connRead := dialBroadcastUDP(gBCASTPORT)
	connWrite := dialBroadcastUDP(gRINGPORT)
	read_chn := make(chan string)
	recivedJoinMsg := false

	defer connRead.Close()
	defer connWrite.Close()


	for {
		time.Sleep(gTIMEOUT*time.Second)
		go readWithTimeout(connRead,read_chn)

		select {
		case msg = <-read_chn:
			recivedJoinMsg = true
			break
		case <-time.After(_timeoutMs * time.Millisecond):
			break
			}		
		if recivedJoinMsg {
			addrWrite, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", gBroadcastIP, gRINGPORT))
			connWrite.WriteTo([]byte(gBCASTPING), addrWrite)
			nodes,_ := json.Marshal(peers.GetAll())
			messages.SendMessage(NodeChange, nodes)
			peers.AddTail(msg)
			if peers.GetRelativeTo(peers.Tail,-1) == peers.GetRelativeTo(peers.Self, 0){
				messages.ConnectTo(peers.GetRelativeTo(peers.Self, 1))
			}
			if err != nil {
				fmt.Println("Failed to broadcast")
				fmt.Println(err)
			}
		}
	}
}

func maintainRing() {
	var nodesList []string
	messages.Start()
	for {
		fmt.Println("Maintaining ring...")
		fmt.Println(peers.GetAll())
		nodes := messages.Receive(NodeChange)
		json.Unmarshal(nodes, &nodesList)
		//fmt.Println(nodesList)
		peers.Set(nodesList)
		messages.SendMessage(Maintain, nodes)
	}
}

func listenCallsIn() {
	for {
		msg := messages.Receive(CallList)
		fmt.Println(string(msg))
		/*state := store.ElevatorState{}
    	gob.NewDecoder(msg).Decode(&state)

		store.RecieveElevState <- state
		*/
	}
}

func listenCallsOut() {
	for {
		states := <-store.SendElevState
		buf := &bytes.Buffer{}
		gob.NewEncoder(buf).Encode(states)
		bs := buf.Bytes()
		messages.SendMessage(CallList,bs)
	}
}

// send msg to 255
func sendJoinMSG() bool{
	// Connect til
	var msg string
	connWrite := dialBroadcastUDP(gBCASTPORT)
	connRead := dialBroadcastUDP(gRINGPORT)
	defer connRead.Close()
	defer connWrite.Close()
	read_chn := make(chan string)


	for i := 0; i < gConnectAttempts; i++ {
		time.Sleep(gTIMEOUT*time.Second)

		selfIP := peers.GetRelativeTo(peers.Self, 0)
		addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", gBroadcastIP, gBCASTPORT))
		connWrite.WriteTo([]byte(selfIP), addr)
		
		go readWithTimeout(connRead,read_chn)
		select {
			case msg = <-read_chn:
				break
			case <-time.After(_timeoutMs * time.Millisecond):
				fmt.Println("TImed out")
				break
		}		
		fmt.Printf("%s -- \n", msg)
		if msg == gBCASTPING {
			return true
		}
	}
	return false
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


func readWithTimeout(conn net.PacketConn, msg chan<- string) {
	buffer := make([]byte, 100) 
	n,_,_ := conn.ReadFrom(buffer[0:])
	msg <- string(buffer[:n])
}