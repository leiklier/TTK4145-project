package elevators

import "errors"

type hallCall_s struct {
	IsUp   bool
	IsDown bool
}

type Direction int

const (
	DirectionUp   Direction = 1
	DirectionDown           = -1
	DirectionIdle           = 0
)

type Elevator_s struct {
	ip              string
	currentFloor    int
	NumFloors       int
	directionMoving Direction
	hallCalls       []hallCall_s
	cabCalls        []bool
}

func New(peerIP string, numFloors int, currentFloor int) Elevator_s {
	elevator := Elevator_s{
		ip:              peerIP,
		currentFloor:    currentFloor,
		NumFloors:       numFloors,
		directionMoving: DirectionIdle,
		hallCalls:       make([]hallCall_s, numFloors),
		cabCalls:        make([]bool, numFloors),
	}

	return elevator
}

func (e Elevator_s) GetIP() string {
	return e.ip
}

func (e Elevator_s) GetCurrentFloor() int {
	return e.currentFloor
}

func (e Elevator_s) SetCurrentFloor(currentFloor int) {
	e.currentFloor = currentFloor
}

func (e Elevator_s) GetDirectionMoving() Direction {
	return e.directionMoving
}

func (e Elevator_s) SetDirectionMoving(newDirection Direction) {
	e.directionMoving = newDirection
}

func (e Elevator_s) GetHallCalls() []hallCall_s {
	return e.hallCalls
}

func (e Elevator_s) AddHallCall(floor int, direction Direction) error {
	if floor > e.NumFloors-1 {
		return errors.New("ERR_INVALID_FLOOR")
	}

	if direction == DirectionUp {
		e.hallCalls[floor].IsUp = true
	} else if direction == DirectionDown {
		e.hallCalls[floor].IsDown = true
	}

	return nil
}

func (e Elevator_s) RemoveHallCalls(floor int) error {
	if floor > e.NumFloors-1 {
		return errors.New("ERR_INVALID_FLOOR")
	}

	e.hallCalls[floor].IsUp = false
	e.hallCalls[floor].IsDown = false
	return nil
}

func (e Elevator_s) AddCabCall(floor int) error {
	if floor > e.NumFloors-1 {
		return errors.New("ERR_INVALID_FLOOR")
	}

	e.cabCalls[floor] = true
	return nil
}

func (e Elevator_s) RemoveCabCall(floor int) error {
	if floor > e.NumFloors-1 {
		return errors.New("ERR_INVALID_FLOOR")
	}

	e.cabCalls[floor] = false
	return nil
}
