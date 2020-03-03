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

func SendHallCall(ip string, ip string, hCall store.HallCall) bool {
	hCallBytes, _ := json.Marshal(hCall)
	return ring.SendDM(Call, ip, hCallBytes)
}

func ReceiveHallCall() store.HallCall {
	hCall := store.HallCall
	hCallBytes := ring.Receive(Call)

	json.Unmarshal(hCallBytes, &hCall)
	return hCall
}
