### Funksjonalitet heisen må ha

1. Kjøre opp og ned.
2. Den henter andre calls automatisk (du må ikke trykke en gang til).
3. Den stopper når den skal cleare (den kjører ikke bare rett fordi).
4. Den kan kjøre videre i "single-mode". Den stopper ikke opp og venter på at du skal trykke igjen.
5. Finne ut at heisen ikke funker


# Fra oppgaven:

## No orders are lost


  Once the light on a hall call button (buttons for calling an elevator to that floor; top 6 buttons on the control panel) is turned on, an elevator should arrive at that floor
  
  - Funker under normale forhold

  Similarly for a cab call (for telling the elevator what floor you want to exit at; front 4 buttons on the control panel), but only the elevator at that specific workspace should take the order

  - Funker også under normale forhold

This means handling network packet loss, losing network connection entirely, software that crashes, and losing power - both to the elevator motor and the machine that controls the elevator

  1. For cab orders, handling loss of power/software crash implies that the orders are executed once service is restored

  - Blir ikke gjort

  2. The time used to detect these failures should be reasonable, ie. on the order of magnitude of seconds (not minutes)

  - Blir ikke gjort 
    
3. Network packet loss is not an error, and can occur at any time
- Chaise

If the elevator is disconnected from the network, it should still serve all the currently active orders (ie. whatever lights are showing)

1. It should also keep taking new cab calls, so that people can exit the elevator even if it is disconnected from the network
- Må testes

2. The elevator software should not require reinitialization (manual restart) after intermittent network or motor power loss
- Må testes

## Multiple elevators should be more efficient than one


The orders should be distributed across the elevators in a reasonable way. Ex: If all three elevators are idle and two of them are at the bottom floor, then a new order at the top floor should be handled by the closest elevator (ie. neither of the two at the bottom).
- Funker

You are free to choose and design your own "cost function" of some sort: Minimal movement, minimal waiting time, etc.
- Jaja

The project is not about creating the "best" or "optimal" distribution of orders. It only has to be clear that the elevators are cooperating and communicating.

- Vår cost function er uansett best så dette er blitt gjort med glans, bra jobbet Jakob.

## An individual elevator should behave sensibly and efficiently

1. No stopping at every floor "just to be safe"
- Nei er du gal

2. The hall "call upward" and "call downward" buttons should behave differently. Ex: If the elevator is moving from floor 1 up to floor 4 and there is a downward order at floor 3, then the elevator should not stop on its way upward, but should return back to floor 3 on its way down
- Seffern

## The lights and buttons should function as expected

The hall call buttons on all workspaces should let you summon an elevator

- Funker

Under normal circumstances, the lights on the hall buttons should show the same thing on all workspaces. Under circumstances with high packet loss, at least one light must work as expected
- Funker ikke

The cab button lights should not be shared between elevators

- Funker

The cab and hall button lights should turn on as soon as is reasonable after the button has been pressed

- Funker

The cab and hall button lights should turn off when the corresponding order has been serviced

- Funker

The "door open" lamp should be used as a substitute for an actual door, and as such should not be switched on while the elevator is moving. The duration for keeping the door open should be in the 1-5 second range

- Funker

## Unspecified behaviour

Some things are left intentionally unspecified. Their implementation will not be tested, and are therefore up to you. Which orders are cleared when stopping at a floor: You can clear only the orders in the direction of travel, or assume that everyone enters/exits the elevator when the door opens

- Vi clearer alt

How the elevator behaves when it cannot connect to the network (router) during initialization. You can either enter a "single-elevator" mode, or refuse to start

- Ikke bestemt

How the hall (call up, call down) buttons work when the elevator is disconnected from the network. You can optionally refuse to take these new orders

- Ikke bestemt

Stop button & obstruction switch are disabled. Their functionality (if/when implemented) is up to you.

- Dropper vel dette?