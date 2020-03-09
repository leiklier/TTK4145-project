package order_distributor

import (
	"encoding/json"

	"../network/ring"
	"../sync/store"
	"../sync/elevators"
	"../sync/watchdog"
)

const (
	State = "State"
	Call = "Call"
)

func SendElevState(state store.ElevatorState) bool {
	stateBytes := json.Marshal(state)
	return ring.BroadcastMessage(StateChange, stateBytes)
}

func SendHallCall(ip string, hCall elevators.HallCall) bool {
	hCallBytes, err := json.Marshal(hCall) 
	if err != nil {
		return false
	}
	return ring.SendToPeer(Call, ip, hCallBytes)
}

func ListenElevatorUpdate() {
	call_channel = ring.GetReceiver(Call)
	state_channel = ring.GetReceiver(State)

	callMap := make(map[string][]byte)
	hCall := elevators.HallCall

	for select {
		case state := <- state_channel:
			store.Update(state)
			// watchdog.AddState(state)
			break
		case call := <- call_channel:
			json.Unmarshal(call, &dataMap)
			hCallBytes, found := dataMap[selfIP]
			if found {
				json.Unmarshal(hCallBytes, &hCall)
				store.AddCall(hCall)
			}
			break
	}
}
