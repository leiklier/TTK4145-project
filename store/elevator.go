package elevator

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
	ip           string
	CurrentFloor int
	NumFloors    int
	Direction    direction_t
	HallCalls    []hallCall_s
	CabCalls     []bool
}

func New(peerIP string, numFloors int, currentFloor int) {
	elevator := Elevator_s{
		ip:           peerIP,
		CurrentFloor: currentFloor,
		NumFloors:    numFloors,
		Direction:    DirectionIdle,
		HallCalls:    make([]hallCall_s, numFloors),
		CabCalls:     make([]bool, numFloors),
	}

	return elevator
}

func (e Elevator_s) GetIP() string {
	return e.ip
}

func (e Elevator_s) SetCurrentFloor(currentFloor int) {
	e.CurrentFloor = currentFloor
}
