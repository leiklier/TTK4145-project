package elevators

import (
	"sync"
)

type hallCall_s struct {
	IsUp   bool
	IsDown bool
}

type direction_t int

const (
	DirectionUp direction_t = iota
	DirectionDown
	DirectionIdle
)

type Elevator_s struct {
	mutex        sync.Mutex
	ip           string
	currentFloor int
	NumFloors    int
	direction    direction_t
	hallCalls    []hallCall_s
	cabCalls     []bool
}

func New(peerIP string, numFloors int, currentFloor int) Elevator_s {
	elevator := Elevator_s{
		ip:           peerIP,
		currentFloor: currentFloor,
		NumFloors:    numFloors,
		direction:    DirectionIdle,
		hallCalls:    make([]hallCall_s, numFloors),
		cabCalls:     make([]bool, numFloors),
	}

	return elevator
}

func (e Elevator_s) GetIP() string {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	return e.ip
}

func (e Elevator_s) GetCurrentFloot() int {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	return e.currentFloor
}

func (e Elevator_s) SetCurrentFloor(currentFloor int) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.currentFloor = currentFloor
}

func (e Elevator_s) GetDirection() direction_t {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	return e.direction
}

func (e Elevator_s) SetDirection(newDirection direction_t) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.direction = newDirection
}
