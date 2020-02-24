package main

import (
	"time"

	"./network/ring"
)

func main() {
	time.Sleep(1 * time.Second)
	ring.Init()
}
