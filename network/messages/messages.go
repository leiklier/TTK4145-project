package messages

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	"../peers"
	"../receivers"
)

// Enums
const (
	Broadcast = iota
	Ping
	PingAck
)

type Message struct {
	Purpose  string
	Type     int    // Broadcast or Ping or PingAck
	SenderIP string // Only necessary for Broadcast (we need to know where it started...)
	Data     []byte
}

// Variables
var gIsInitialized = false
var gPort = 6969

// Channels
var gServerIPChannel = make(chan string)
var gConnectedToServerChannel = make(chan string)
var gDisconnectedFromServerChannel = make(chan string)

var gSendForwardChannel = make(chan Message, 100)
var gSendBackwardChannel = make(chan Message, 100)

var gPingAckReceivedChannel = make(chan Message, 100)

func ConnectTo(IP string) error {
	initialize()
	gServerIPChannel <- IP
	select {
	case <-gConnectedToServerChannel:
		return nil
	case <-time.After(2 * time.Second):
		return errors.New("TIMED_OUT")
	}
}

func ServerDisconnected() string {
	return <-gDisconnectedFromServerChannel
}

// SendMessage takes a byte array and sends it
// to the node which it is connected to by ConnectTo
// purpose is used to filter the message on the receiving end
func SendMessage(purpose string, data []byte) {
	initialize()

	localIP := peers.GetRelativeTo(peers.Self, 0)
	message := Message{
		Purpose:  purpose,
		Type:     Broadcast,
		SenderIP: localIP,
		Data:     data,
	}
	gSendForwardChannel <- message
}

func Receive(purpose string) []byte {
	initialize()
	return <-receivers.GetChannel(purpose)
}

func Start() {
	initialize()
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
	var shouldDisconnectChannel = make(chan bool, 10)
	go handleOutboundConnection(serverIP, shouldDisconnectChannel)

	// We only want one active client at all times:
	for {
		serverIP := <-gServerIPChannel
		shouldDisconnectChannel <- true
		shouldDisconnectChannel = make(chan bool, 10)
		go handleOutboundConnection(serverIP, shouldDisconnectChannel)
	}
}

func handleOutboundConnection(serverIP string, shouldDisconnectChannel chan bool) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIP, gPort))
	if err != nil {
		fmt.Printf("TCP client connect error: %s", err)
		return
	}

	defer func() {
		defer fmt.Printf("I am disconnecting\n")
		defer conn.Close()
		gDisconnectedFromServerChannel <- serverIP
	}()
	gConnectedToServerChannel <- serverIP

	bytesReceivedChannel := make(chan []byte)
	readErrorsChannel := make(chan error)
	go readMessages(conn, bytesReceivedChannel, readErrorsChannel)

	for {
		select {
		case bytesReceived := <-bytesReceivedChannel:
			// We have received a message:
			if err != nil {
				fmt.Printf("TCP server receive error: %s", err)
				return
			}
			var messageReceived Message
			json.Unmarshal(bytesReceived, &messageReceived)

			if messageReceived.Type != PingAck {
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

		case <-readErrorsChannel:
			return
		}

	}
}

func server() {
	// Boot up TCP server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", gPort))
	if err != nil {
		fmt.Printf("TCP server listener error: %s", err)
	}

	// Listen to incoming connections
	var shouldDisconnectChannel = make(chan bool, 10)
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
		shouldDisconnectChannel = make(chan bool, 10)
		handleIncomingConnection(conn, shouldDisconnectChannel)

	}
}

func handleIncomingConnection(conn net.Conn, shouldDisconnectChannel chan bool) {
	defer conn.Close()

	bytesReceivedChannel := make(chan []byte)
	readErrorsChannel := make(chan error)
	go readMessages(conn, bytesReceivedChannel, readErrorsChannel)

	for {
		select {
		// We have received a message
		case bytesReceived := <-bytesReceivedChannel:
			var messageReceived Message
			json.Unmarshal(bytesReceived, &messageReceived)

			switch messageReceived.Type {
			case Broadcast:
				localIP := peers.GetRelativeTo(peers.Self, 0)

				if messageReceived.SenderIP != localIP {
					// We should forward the message to next node
					receivers.GetChannel(messageReceived.Purpose) <- messageReceived.Data
					gSendForwardChannel <- messageReceived
				}

				break
			case Ping:
				messageToSend := Message{
					Type: PingAck,
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

		case <-readErrorsChannel:
			return
		}

	}
}

func readMessages(conn net.Conn, receiveChannel chan []byte, errorChannel chan error) {
	for {
		bytesReceived, err := bufio.NewReader(conn).ReadBytes('\n')
		if err != nil {
			errorChannel <- err
			return
		}
		receiveChannel <- bytesReceived

	}
}
