package elevators

import "errors"

type hallCall_s struct {
	IsUp   bool
	IsDown bool
}

type MoveDirection_e int

const (
	DirectionUp MoveDirection_e = iota
	DirectionDown
	DirectionIdle
)

type HCDirection_e int

const (
	HCDirectionUp HCDirection_e = iota
	HCDirectionDown
)

type Elevator_s struct {
	ip              string
	currentFloor    int
	NumFloors       int
	directionMoving MoveDirection_e
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

func (e Elevator_s) GetCurrentFloot() int {
	return e.currentFloor
}

func (e Elevator_s) SetCurrentFloor(currentFloor int) {
	e.currentFloor = currentFloor
}

func (e Elevator_s) GetDirectionMoving() MoveDirection_e {
	return e.directionMoving
}

func (e Elevator_s) SetDirectionMoving(newDirection MoveDirection_e) {
	e.directionMoving = newDirection
}

func (e Elevator_s) AddHallCall(floor int, direction HCDirection_e) error {
	if floor > e.NumFloors-1 {
		return errors.New("ERR_INVALID_FLOOR")
	}

	if direction == HCDirectionUp {
		e.hallCalls[floor].IsUp = true
	} else if direction == HCDirectionDown {
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
