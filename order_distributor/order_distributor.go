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

func Init() {
	go SendUpdate()
	go ListenElevatorUpdate()
}

func SendElevState(state elevators.Elevator_s) bool {
	stateBytes, _ := state.Marshal()
	return ring.BroadcastMessage(State, stateBytes)
}

func SendHallCall(ip string, hCall elevators.HallCall_s) bool { // Not tested
	fmt.Printf("Sending to %s\n", ip)
	hCallBytes, err := json.Marshal(hCall)
	if err != nil {
		return false
	}
	return ring.SendToPeer(Call, ip, hCallBytes)
}

func SendUpdate() {
	selfIP := peers.GetRelativeTo(peers.Self, 0)

	for {
		select {
		case <-time.After(3 * time.Second):
			state, _ := store.GetElevator(selfIP)
			SendElevState(state)
			break
		}
	}
}

func ListenElevatorUpdate() {
	call_channel := ring.GetReceiver(Call)
	state_channel := ring.GetReceiver(State)

	callMap := make(map[string][]byte)
	hCall := elevators.HallCall_s{}
	selfIP := peers.GetRelativeTo(peers.Self, 0)

	for {
		select {
		case stateBytes := <-state_channel:
			state := elevators.UnmarshalElevatorState(stateBytes)
			store.UpdateState(state)
			break
		case call := <-call_channel:
			json.Unmarshal(call, &callMap)
			hCallBytes, found := callMap[selfIP]
			fmt.Printf("Received from %s\n", callMap)
			if found {
				json.Unmarshal(hCallBytes, &hCall)
				store.AddHallCall(selfIP, hCall)
			}

		}
	}
}
