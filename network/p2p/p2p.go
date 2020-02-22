package p2p

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

type Message struct {
	Purpose string
	Data    string
}

type Receiver struct {
	Message         Message
	ResponseChannel chan Message
}

var _isStarted = false
var _port = 0
var _receiveChannel = make(chan Receiver, 100)

func Start(port int) {
	if _isStarted {
		return
	}
	_isStarted = true
	_port = port

	go server()
}

func server() {
	// boot up tcp server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", _port))
	if err != nil {
		log.Fatal("tcp server listener error:", err)
	}

	// Listen to all incoming connections
	for {
		// accept new connection
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("tcp server accept error", err)
		}

		// spawn off goroutine to able to accept new connections
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// read buffer from client after enter is hit
	bufferBytes, err := bufio.NewReader(conn).ReadBytes('\n')
	var message Message
	json.Unmarshal(bufferBytes, &message)

	if err != nil {
		log.Println("client left..")
		conn.Close()
	}

	receiver := Receiver{
		Message:         message,
		ResponseChannel: make(chan Message),
	}
	_receiveChannel <- receiver

	response := <-receiver.ResponseChannel
	serializedResponse, _ := json.Marshal(response)
	conn.Write([]byte(serializedResponse))
}

func SendTo(ip string, message Message) Message {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, _port))
	if err != nil {
		log.Fatal(err)
		return Message{}
	}

	serializedMessage, _ := json.Marshal(message)

	fmt.Fprintf(conn, string(serializedMessage)+"\n\000")
	serializedResponse, _ := bufio.NewReader(conn).ReadBytes('\n')
	var response Message
	json.Unmarshal(serializedResponse, &response)

	return response
}

func Receive() (Message, func(response Message)) {
	receiver := <-_receiveChannel

	return receiver.Message, func(response Message) {
		receiver.ResponseChannel <- response
	}
}
