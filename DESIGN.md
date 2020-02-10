# Virtual no-master ring network with broadcast and send/receive

## Overview

The network is organized as a ring of nodes, where they are sorted on the time they joined the network. The node which joined first is called HEAD, whereas the node which joined last is called TAIL. This way, no master is needed, and fault tolerance can easily be added by higher level code without a need for a single source of truth.

### Joining the network

There are two processes responsible in joining a network:

1. The TAIL does at all times listen for `JOIN` messages on `#.#.#.255`. When such message is received, the node which sent the message is added as a new TAIL, and the previous TAIL is connected to the new TAIL instead of its HEAD. Now the ring has expanded.

    **Before:**
    ```
    (TAIL -> ) HEAD -> #2 -> #3 -> (...) -> #(n-1) -> TAIL ( -> HEAD)
    ```

    **After:**
    ```
    (TAIL -> ) HEAD -> #2 -> #3 -> (...) -> #(n-1) -> #n -> TAIL ( -> HEAD)
    ```

2. A node which wants to join the network sends at an interval `JOIN_MSG_INTERVAL` a `JOIN` message on `#.#.#.255`. At the same time it should act as a TAIL listening for other `JOIN` messages.

    This solves the problem of creating the network initially when no network exists. Since `JOIN` messages are not sent continuously, there will be a race condition where one of two nodes will send the `JOIN` message first. The node which sends this message will become HEAD, whereas the node which receives the message becomes TAIL. The `JOIN` message should be done in a full handshake such that a node cannot be TAIL in two networks. Also, a node should node should not start to listen for `JOIN` messages BEFORE it has tried to join a network.

### Checking node availabilty

At intervals of `PING_INTERVAL`, each `node #i` sends a dummy TCP datagram to `node #(i+1)` to ensure that the node is alive and responding. This is done because TCP does not check whether or not a connection is working at all times - it merely makes it available for the program to use for communication.

If no response are received, `node #(i+1)` is considered unavailable, and so `discard_node(node_ip)` is issued to kick it off the network.

1. If a function `check_health() string` is provided as a callback to the initialization of the network, this function is run, and a string (most probably this should be a SHA256 or CRC-20 of the state synced over the network) is returned.

2. `node #i`then requests from  `node #(i+1)` the result of its `check_health() string`.

3. From here, there are multiple possible scenarioes:
    1. `node #(i+1)` returns the same result from `check_health()` as `node #1`. Since the node responds and they give the same result from the `check_health()`, they are both considered healthy. No further actions should be taken.
    2. One node returns a string different from the other. We now have a responding node, but their states are out of sync! To find out which node has the state which is considered "correct" by the network, `node #i` requests a `check_health()` from `node #(i-1)`. Hopefully, two nodes returns the same string. Let the one of `node #1` and `node #(i+1)` which holds the "majority votes" be `ip_synced`, and the remaining be `ip_to_sync`. Then issuing `sync_state(ip_synced, ip_to_sync)` will transmit the state the right direction and make them both in sync.
    3. `node #(i+1)` does not respond within a time of size `MAX_WAIT_TIME_S`. The node should be considered unavailable, and needs to be disconnected from the network. This is done with the `discard_node(node_ip)` command.


### Broadcasting messages

In its most simple form, a broadcast message can be sent to `#.#.#.255` using UDP, and it is then distributed to all nodes connected to the subnet. However, since UDP does not guarantee delivery, it requires a custom implemented handshake, and this can be quite cumbersome and error prone.

A quite compelling alternative would be to use a list of the IP addresses of the nodes, and emulate a broadcast by sending a TCP datagram to each and every node on the network. If a node does not respond, we first try to retransmit the datagram. If the node is still unavailable, then we simply `discard_node(node_ip)` - `node_ip` being the IP of the malfunctioning node. This would be the preferred implementation, had it not been for the fact that the order in which the nodes receive messages makes the health check unusable. Also, it is not scalable if there are lots of nodes connected to the network.

Instead, we have taken advantage of the ring topology of the network.



### Peer-to-peer messages


### Message format

```

```

### Interface

#### Network
```golang
network.init(bcast_ack_cb, ping_failed_cb)
network.broadcast(message)
network.send_to(ip, message)
network.get_peers()
network.peer_update()
network.receive(buffer)
```
#### State

**GLOBAL**:
**Nodes**:
| IP_ADDRESS    | CURRENT_FLOOR | DIRECTION | HALL_CALLS_A | CAB_CALLS_A
| ------------- | ------------- | --------- | ---------- | ---------------|
| HEAD          |  3            |      UP   | [call_1, call_2]   | [call_1, call_2] |
| #1            | Content Cell  | |
| #2            | Content Cell  | |
| TAIL          | Content Cell  | |


**NB**: We use **_A** postfix for assigned calls, and **_P** for physical calls.

**External Lights** (same for each elevator):
| FLOOR | UP | DOWN |
| --- | --- | --- |
| 1 | true | false |
| 2 | true | false |
| 3 | true | false |
| 4 | true | false |

**LOCAL**:

**Internal lights** (local to the elevator, so this is not shared):
| FLOOR | is_on |
| --- | --- |
| 1 | true |
| 2 | false |
| 3 | false |
| 4 | true |





```
state.get_node(node_ip)
state.add_node(node_ip, current_floor, current_direction, hall_calls_a, cab_calls_a)
state.add_call(node_ip, call_type, call)
state.remove_call(node_ip, call_type, call_id)
state.set_floor(node_ip, floor)
state.set_direction(node_ip, direction)
state.get_lights(floor_number)
```

#### Event scheduler
```
es.add_call(call_type, call)
es.set_floor(floor)
es.set_direction(direction)
```

#### Hardware driver
```
    elevio.Init(addr string, numFloors int)
    elevio.SetMotorDirection(dir MotorDirection)
    elevio.SetButtonLamp(button ButtonType, floor int, value bool)
    elevio.SetFloorIndicator(floor int)
    elevio.SetDoorOpenLamp(value bool)
    elevio.SetStopLamp(value bool)
    elevio.PollButtons(receiver chan<- ButtonEvent)
    elevio.PollFloorSensor(receiver chan<- int)
    elevio.PollStopButton(receiver chan<- bool)
    elevio.PollObstructionSwitch(receiver chan<- bool)
```
