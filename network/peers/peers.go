package peers

import (
	"fmt"
	"net"
	"strings"
)

// Enums
const (
	Get = iota
	Replace
	Delete
	Append
	Self
	Head
	Tail
	Added    = "Added"    // Used when a peer was added
	Removed  = "Removed"  // Used when a peer was removed
	Replaced = "Replaced" // Used when all peers are replaced by new
)

type ControlSignal struct {
	Command         int // Get or Append
	Payload         []string
	ResponseChannel chan []string
}

var controlChannel = make(chan ControlSignal, 100)

// Local variables
var localHostname string
var err error // Fucking go

// Set takes an array of hostnames, DELETES all existing
// peers and adds the new Hostname addresses instead, in the same
// order as they appear in the IPs array.
func Set(hostnames []string) {
	controlSignal := ControlSignal{
		Command: Replace,
		Payload: hostnames,
	}
	controlChannel <- controlSignal
}

func BecomeHead() {
	controlSignal := ControlSignal{
		Command: Head,
	}
	controlChannel <- controlSignal
}

// Remove deletes the peer with a certain hostname
func Remove(hostname string) {
	controlSignal := ControlSignal{
		Command: Delete,
		Payload: []string{hostname},
	}
	controlChannel <- controlSignal
}

// AddTail takes an IP address in the form of a string,
// and adds it at the end of the list of peers, thus
// creating a new tail. It returns nothing
func AddTail(hostname string) bool {
	if stringInSlice(hostname, GetAll()) {
		return false
	}
	controlSignal := ControlSignal{
		Command: Append,
		Payload: []string{hostname},
	}

	controlChannel <- controlSignal
	return true
}

// GetAll returns the array of peers in the correct order
// so the first element is HEAD and the last element is Tail
func GetAll() []string {
	controlSignal := ControlSignal{
		Command:         Get,
		ResponseChannel: make(chan []string),
	}
	controlChannel <- controlSignal
	peers := <-controlSignal.ResponseChannel
	return peers
}

// GetRelativeTo takes either peers.Head, peers.Tail or peers.Self
// as first argument. Then it returns the ip of that peer if offset=0.
// If offset is not 0, then it adds that such that i.e. with role=peers.Self
// and offset=1, it returns the peer AFTER Self, and with offset=-1 it returns
// the peer BEFORE Self.
func GetRelativeTo(role int, offset int) string {
	peers := GetAll()

	var indexOfRole int
	if role == Head {
		indexOfRole = 0
	} else if role == Tail {
		indexOfRole = len(peers) - 1
	} else if role == Self {
		for index, peer := range peers {
			if peer == localHostname {
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

func Init(connectPort string) error {
	localIP, err := getLocalIP()
	if err != nil {
		return err
	}
	localHostname = localIP + ":" + connectPort
	fmt.Println(localHostname)
	go peersServer()
	return nil
}

func peersServer() {
	peers := make([]string, 1)
	peers[0] = localHostname
	for {
		controlSignal := <-controlChannel
		switch controlSignal.Command {
		case Get:
			controlSignal.ResponseChannel <- peers
			break

		case Append:
			peers = append(peers, controlSignal.Payload...)
			break

		case Replace:
			peers = controlSignal.Payload
			selfInPeers := false
			for _, peer := range peers {
				if peer == localHostname {
					selfInPeers = true
				}
			}

			if selfInPeers == false {
				peers = append(peers, localHostname)
			}
			break

		case Head:
			var rotation int
			var val string
			var newPeers []string

			for rotation, val = range peers {
				if val == localHostname {
					break
				}
			}
			size := len(peers)
			for i := 0; i < rotation; i++ {
				newPeers = peers[1:size]
				newPeers = append(newPeers, peers[0])
				peers = newPeers
			}
			break

		case Delete:
			peerToRemove := controlSignal.Payload[0]
			if peerToRemove != localHostname { // Don't delete yourself

				for i, peer := range peers {
					if peer == peerToRemove {
						copy(peers[i:], peers[i+1:]) // Shift peers[i+1:] left one index.
						peers[len(peers)-1] = ""     // Erase last element (write zero value).
						peers = peers[:len(peers)-1] // Truncate slice.
						break
					}
				}
				break
			}
		}
	}
}

func IsEqualTo(peersToCompare []string) bool {
	currentPeers := GetAll()

	// If one is nil, the other must also be nil.
	if peersToCompare == nil {
		return false
	}

	if len(peersToCompare) != len(currentPeers) {
		return false
	}

	for i := range peersToCompare {
		if peersToCompare[i] != currentPeers[i] {
			return false
		}
	}

	return true
}

func IsAlone() bool {
	if len(GetAll()) == 1 {
		return true
	}
	return false
}

func IsHead() bool {
	return GetRelativeTo(Head, 0) == GetRelativeTo(Self, 0)
}

func IsNextTail() bool {
	return GetRelativeTo(Self, 1) == GetRelativeTo(Tail, 0)
}

func getLocalIP() (string, error) {
	conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
	if err != nil {
		return "", err
	}
	defer conn.Close()
	return strings.Split(conn.LocalAddr().String(), ":")[0], nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
