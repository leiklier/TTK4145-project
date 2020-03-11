package nextfloor

import (
	"../../network/peers"
	"../elevators"
	"../store"

	"time"
)

var selfIP = peers.GetRelativeTo(peers.Self, 0)

// SubscribeToDestinationUpdates
func SubscribeToDestinationUpdates(nextFloor chan int) {
	for {
		// elev, err := store.GetElevator(peers.GetRelativeTo(peers.Self, 0))
		// if err != nil {
		// 	fmt.Println("Can't get elevator, here is the error:")
		// 	fmt.Println(err)
		// 	fmt.Println("Retrying in 1 second...")
		// 	time.Sleep(1)
		// 	continue
		// }
		// currDir, _ := store.GetDirectionMoving(selfIP)
		time.Sleep(time.Duration(2 * time.Second))

		prevFloor, _ := store.GetPreviousFloor(selfIP)
		currFloor, _ := store.GetCurrentFloor(selfIP)
		cabCalls, _ := store.GetAllCabCalls(selfIP)
		hallCalls, _ := store.GetAllHallCalls(selfIP)

		switch currFloor {
		// Bottom floor
		case 0:
			nearestCab := store.NumFloors + 1
			for i, cab := range cabCalls {
				if cab {
					nearestCab = i
					break
				}
			}
			nearestHall := store.NumFloors + 1
			for i, hc := range hallCalls {
				if hc.Direction != elevators.DirectionIdle {
					nearestHall = i
					break
				}
			}
			if nearestCab == store.NumFloors +1 && nearestHall ==store.NumFloors +1 {
				// Nothing has changed.
				nextFloor <- currFloor
			}else if nearestCab < nearestHall {
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
				if hallCalls[i].Direction != elevators.DirectionIdle {
					nearestHall = i
					break
				}
			}

			if nearestCab == store.NumFloors + 1 && nearestHall== store.NumFloors + 1 {
				nextFloor <- currFloor	
			} else if nearestCab < nearestHall {
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
						if hc.Direction != elevators.DirectionIdle {
							nextHall = i
							// Setter at det er samme retning om det er oppover, ELLER at det er i
							// 4 etasje. Kan breake bare om det er samme retning, fordi da er det
							// nærmest og rett retning= optimalt!
							sameDirection = (hc.Direction == elevators.DirectionUp || i == store.NumFloors-1)
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
							if hallCalls[i].Direction != elevators.DirectionIdle {
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
				// Det er ingenting å gjøre, nextfloor er da currentfloor
				if nextCab == store.NumFloors+1 && nextHall == store.NumFloors+1 {
					nextFloor <- currFloor

					// Oppover sjekking
				} else if nextCab > currFloor || nextHall > currFloor {
					// Begge oppfylt, HallCall samme retning, begge like aktuelle
					if nextCab > currFloor && nextHall > currFloor && sameDirection {
						// Finne ut hvilke som er nermest
						if nextCab < nextHall {
							nextFloor <- nextCab
						} else {
							nextFloor <- nextHall
						}

						// Begge oppfylt, men HallCall i feil retning
					} else if nextCab > currFloor && nextHall > currFloor && !sameDirection {
						// Da er cabCall vi henter.
						nextFloor <- nextCab

						// Bare nextCab er oppfylt, da henter vi den
					} else if nextCab > currFloor && !(nextHall > currFloor) {
						nextFloor <- nextCab
					} else if !(nextCab > currFloor) && nextHall > currFloor {
						// Bare HallCall, da spiller det ingen rolle hvilen vei
						nextFloor <- nextHall

						// Nedover sjekking, hverken cab eller hall er oppover git add .
					} else {
						// Vi gir litt faen forid det var tidspress, og sender bare den som er nærmest
						if nextCab < nextHall {
							nextFloor <- nextCab
						} else {
							nextFloor <- nextHall
						}
					}
					// Moving  down

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
					for i := store.NumFloors - 1; i >= 0; i-- {
						if i >= currFloor {
							continue // Sjekker kun nedover
						} else {
							if hallCalls[i].Direction != elevators.DirectionIdle {
								nextHall = i
								// Samme retning nedover eller at vi er i 0te etasje.
								sameDirection = hallCalls[i].Direction == elevators.DirectionDown || i == 0
								if sameDirection {
									break
								}
							}
						}
					}
					if nextHall == store.NumFloors+1 {
						// Ingenting nedover, vi sjekker oppover
						for i, hc := range hallCalls {
							if i <= currFloor {
								continue // Vi sjekker kun oppover
							} else {
								if hc.Direction != elevators.DirectionIdle {
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
					// Sammenlinke
					// Sammenlikne
					// Det er ingenting å gjøre, nextfloor er da currentfloor
					if nextCab == store.NumFloors+1 && nextHall == store.NumFloors+1 {
						nextFloor <- currFloor

						// Oppover sjekking
					} else if nextCab < currFloor || nextHall < currFloor {
						// Begge oppfylt, HallCall samme retning, begge like aktuelle
						if nextCab > currFloor && nextHall > currFloor && sameDirection {
							// Finne ut hvilke som er nermest
							if nextCab > nextHall {
								nextFloor <- nextCab
							} else {
								nextFloor <- nextHall
							}

							// Begge oppfylt, men HallCall i feil retning
						} else if nextCab < currFloor && nextHall < currFloor && !sameDirection {
							// Da er cabCall vi henter.
							nextFloor <- nextCab

							// Bare nextCab er oppfylt, da henter vi den
						} else if nextCab < currFloor && !(nextHall < currFloor) {
							nextFloor <- nextCab
						} else if !(nextCab < currFloor) && nextHall < currFloor {
							// Bare HallCall, da spiller det ingen rolle hvilen vei
							nextFloor <- nextHall

							// Nedover sjekking, hverken cab eller hall er oppover git add .
						} else {
							// Vi gir litt faen forid det var tidspress, og sender bare den som er nærmest
							if nextCab > nextHall {
								nextFloor <- nextCab
							} else {
								nextFloor <- nextHall
							}
						}
						// Moving  down

					}
					// Sammenlikne over

				}

			}

		}

		// NB!! Må bestemme oss for clear variant. Dersom vi
		// tømmer alle på et gulv er det umulig at det er noe ordre å ta på dette
		// gulvet

	}
}

func SubscribeToDestinationUpdates2(nextFloor <-chan int) {
	for {
		time.Sleep(time.Duration(2 * time.Second))

		cabCalls, _ := store.GetAllCabCalls(selfIP)
		hallCalls, _ := store.GetAllHallCalls(selfIP)
		currDir, _ := store.GetDirectionMoving(selfIP)

		switch currDir {
		case elevators.DirectionDown:
			// Do smth
		case elevators.DirectionUp:
			// Do smth
		case elevators.DirectionIdle:
			// So smth
		default:
			// Da må det vel være dir both da.. det er jo litt synd om det skjer
		}
	}
}
