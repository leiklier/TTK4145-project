package main

import (
	"fmt"

	"../store"
)

var cc = []bool{true, false, false, false}

func main() {
	store.WriteCCBackup(cc, store.BACKUPNAME)
	fmt.Println("Saved to", store.BACKUPNAME)
	fmt.Println("Reading from:", store.BACKUPNAME)
	s := store.ReadCCBackup(store.BACKUPNAME)
	fmt.Println(s)
}
