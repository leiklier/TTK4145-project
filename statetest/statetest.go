package main

import (
	"../store"
	"fmt"
)



var hc1 = []store.HallCall {store.HC_none,store.HC_up,store.HC_none,store.HC_none}
var hc2 = []store.HallCall {store.HC_up,store.HC_up,store.HC_down,store.HC_none}
var hc3 = []store.HallCall {store.HC_up,store.HC_down,store.HC_none,store.HC_down}



func main() {
	fmt.Println(hc1)
}