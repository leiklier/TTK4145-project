package messages

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"

	"../peers"
)

// Enums
const (
	Broadcast = iota
	Ping
	PingAck
	Direct
)

type Message struct {
	Purpose  int    // Broadcast or Ping or PingAck og Forward or Backward
	SenderIP string // Only necessary for Broadcast (we need to know where it started...)
	Data     []byte
}

// Variables
var gIsInitialized = false
var gPort = 69420

// Channels
var gServerIPChannel = make(chan string)

var gSendForwardChannel = make(chan Message, 100)
var gSendBackwardChannel = make(chan Message, 100)

var gDirectReceivedChannel = make(chan Message, 100)
var gBroadcastReceivedChannel = make(chan Message, 100)
var gPingAckReceivedChannel = make(chan Message, 100)

func ConnectTo(IP string) {
	initialize()
	gServerIPChannel <- IP
}

func Send(message Message) {
	initialize()
	if message.Purpose == PingAck {
		gSendBackwardChannel <- message
		return
	}
	gSendForwardChannel <- message
}

func Receive(purpose int) {
	switch purpose {
	case Direct:
		return <-gDirectReceivedChannel
	case Broadcast:
		return <-gBroadcastReceivedChannel
	case PingAck:
		return <-gPingAckReceivedChannel
	}
}

func initialize() {
	if gIsInitialized {
		return
	}
	gIsInitialized = true
	go client()
	go server()
}

func client() {
	serverIP := <-gServerIPChannel
	var shouldDisconnectChannel = make(chan bool)
	go handleOutboundConnection(serverIP, shouldDisconnectChannel)

	// We only want one active client at all times:
	for {
		serverIP := <-gServerIPChannel
		shouldDisconnectChannel <- true
		shouldDisconnectChannel = make(chan bool)
		go handleOutboundConnection(serverIP, shouldDisconnectChannel)
	}
}

func handleOutboundConnection(serverIP, shouldDisconnectChannel chan bool) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIP, gPort))
	if err != nil {
		fmt.Printf("TCP client connect error: %s", err)
		return
	}
	defer conn.Close()

	for {
		select {
		case receiveBuffer, err := bufio.NewReader(conn).ReadBytes('\n'):
			// We have received a message:
			if err != nil {
				fmt.Printf("TCP server receive error: %s", err)
				conn.Close()
			}
			var messageReceived Message
			json.Unmarshal(receiveBuffer, &messageReceived)

			if messageReceived.Purpose != PingAck {
				break
			}

			gPingAckReceivedChannel <- messageReceived
			break

		case messageToSend := <-gSendForwardChannel:
			// We want to transmit a message forwards:
			serializedMessage, _ := json.Marshal(messageToSend)
			fmt.Fprintf(conn, string(serializedMessage)+"\n\000")
			break

		case <-shouldDisconnectChannel:
			return
		}

	}
}

func server() {
	// Boot up TCP server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", _port))
	if err != nil {
		fmt.Printf("TCP server listener error: %s", err)
	}

	// Listen to incoming connections
	var shouldDisconnectChannel = make(chan bool)
	for {
		// Accept a new connection
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("TCP server accept error: %s", err)
			break
		}
		// A new client connected to us, so disconnect to the one
		// already connected because we only accept one connection
		// at all times
		shouldDisconnectChannel <- true
		shouldDisconnectChannel = make(chan bool)
		handleIncomingConnection(conn, shouldDisconnectChannel)

	}
}

func handleIncomingConnection(conn net.Conn, shouldDisconnectChannel chan bool) {
	defer conn.Close()

	for {
		select {
		// We have received a message
		case receiveBuffer, err := bufio.NewReader(conn).ReadBytes('\n'):
			if err != nil {
				fmt.Printf("TCP server receive error: %s", err)
				conn.Close()
			}
			var messageReceived Message
			json.Unmarshal(receiveBuffer, &messageReceived)

			switch messageReceived.Purpose {
			case Direct:
				gDirectReceivedChannel <- messageReceived
				break
			case Broadcast:
				localIP := peers.GetRelativeTo(peers.Self, 0)
				gBroadcastReceivedChannel <- messageReceived

				if messageReceived.SenderIP != localIP {
					// We should forward the message to next node
					Send(messageReceived)
				}

				break
			case Ping:
				messageToSend := Message{
					Purpose: PingAck,
				}
				gSendBackwardChannel <- messageToSend
				break
			}
			break

		// We want to transmit a message backwards:
		case messageToSend := <-gSendBackwardChannel:
			serializedMessage, _ := json.Marshal(messageToSend)
			fmt.Fprintf(conn, string(serializedMessage)+"\n\000")
			break

		case <-shouldDisconnectChannel:
			return
		}

	}
}
