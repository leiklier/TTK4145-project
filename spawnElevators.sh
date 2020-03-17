#/bin/bash

tmux \
  new-session  "./SimElevatorServer --port 15657 ; read" \; \
  split-window "./SimElevatorServer --port 15658 ; read" \; \
  split-window "./SimElevatorServer --port 15659 ; read" \; \
  select-layout even-vertical