package elevators

import (
	"encoding/json"
	"fmt"
)

type HallCall_s struct {
	Floor     int
	Direction Direction_e
}

type Direction_e int

const (
	DirectionUp   Direction_e = 1
	DirectionDown Direction_e = -1
	DirectionBoth Direction_e = 2 // Only used for HallCalls
	DirectionIdle Direction_e = 0
)

type Elevator_s struct {
	hostname        string
	currentFloor    int
	NumFloors       int
	prevFloor       int
	directionMoving Direction_e
	hallCalls       []HallCall_s
	cabCalls        []bool
	isWorking       bool
}

func New(peerHostname string, numFloors int, currentFloor int) Elevator_s {
	var hallCalls = make([]HallCall_s, numFloors)
	for i, _ := range hallCalls {
		hc := HallCall_s{Floor: i, Direction: DirectionIdle}
		hallCalls[i] = hc
	}

	elevator := Elevator_s{
		hostname:        peerHostname,
		currentFloor:    currentFloor,
		prevFloor:       currentFloor, // initialize with same floor
		NumFloors:       numFloors,
		directionMoving: DirectionIdle,
		hallCalls:       hallCalls,
		cabCalls:        make([]bool, numFloors),
	}

	return elevator
}

func (e Elevator_s) GetHostname() string {
	return e.hostname
}

func (e Elevator_s) GetCurrentFloor() int {
	return e.currentFloor
}

func (e *Elevator_s) SetCurrentFloor(newCurrentFloor int) {
	fmt.Print("Elevator set: ")
	fmt.Println(newCurrentFloor)
	e.currentFloor = newCurrentFloor
}

func (e Elevator_s) GetDirectionMoving() Direction_e {
	return e.directionMoving
}
func (e Elevator_s) GetPreviousFloor() int {
	return e.prevFloor
}

func (e *Elevator_s) SetDirectionMoving(newDirection Direction_e) {
	e.directionMoving = newDirection
}

func (e Elevator_s) GetAllHallCalls() []HallCall_s {
	return e.hallCalls
}

func (e *Elevator_s) AddHallCall(hallCall HallCall_s) {
	if hallCall.Direction == DirectionUp && e.hallCalls[hallCall.Floor].Direction == DirectionDown {
		e.hallCalls[hallCall.Floor].Direction = DirectionBoth
	} else if hallCall.Direction == DirectionDown && e.hallCalls[hallCall.Floor].Direction == DirectionUp {
		e.hallCalls[hallCall.Floor].Direction = DirectionBoth
	} else {
		e.hallCalls[hallCall.Floor].Direction = hallCall.Direction
	}
}

func (e *Elevator_s) RemoveHallCalls(floor int) {
	e.hallCalls[floor].Direction = DirectionIdle
}

func (e *Elevator_s) AddCabCall(floor int) {
	e.cabCalls[floor] = true
}

func (e *Elevator_s) RemoveCabCall(floor int) {
	e.cabCalls[floor] = false
}

func (e Elevator_s) GetAllCabCalls() []bool {
	return e.cabCalls
}

func (e Elevator_s) Marshal() ([]byte, error) {
	j, err := json.Marshal(struct {
		Hostname        string
		CurrentFloor    int
		NumFloors       int
		PrevFloor       int
		DirectionMoving Direction_e
		HallCalls       []HallCall_s
		CabCalls        []bool
		IsWorking       bool
	}{
		Hostname:        e.hostname,
		CurrentFloor:    e.currentFloor,
		NumFloors:       e.NumFloors,
		PrevFloor:       e.prevFloor,
		DirectionMoving: e.directionMoving,
		HallCalls:       e.hallCalls,
		CabCalls:        e.cabCalls,
		IsWorking:       e.isWorking,
	})
	if err != nil {
		return nil, err
	}
	return j, nil

}

func UnmarshalElevatorState(elevatorBytes []byte) Elevator_s { // Yeet into store?
	type tempElevator_s struct {
		Hostname        string
		CurrentFloor    int
		NumFloors       int
		PrevFloor       int
		DirectionMoving Direction_e
		HallCalls       []HallCall_s
		CabCalls        []bool
		IsWorking       bool
	}
	var tempElevator tempElevator_s
	json.Unmarshal(elevatorBytes, &tempElevator)
	elevator := Elevator_s{
		hostname:        tempElevator.Hostname,
		currentFloor:    tempElevator.CurrentFloor,
		prevFloor:       tempElevator.PrevFloor,
		NumFloors:       tempElevator.NumFloors,
		directionMoving: tempElevator.DirectionMoving,
		hallCalls:       tempElevator.HallCalls,
		cabCalls:        tempElevator.CabCalls,
		isWorking:       tempElevator.IsWorking,
	}

	return elevator
}
