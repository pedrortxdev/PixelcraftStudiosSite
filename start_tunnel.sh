#!/bin/bash

# Resolve Script Directory
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Load environment variables from backend/.env relative to script
if [ -f "$DIR/backend/.env" ]; then
    # Use allexport to ensure vars are exported
    set -o allexport
    source "$DIR/backend/.env"
    set +o allexport
else
    echo "Error: backend/.env not found at $DIR/backend/.env"
    exit 1
fi

REMOTE_HOST=${SSH_TUNNEL_REMOTE_HOST:-35.182.175.91}
REMOTE_USER=${SSH_TUNNEL_REMOTE_USER:-ubuntu}
RAW_KEY_PATH=${SSH_TUNNEL_KEY_PATH:-LightsailDefaultKey-ca-central-1.pem}
LOCAL_PORT=250
REMOTE_PORT=25

# Sanitize Key Path (remove quotes, whitespace)
RAW_KEY_PATH=$(echo "$RAW_KEY_PATH" | tr -d '"' | tr -d "'")

# Resolve Key Path
# 1. Try absolute
if [ -f "$RAW_KEY_PATH" ]; then
    KEY_PATH="$RAW_KEY_PATH"
# 2. Try relative to script (stripping ../)
else
    CLEAN_NAME=$(basename "$RAW_KEY_PATH")
    if [ -f "$DIR/$CLEAN_NAME" ]; then
        KEY_PATH="$DIR/$CLEAN_NAME"
    else
        # Fallback: assume it's just in the root
        KEY_PATH="$DIR/LightsailDefaultKey-ca-central-1.pem"
    fi
fi

echo "Starting SSH Tunnel..."
echo "Remote: $REMOTE_USER@$REMOTE_HOST:$REMOTE_PORT"
echo "Local Port: $LOCAL_PORT"
echo "Key: $KEY_PATH"

# Ensure permissions
if [ -f "$KEY_PATH" ]; then
    chmod 600 "$KEY_PATH"
else
    echo "Error: Key file not found at $KEY_PATH"
    exit 1
fi

# Cleanup previous tunnel (if any)
PID=$(pgrep -f "ssh.*-L.*:$LOCAL_PORT:localhost:$REMOTE_PORT")
if [ -n "$PID" ]; then
    echo "Killing existing tunnel (PID: $PID)..."
    kill $PID
    sleep 2
fi

# Start Tunnel in FOREGROUND for PM2
# -N: no remote command
# -L: port forwarding
ssh -i "$KEY_PATH" -N -L 0.0.0.0:$LOCAL_PORT:localhost:$REMOTE_PORT $REMOTE_USER@$REMOTE_HOST -o StrictHostKeyChecking=no -o ServerAliveInterval=60 -o ServerAliveCountMax=3
