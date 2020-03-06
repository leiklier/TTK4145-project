package order_distributor

import (
	"encoding/json"

	"../network/ring"
	"../store"
)

func SendElevState(state store.ElevatorState) bool {
	stateBytes := json.Marshal(state)
	return ring.SendMessage(StateChange, stateBytes)
}

func ReceiveElevState() store.ElevatorState {
	state := store.ElevatorState
	stateBytes := ring.Receive(StateChange)
	json.Unmarshal(stateBytes, &state)
	return state, true
}
