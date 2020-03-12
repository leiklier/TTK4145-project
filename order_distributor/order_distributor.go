package order_distributor

import (
	"encoding/json"

	"../network/peers"
	"../network/ring"

	"../sync/elevators"
	"../sync/store"
)

const (
	State = "State"
	Call  = "Call"
)

func SendElevState(states elevators.Elevator_s) bool {
	stateBytes, _ := json.Marshal(states)
	return ring.BroadcastMessage(State, stateBytes)
}

func SendHallCall(ip string, hCall elevators.HallCall_s) bool {
	hCallBytes, err := json.Marshal(hCall)
	if err != nil {
		return false
	}
	return ring.SendToPeer(Call, ip, hCallBytes)
}

func ListenElevatorUpdate() {
	call_channel := ring.GetReceiver(Call)
	state_channel := ring.GetReceiver(State)

	callMap := make(map[string][]byte)
	state := elevators.Elevator_s{}
	hCall := elevators.HallCall_s{}
	selfIP := peers.GetRelativeTo(peers.Self, 0)

	for {
		select {
		case stateBytes := <-state_channel:
			json.Unmarshal(stateBytes, &state)
			store.UpdateState(state)
			break
		case call := <-call_channel:
			json.Unmarshal(call, &callMap)
			hCallBytes, found := callMap[selfIP]
			if found {
				json.Unmarshal(hCallBytes, &hCall)
				store.AddHallCall(selfIP, hCall)
			}
		case <-ring.NewNeighbourNode: // Do something with the newIP, eg add elevator state
			allStates := store.GetAll()
			for _, state := range allStates {
				SendElevState(state)
			}
			break

		}
	}
}
