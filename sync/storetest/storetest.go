package main

import (
	"fmt"
	"strings"

	"../store"
)

var cc = []bool{true, false, false, false}

func main() {
	store.WriteCCBackup(cc, store.BACKUPNAME)
	fmt.Println("Saved to", store.BACKUPNAME)
	fmt.Println("Reading from:", store.BACKUPNAME)
	s := store.ReadCCBackup(store.BACKUPNAME)
	fmt.Println(s)

	fmt.Println("Splitting")

	onlyCC := strings.Split(s, ";")[0]
	fmt.Println(onlyCC)
	ccs := strings.Split(onlyCC, ",")
	var newcc []bool
	for _, i := range ccs {
		if i == "true" {
			newcc = append(newcc, true)
		} else {
			newcc = append(newcc, false)
		}
	}
	fmt.Println(newcc)
}
