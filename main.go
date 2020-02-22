package main

import (
	"time"

	".network/ring"
)

func main() {
	time.Sleep(10 * time.Second)
	ring.Init()
}
