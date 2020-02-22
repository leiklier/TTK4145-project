package messages

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
)

// Enums
const (
	Broadcast = iota
	Ping
	PingAck
	Forward
	Backward
)

type Message struct {
	Purpose    int // Broadcast or Ping or PingAck og Forward or Backward
	Data       []byte
}

// Variables
var gIsInitialized = false
var gPort = 69420

// Channels
var gServerIPChannel = make(chan string)

var gSendBackwardChannel = make(chan Message, 100)
var gSendForwardChannel = make(chan Message, 100)

var gForwardReceivedChannel = make(chan Message, 100)
var gBackwardReceivedChannel = make(chan Message, 100)
var gBroadcastReceivedChannel = make(chan Message, 100)
var gPingAckReceivedChannel = make(chan Message, 100)

func ConnectTo(IP string) {
	initialize()
	gServerIPChannel <- IP
}

func Send(message Message) {

}

func initialize() {
	if gIsInitialized {
		return
	}
	gIsInitialized = true
	go server()
}

func client(serverIP) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIP, gPort))
	if err != nil {
		fmt.Printf("TCP client connect error: %s", err)
		return
	}

	for {
		select {
		case receiveBuffer, err := bufio.NewReader(conn).ReadBytes('\n'):
			if err != nil {
				fmt.Printf("TCP server receive error: %s", err)
				conn.Close()
			}
			var messageReceived Message
			json.Unmarshal(receiveBuffer, &messageReceived)

			switch(messageReceived.Purpose) {
			case PingAck:
				gPingAckReceivedChannel <- messageReceived
				break
			case Backward
			}
			break
		}
	}
}

func server() {
	// Boot up tcp server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", _port))
	if err != nil {
		fmt.Printf("TCP server listener error: %s", err)
	}

	// Listen to all incoming connections
	for {
		// Accept a new connection
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("TCP server accept error: %s", err)
		}

		// Spawn off goroutine to be able to accept new connections
		// while this one is handled
		go handleIncomingConnection(conn)
	}
}

func handleIncomingConnection(conn net.Conn) {
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

			switch(messageReceived.Purpose) {
			case Ping:
				messageToSend := Message {
					Purpose: PingAck,
				}
				gSendBackwardChannel <- messageToSend
				break

			case Forward:
				gForwardReceivedChannel <- messageReceived
				break
			}
			break

		// We want to transmit a message backwards:
		case messageToSend:= <-gSendBackwardChannel:
			serializedMessage, _ := json.Marshal(messageToSend)
			fmt.Fprintf(conn, string(serializedMessage)+"\n\000")
			break
		}
	}
}
