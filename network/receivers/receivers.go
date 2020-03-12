package receivers

// Enums
const (
	GetReceiver = iota
)

type Receiver struct {
	Name    string
	Channel chan []byte
}

type ControlSignal struct {
	Command         int
	Payload         string
	ResponseChannel chan Receiver
}

var gControlChannel = make(chan ControlSignal, 100)

func init() {
	go receiverServer()
}
func GetChannel(name string) chan []byte {
	controlSignal := ControlSignal{
		Command:         GetReceiver,
		Payload:         name,
		ResponseChannel: make(chan Receiver),
	}
	gControlChannel <- controlSignal

	receiver := <-controlSignal.ResponseChannel
	return receiver.Channel
}

func receiverServer() {
	var receivers []Receiver
	for {
		controlSignal := <-gControlChannel

		switch controlSignal.Command {
		case GetReceiver:
			name := controlSignal.Payload
			receiverDoesExist := false

			for _, receiver := range receivers {
				if receiver.Name == name {
					receiverDoesExist = true
					controlSignal.ResponseChannel <- receiver
					break
				}
			}

			if !receiverDoesExist {
				// No such receiver exists, so create a new one, add it to our list
				// of receivers and return it on response:
				receiver := Receiver{
					Name:    name,
					Channel: make(chan []byte, 100),
				}
				receivers = append(receivers, receiver)
				controlSignal.ResponseChannel <- receiver
			}

			break
		}
	}
}
