package nextfloor

import (
	"fmt"

	"../../network/peers"
	"../elevators"
	"../store"
)

// SubscribeToDestinationUpdates
// Kan kjøres både som goroutine eller som funksjonskall
func SubscribeToDestinationUpdates(nextFloor chan int) {
	for {
		cabCalls, _ := store.GetAllCabCalls(peers.GetRelativeTo(peers.Self, 0))
		hallCalls, _ := store.GetAllHallCalls(peers.GetRelativeTo(peers.Self, 0))
		currDir, _ := store.GetDirectionMoving(peers.GetRelativeTo(peers.Self, 0))
		currFloor, _ := store.GetCurrentFloor(peers.GetRelativeTo(peers.Self, 0))

		var nextCab int

		switch currDir {
		case elevators.DirectionDown:
			// Looking for hot single moms downstairs...
			for i := store.NumFloors - 1; i >= 0; i-- {
				if i >= currFloor {
					continue // Sjekker kun nedover
				} else {
					if cabCalls[i] {
						nextCab = i
						nextFloor <- nextCab
						break
					}
				}
			}
			for i := store.NumFloors - 1; i >= 0; i-- {
				if i >= currFloor {
					continue // Sjekker kun nedover
				} else {
					if hallCalls[i].Direction != elevators.DirectionIdle {
						nextHall := i
						if nextHall < nextCab {
							nextFloor <- nextHall
						}
						break
					}
				}
			}
		case elevators.DirectionUp:
			// Looking for hot single moms upstairs...
			for i, cab := range cabCalls {
				if i <= currFloor {
					continue // Vi sjekker kun oppover
				} else if cab {
					nextCab = i
					nextFloor <- nextCab
					break
				}
			}
			for i, hc := range hallCalls {
				if i <= currFloor {
					continue // Vi sjekker kun oppover
				} else {
					if hc.Direction != elevators.DirectionIdle {
						nextHall := i
						if nextHall < nextCab {
							nextFloor <- nextHall
						}
						break
					}
				}
			}
			// Denne kan egntlig håndere hele suppen hvis vi kun sjekker på idle
		case elevators.DirectionIdle:
			// Looking for hot single moms everywhere
			for counter := 1; counter < store.NumFloors; counter++ {
				lowerSearchIndex := currFloor - counter  // Sjekk om er under 0
				higherSearchIndex := currFloor + counter // Sjekk om er over NumFloors-1

				if !(lowerSearchIndex < 0) {
					if cabCalls[lowerSearchIndex] {
						nextFloor <- lowerSearchIndex
						break
					}
					if hallCalls[lowerSearchIndex].Direction != elevators.DirectionIdle {
						nextFloor <- lowerSearchIndex
						break
					}
				}
				if !(higherSearchIndex > store.NumFloors-1) {
					if cabCalls[higherSearchIndex] {
						nextFloor <- higherSearchIndex
						break
					}
					if hallCalls[higherSearchIndex].Direction != elevators.DirectionIdle {
						nextFloor <- higherSearchIndex
						break
					}
				}

			}
		default:
			fmt.Println("Get the Bible and pray!")
		}
	}
}
