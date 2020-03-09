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

func SendElevState(states elevators.Elevator_s) bool {
	stateBytes := json.Marshal(states)
	return ring.BroadcastMessage(State, stateBytes)
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
	states := elevators.Elevator_s
	hCall := elevators.HallCall

	for select {
		case stateBytes := <- state_channel:
			json.Unmarshal(stateBytes, &states)
			store.Update(states)
			break
		case call := <- call_channel:
			json.Unmarshal(call, &dataMap)
			hCallBytes, found := dataMap[selfIP]
			if found {
				json.Unmarshal(hCallBytes, &hCall)
				store.AddCall(hCall)
			}
		case newIP := <- ring.NewNeighbourNode:
			allStates := store.GetAll()
			for state in allStates{
				ring.SendElevState(state)
			}
			break

	}
}
