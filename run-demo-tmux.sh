#!/bin/bash

# Runs server and client example in split tmux panes
# Only works when already inside a tmux session

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Check if we're inside a tmux session
if [ -z "$TMUX" ]; then
    echo "Error: This script must be run from inside a tmux session."
    exit 1
fi

CURRENT_WINDOW=$(tmux display-message -p '#I')
CURRENT_SESSION=$(tmux display-message -p '#S')

# Kill all panes in current window except the first one
PANE_COUNT=$(tmux list-panes | wc -l)

if [ $PANE_COUNT -gt 1 ]; then
    tmux kill-pane -a
fi

# Clear the current pane and stop any running processes
tmux send-keys C-c
sleep 0.2
tmux send-keys "clear" C-m

# Start server in current (left) pane
tmux send-keys "go run examples/server/main.go" C-m

# Split vertically and run client in right pane
tmux split-window -h -c "$PROJECT_DIR"

# Start client in right pane
tmux send-keys "sleep 0.2; go run examples/client/main.go" C-m

# Select left pane (server)
tmux select-pane -L
