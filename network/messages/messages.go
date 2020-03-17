package messages

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"../receivers"

	"../peers"
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
var gInnPort string
var gOutPort int

//Public channels
var DisconnectedFromServerChannel = make(chan string)

// Channels
var gServerIPChannel = make(chan string)
var gConnectedToServerChannel = make(chan string)

// TODO: Make these channel names more meaningful
var gSendForwardChannel = make(chan Message, 100)
var gSendBackwardChannel = make(chan Message, 100)

func Init(innPort string, outPort string) {
	gInnPort = innPort
	gOutPort, _ = strconv.Atoi(outPort)
	go client()
	go server()
}

func ConnectTo(IP string) error {
	gServerIPChannel <- IP
	select {
	case <-gConnectedToServerChannel:
		return nil
	case <-time.After(2 * time.Second):
		return errors.New("TIMED_OUT")
	}
}

// SendMessage takes a byte array and sends it
// to the node which it is connected to by ConnectTo
// purpose is used to filter the message on the receiving end
func SendMessage(purpose string, data []byte) bool {
	if peers.IsAlone() {
		return false
	}
	localIP := peers.GetRelativeTo(peers.Self, 0)
	message := Message{
		Purpose:  purpose,
		Type:     Broadcast,
		SenderIP: localIP,
		Data:     data,
	}
	gSendForwardChannel <- message
	return true
}

func GetReceiver(purpose string) chan []byte {
	return receivers.GetChannel(purpose)
}

func client() {
	serverIP := <-gServerIPChannel

	self := peers.GetRelativeTo(peers.Self, 0)
	splittedMsg := strings.SplitN(self, ":", 2)
	dialer := &net.Dialer{
		LocalAddr: &net.TCPAddr{
			IP:   net.ParseIP(splittedMsg[0]),
			Port: gOutPort,
		},
	}
	var shouldDisconnectChannel = make(chan bool, 10)
	go handleOutboundConnection(serverIP, dialer, shouldDisconnectChannel)

	// We only want one active client at all times:
	for {
		serverIP := <-gServerIPChannel
		shouldDisconnectChannel <- true
		shouldDisconnectChannel = make(chan bool, 10)
		go handleOutboundConnection(serverIP, dialer, shouldDisconnectChannel)
	}
}

func handleOutboundConnection(server string, dialer *net.Dialer, shouldDisconnectChannel chan bool) {
	conn, err := dialer.Dial("tcp", server)
	if err != nil {
		fmt.Printf("TCP client connect error: %s", err)
		return
	}

	defer func() {
		fmt.Printf("messages: lost connection to server with IP %s\n", server)
		conn.Close()
		if err != nil {
			DisconnectedFromServerChannel <- server
		}
	}()

	gConnectedToServerChannel <- server

	shouldSendPingTicker := time.NewTicker(500 * time.Millisecond)

	pingAckReceivedChannel := make(chan Message, 100)
	connErrorChannel := make(chan error)

	// Read new messages from conn and send them on pingAckReceivedChannel.
	// Errors are sent back on connErrorChannel.
	go receiveMessages(conn, pingAckReceivedChannel, connErrorChannel)

	// Send messages that are passed to gSendForwardChannel
	// Errors are sent back on connErrorChannel.
	go sendMessages(conn, gSendForwardChannel, connErrorChannel)

	for {
		select {
		case <-shouldSendPingTicker.C:
			// Send a ping message at regular intervals to check that
			// the connection is still alive
			messageToSend := Message{
				Type: Ping,
			}
			gSendForwardChannel <- messageToSend
			select {
			case <-pingAckReceivedChannel:
				// fmt.Println("Ping ping")
				// We received a PingAck, so everything works fine
				break
			case <-time.After(1 * time.Second):
				// Cannot retrieve PingAck, so the connection is
				// not working properly

				err = errors.New("ERR_SERVER_DISCONNECTED")
				return
			}
			break

		case <-shouldDisconnectChannel:
			fmt.Printf("Disconnecting from: %s\n", server)
			return

		case <-connErrorChannel:
			err = errors.New("ERR_SERVER_DISCONNECTED")
			return
		}

	}
}

func server() {
	// Boot up TCP server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", gInnPort))
	if err != nil {
		fmt.Printf("TCP server listener error: %s", err) // TODO: Maybe do something with this error
	}

	// Listen to incoming connections
	var shouldDisconnectChannel = make(chan bool, 10)
	for {
		// Accept a new connection
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("TCP server accept error: %s", err) // TODO: and this one
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

	messageReceivedChannel := make(chan Message, 100)
	connErrorChannel := make(chan error)

	// Read new messages from conn and send them on pingAckReceivedChannel.
	// Errors are sent back on connErrorChannel.
	go receiveMessages(conn, messageReceivedChannel, connErrorChannel)

	// Send messages that are passed to gSendForwardChannel
	// Errors are sent back on connErrorChannel.
	go sendMessages(conn, gSendBackwardChannel, connErrorChannel)

	for {
		select {
		// We have received a message
		case messageReceived := <-messageReceivedChannel:

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

		case <-shouldDisconnectChannel:
			return

		case <-connErrorChannel:
			return
		}

	}
}

func receiveMessages(conn net.Conn, receiveChannel chan Message, errorChannel chan error) {
	for {
		bytesReceived, err := bufio.NewReader(conn).ReadBytes('\n')
		if err != nil {
			errorChannel <- err
			return
		}

		var messageReceived Message
		json.Unmarshal(bytesReceived, &messageReceived)

		receiveChannel <- messageReceived

	}
}

func sendMessages(conn net.Conn, messageToSendChannel chan Message, errorChannel chan error) {
	for {
		messageToSend := <-messageToSendChannel
		serializedMessage, _ := json.Marshal(messageToSend)

		_, err := fmt.Fprintf(conn, string(serializedMessage)+"\n\000")

		if err != nil {
			// We need to retransmit the message, to pass it back to the channel.
			// However, the connection is not working so disconnect.
			messageToSendChannel <- messageToSend
			errorChannel <- err
			return
		}
	}
}
