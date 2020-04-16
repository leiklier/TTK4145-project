package order_distributor

import (
	"encoding/json"
	"fmt"
	"time"

	"../network/peers"
	"../network/ring"

	"../sync/elevators"
	"../sync/store"
)

const (
	State = "State"
	Call  = "Call"
)

var SendStateUpdate = make(chan bool)
var selfIP string

func Init() {
	selfIP = peers.GetRelativeTo(peers.Self, 0)

	go SendUpdate()
	go ListenElevatorUpdate()
	go RemovedPeerListener()
}

func SendElevState(state elevators.Elevator_s) bool {
	stateBytes, _ := state.Marshal()
	return ring.BroadcastMessage(State, stateBytes)
}

func SendHallCall(ip string, hCall elevators.HallCall_s) bool { // Not tested
	hCallBytes, err := json.Marshal(hCall)
	if err != nil {
		return false
	}
	return ring.SendToPeer(Call, ip, hCallBytes)
}

func SendUpdate() {

	for {
		select {
		case <-SendStateUpdate:
			state, _ := store.GetElevator(selfIP)
			SendElevState(state)
			break

		case <-time.After(10 * time.Second):
			state, _ := store.GetElevator(selfIP)
			SendElevState(state)
			break
		}
	}
}

func RemovedPeerListener() {
	for {
		select {
		case disconectedPeer := <-ring.DisconnectedPeer:
			hall_calls, _ := store.GetAllHallCalls(disconectedPeer)
			store.Remove(disconectedPeer)
			for _, hc := range hall_calls {
				mostSuitedIP := store.MostSuitedElevator(hc.Floor, hc.Direction)
				store.AddHallCall(selfIP, hc)
				SendHallCall(mostSuitedIP, hc)
			}
		}
	}
}

func ListenElevatorUpdate() {
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
			// hCallBytes, found := callMap[selfIP]
			for ip, hCallBytes := range callMap {
				json.Unmarshal(hCallBytes, &hCall)
				fmt.Println("Received most suited: ", ip)
				fmt.Println("Hallcall received: ", hCall)
				store.AddHallCall(ip, hCall)
			}

		}
	}
}
