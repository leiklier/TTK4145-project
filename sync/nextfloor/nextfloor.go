package nextfloor

import (
	"fmt"
	"time"

	"../../network/peers"
	"../elevators"
)

// SubscribeToDestinationUpdates
func SubscribeToDestinationUpdates(nextFloor chan int) {
	for {
		elev, err := store.getElevator(peers.GetRelativeTo(peers.Self, 0))
		if err != nil {
			fmt.Println("Can't get elevator, here is the error:")
			fmt.Println(err)
			fmt.Println("Retrying in 1 second...")
			time.Sleep(1)
			continue
		}
		currDir := elev.Direction_e
		prevFloor := elev.prevFloor
		currFloor := elev.currentFloor
		cabCalls := elev.cabCalls
		hallCalls := elev.hallCalls

		switch currFloor {
		// Bottom floor
		case 0:
			nearestCab := 0
			for i, cab := range cabCalls {
				if cab {
					nearestCab = i
					break
				}
			}
			nearestHall := 0
			for i, hc := range hallCalls {
				if hc.Direction_e != elevators.DirectionIdle {
					nearestHall = i
					break
				}
			}
			if nearestCab < nearestHall {
				nextFloor <- nearestCab
			} else if nearestHall < nearestCab {
				nextFloor <- nearestHall
			} else {
				nextFloor <- nearestCab // Spiller ingen rolle
			}

		// Top floor
		case store.NumFloors - 1:
			nearestCab := store.NumFloors + 1 // høyere enn høyest
			for i := store.NumFloors - 1; i >= 0; i-- {
				if cabCalls[i] {
					nearestCab = i
					break
				}
			}
			nearestHall := store.NumFloors + 1 // Høyere enn høyest
			for i := store.NumFloors - 1; i >= 0; i-- {
				if hallCalls[i].Direction_e != elevators.DirectionIdle {
					nearestHall = i
					break
				}
			}
			if nearestCab < nearestHall {
				nextFloor <- nearestCab
			} else if nearestHall < nearestCab {
				nextFloor <- nearestHall
			} else {
				nextFloor <- nearestCab // Spiller ingen rolle
			}

		default:
			if currFloor > prevFloor {
				// Vi skal opp dersom det er noe å ta oppover

				// Cabcheck:
				nextCab := store.NumFloors + 1 // Defaulter med noe som er for høyt
				for i, cab := range cabCalls {
					if i <= currFloor {
						continue // Vi sjekker kun oppover
					} else if cab {
						nextCab = i
						break
					}
				}
				if nextCab == store.NumFloors+1 {
					// Ingenting oppover, vi sjekker nedover.
					for i := store.NumFloors - 1; i >= 0; i-- {
						if i >= currFloor {
							continue // Sjekker kun nedover
						} else {
							if cabCalls[i] {
								nextCab = i
								break
							}
						}
					}
				}
				// Cabcheck over

				// Hallcheck:
				nextHall := store.NumFloors + 1
				sameDirection := false // Vil helst ha oppover
				for i, hc := range hallCalls {
					if i <= currFloor {
						continue // Vi sjekker kun oppover
					} else {
						if hc.Direction_e != elevators.DirectionIdle {
							nextHall = i
							// Setter at det er samme retning om det er oppover, ELLER at det er i
							// 4 etasje. Kan breake bare om det er samme retning, fordi da er det
							// nærmest og rett retning= optimalt!
							sameDirection = (hc.Direction_e == elevators.DirectionUp || i == store.NumFloors-1)
							if sameDirection {
								break
							}
						}
					}
				}
				if nextHall == store.NumFloors+1 {
					// Ingenting oppover, vi sjekker nedover.
					for i := store.NumFloors - 1; i >= 0; i-- {
						if i >= currFloor {
							continue // Sjekker kun nedover
						} else {
							if hallCalls[i].Direction_e != elevators.DirectionIdle {
								nextHall = i
								// Her spiller retningen på HC ingen rolle, ettersom vi må
								// nedover uansett. Vi lar det forbli default, false og
								// kan trygt breake ut
								break
							}
						}
					}
				}
				// Hallcheck over

				// Sammenlikne
				if nextCab == store.NumFloors+1 && nextHall == store.NumFloors+1 {
					// Det er ingenting å gjøre, nextfloor er da currentfloor
					nextFloor <- currFloor
				}

				if nextCab > currFloor && nextHall > currFloor && sameDirection {
					// Finne ut hvilke som er nermest
					if nextCab < nextHall {
						nextFloor <- nextCab
					} else {
						nextFloor <- nextHall
					}
				}

			} else if currFloor < prevFloor {
				// Vi skal nedover dersom det er noe å ta nedover

				//Cabcheck:
				nextCab := store.NumFloors + 1 // Default for høyt
				for i := store.NumFloors - 1; i >= 0; i-- {
					for i := store.NumFloors - 1; i >= 0; i-- {
						if i >= currFloor {
							continue // Sjekker kun nedover
						} else {
							if cabCalls[i] {
								nextCab = i
								break
							}
						}
					}
				}
				if nextCab == store.NumFloors+1 {
					// Ingenting nedover, vi sjekker oppover
					for i, cab := range cabCalls {
						if i <= currFloor {
							continue // Vi sjekker kun oppover
						} else {
							if cab {
								nextCab = i
								break
							}
						}
					}
				}
				// Cabcheck over

				//Hallcheck
				nextHall := store.NumFloors + 1
				sameDirection := false // Vil helst ha nedover
				for i := store.store.store.store.NumFloors - 1; i >= 0; i-- {
					if i >= currFloor {
						continue // Sjekker kun nedover
					} else {
						if hallCalls[i].Direction_e != elevators.DirectionIdle {
							nextHall = i
							// Samme retning nedover eller at vi er i 0te etasje.
							sameDirection = hallCalls[i].Direction_e == elevators.DirectionDown || i == 0
							if sameDirection {
								break
							}
						}
					}
				}
				if nextHall == store.store.store.store.NumFloors+1 {
					// Ingenting nedover, vi sjekker oppover
					for i, hc := range hallCalls {
						if i <= currFloor {
							continue // Vi sjekker kun oppover
						} else {
							if hc.Direction_e != elevators.DirectionIdle {
								nextHall = i
								// Her spiller retningen på HC ingen rolle, ettersom vi må
								// oppover uansett. Vi lar det forbli default, false og
								// kan trygt breake ut
								break
							}
						}
					}
				}
				// Hallcheck over

				// Sammenlinke

			}

		}

	}

	// NB!! Må bestemme oss for clear variant. Dersom vi
	// tømmer alle på et gulv er det umulig at det er noe ordre å ta på dette
	// gulvet

}
