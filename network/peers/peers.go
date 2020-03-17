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
var localIP string
var err error // Fucking go

// Set takes an array of IP addresses, DELETES all existing
// peers and adds the new IP addresses instead, in the same
// order as they appear in the IPs array.
func Set(IPs []string) {
	controlSignal := ControlSignal{
		Command: Replace,
		Payload: IPs,
	}
	controlChannel <- controlSignal
}

func BecomeHead() {
	controlSignal := ControlSignal{
		Command: Head,
	}
	controlChannel <- controlSignal
}

// Remove deletes the peer with a certain IP
func Remove(IP string) {
	controlSignal := ControlSignal{
		Command: Delete,
		Payload: []string{IP},
	}
	controlChannel <- controlSignal
}

// AddTail takes an IP address in the form of a string,
// and adds it at the end of the list of peers, thus
// creating a new tail. It returns nothing
func AddTail(IP string) bool {
	if stringInSlice(IP, GetAll()) {
		return false
	}
	controlSignal := ControlSignal{
		Command: Append,
		Payload: []string{IP},
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

func Init(connectPort string) error {
	localIP, err = getLocalIP()
	if err != nil {
		return err
	}
	localIP = localIP + ":" + connectPort
	fmt.Println(localIP)
	go peersServer()
	return nil
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
			break

		case Replace:
			peers = controlSignal.Payload

			break

		case Head:
			var rotation int
			var val string
			var newPeers []string

			for rotation, val = range peers {
				if val == localIP {
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
			for i, peer := range peers {
				if peer == peerToRemove {
					copy(peers[i:], peers[i+1:]) // Shift peers[i+1:] left one index.
					peers[len(peers)-1] = ""     // Erase last element (write zero value).
					peers = peers[:len(peers)-1] // Truncate slice.
					break
				}
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
	return GetRelativeTo(Head, 0) == GetRelativeTo(Tail, 0)
}

func IsHead() bool {
	return GetRelativeTo(Head, 0) == GetRelativeTo(Self, 0)
}

func IsNextTail() bool {
	return GetRelativeTo(Self, 1) == GetRelativeTo(Tail, 0)
}

func GetNextPeer() string {
	return GetRelativeTo(Self, 1)
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

func difference(slice1 []string, slice2 []string) (string, bool) {
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
	// Swap the slices, only if it was the first loop
	slice1, slice2 = slice2, slice1
	return "", false
}
