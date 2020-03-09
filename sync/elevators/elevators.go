package elevators

type HallCall_s struct {
	Floor     int
	Direction Direction_e
}

type Direction_e int

const (
	DirectionUp   Direction_e = 1
	DirectionDown             = -1
	DirectionBoth             = 2 // Only used for HallCalls
	DirectionIdle             = 0
)

type Elevator_s struct {
	ip              string
	currentFloor    int
	NumFloors       int
	PrevFloor       int
	directionMoving Direction_e
	hallCalls       []HallCall_s
	cabCalls        []bool
}

func New(peerIP string, numFloors int, currentFloor int) Elevator_s {
	elevator := Elevator_s{
		ip:              peerIP,
		currentFloor:    currentFloor,
		PrevFloor:       currentFloor, // initialize with same floor
		NumFloors:       numFloors,
		directionMoving: DirectionIdle,
		hallCalls:       make([]HallCall_s, numFloors),
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

func (e Elevator_s) GetDirectionMoving() Direction_e {
	return e.directionMoving
}

func (e Elevator_s) SetDirectionMoving(newDirection Direction_e) {
	e.directionMoving = newDirection
}

func (e Elevator_s) GetHallCalls() []HallCall_s {
	return e.hallCalls
}

func (e Elevator_s) AddHallCall(hallCall HallCall_s) {
	if hallCall.Direction == DirectionUp && e.hallCalls[hallCall.Floor].Direction == DirectionDown {
		e.hallCalls[hallCall.Floor].Direction = DirectionBoth
	} else if hallCall.Direction == DirectionDown && e.hallCalls[hallCall.Floor].Direction == DirectionUp {
		e.hallCalls[hallCall.Floor].Direction = DirectionBoth
	} else {
		e.hallCalls[hallCall.Floor].Direction = hallCall.Direction
	}
}

func (e Elevator_s) RemoveHallCalls(floor int) {
	e.hallCalls[floor].Direction = DirectionIdle
}

func (e Elevator_s) AddCabCall(floor int) {
	e.cabCalls[floor] = true
}

func (e Elevator_s) RemoveCabCall(floor int) {
	e.cabCalls[floor] = false
}
