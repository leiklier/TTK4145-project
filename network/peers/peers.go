package peers

import (
	"net"
	"strings"
)

// Enums
const (
	Get = iota
	Append
	Self
	Head
	Tail
	Added   = "Added"
	Removed = "Removed"
)

type ControlSignal struct {
	Command         int // Get or Append
	Payload         []string
	ResponseChannel chan []string
}

var controlChannel = make(chan ControlSignal)

type ChangeEvent struct {
	Event string // Added or Removed
	Peer  string
}

var changeChannel = make(chan ChangeEvent, 100)

// Local variables
var isInitialized = false
var localIP string

func AddTail(IP string) {
	initialize()
	controlSignal := ControlSignal{
		Command: Append,
		Payload: []string{IP},
	}
	controlChannel <- controlSignal
}

func PollUpdate() ChangeEvent {
	return <-changeChannel
}

func GetAll() []string {
	initialize()
	controlSignal := ControlSignal{
		Command:         Get,
		ResponseChannel: make(chan []string),
	}
	controlChannel <- controlSignal
	peers := <-controlSignal.ResponseChannel
	return peers
}

func GetRelativeTo(role int, offset int) string {
	initialize()
	peers := GetAll()

	var indexOfRole int
	if role == Head {
		indexOfRole = 0
	} else if role == Tail {
		indexOfRole = len(peers) - 1
	} else if role == Self {
		for index, peer := range peers {
			if peer == localIP {
				indexOfRole = index
				break
			}
		}
	}

	indexWithOffset := indexOfRole + offset
	indexWithOffset = indexWithOffset % len(peers)
	if indexWithOffset < 0 {
		indexWithOffset += len(peers)
	}

	return peers[indexWithOffset]
}

func initialize() {
	if isInitialized {
		return
	}
	isInitialized = true
	localIP, _ = getLocalIP()
	go peersServer()
}

func peersServer() {
	peers := make([]string, 1)
	peers[0] = localIP
	for {
		controlSignal := <-controlChannel
		switch controlSignal.Command {
		case Get:
			controlSignal.ResponseChannel <- peers
			break
		case Append:
			peers = append(peers, controlSignal.Payload...)
			for _, newPeer := range controlSignal.Payload {
				changeEvent := ChangeEvent{
					Event: Added,
					Peer:  newPeer,
				}
				changeChannel <- changeEvent
			}
			break
		}
	}
}

func getLocalIP() (string, error) {
	conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
	if err != nil {
		return "", err
	}
	defer conn.Close()
	return strings.Split(conn.LocalAddr().String(), ":")[0], nil
}
