package order_distributor

import (
	"encoding/json"
	"time"

	"../network/peers"
	"../network/ring"

	"../sync/elevators"
	"../sync/store"
)

//Message purposes
const (
	State = "State"
	Call  = "Call"
)
const updateInterval = 5

var ShouldSendStateUpdate = make(chan bool, 10)
var selfIP string

func Init() {
	selfIP = peers.GetRelativeTo(peers.Self, 0)

	go sendUpdate()
	go listenElevatorUpdate()
	go removedPeerListener()
}

func SendElevState(state elevators.Elevator_s) bool {
	stateBytes, _ := state.Marshal()
	return ring.BroadcastMessage(State, stateBytes)
}

func SendHallCall(ip string, hCall elevators.HallCall_s) bool {
	hallCallBytes, err := json.Marshal(hCall)
	if err != nil {
		return false
	}
	return ring.SendToPeer(Call, ip, hallCallBytes)
}

func sendUpdate() {
	for {
		select {
		case <-ShouldSendStateUpdate:
			state, _ := store.GetElevator(selfIP)
			SendElevState(state)
			break
		case <-time.After(updateInterval * time.Second):
			state, _ := store.GetElevator(selfIP)
			SendElevState(state)
			break
		}
	}
}

// Distributes the HallCalls assigned to a elevator that has disconnected
func removedPeerListener() {
	for {
		select {
		case disconectedPeer := <-ring.DisconnectedPeer:
			allHallCalls, _ := store.GetAllHallCalls(disconectedPeer)
			store.Remove(disconectedPeer)
			for _, hallCall := range allHallCalls {
				mostSuitedIP := store.MostSuitedElevator(hallCall.Floor, hallCall.Direction)
				store.AddHallCall(selfIP, hallCall)
				SendHallCall(mostSuitedIP, hallCall)
			}
		}
	}
}

// Listens for updates about other elevators and updates store accordingly
func listenElevatorUpdate() {
	call_channel := ring.GetReceiver(Call)
	state_channel := ring.GetReceiver(State)

	for {
		select {
		case stateBytes := <-state_channel:
			state := elevators.UnmarshalElevatorState(stateBytes)
			store.Replace(state)
			break

		case call := <-call_channel:
			callMap := make(map[string][]byte)
			hCall := elevators.HallCall_s{}
			json.Unmarshal(call, &callMap)
			for ip, hCallBytes := range callMap {
				json.Unmarshal(hCallBytes, &hCall)
				store.AddHallCall(ip, hCall)
			}

		}
	}
}
