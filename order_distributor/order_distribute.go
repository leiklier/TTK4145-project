package order_distributor

import (
	"encoding/json"

	"../network/ring"
	"../store"
)

func SendElevState(state store.ElevatorState) {
	Init()
	stateBytes := json.Marshal(state)
	messages.SendMessage(StateChange, stateBytes)
}

func ReceiveElevState() store.ElevatorState {
	ring.Init()
	state := store.ElevatorState
	stateBytes := messages.Receive(StateChange)
	json.Unmarshal(stateBytes, &state)
	return state
}

func SendHallCall(ip string, hCall store.HallCall) {
	ring.Init()
	hCallMap := make(map[string]store.HallCall)
	hCallBytes := json.Marshal(hCallMap)
	messages.SendMessage(Call, hCallBytes)
}

func ReceiveHallCall() store.HallCall {
	Init()
	for {
		hCallMap := make(map[string]store.HallCall)
		hCallBytes := messages.Receive(Call)

		json.Unmarshal(hCallBytes, &hCallMap)
		selfIP := peers.GetRelativeTo(peers.Self, 0)
		hcall, found := hCallMap[selfIP]
		if found {
			return hcall
		}
	}
}
