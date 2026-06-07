#!/bin/bash
# Persistent VNC and Streaming Manager
# This script keeps VNC, noVNC, and ffmpeg streaming running

LOG_DIR="/home/openhands/vnc_logs"
mkdir -p "$LOG_DIR"

echo "=============================================="
echo "VNC/Streaming Persistent Manager"
echo "Started at: $(date)"
echo "=============================================="

# Function to start VNC server
start_vnc() {
    echo "[$(date)] Starting VNC server..."
    export DISPLAY=:1
    
    # Kill existing
    vncserver -kill :1 2>/dev/null || true
    sleep 1
    
    # Start VNC with Xfce4
    vncserver :1 -geometry 1280x720 -depth 24 -localhost no \
        -xstartup /home/openhands/.vnc/xstartup \
        >> "$LOG_DIR/vnc.log" 2>&1
    
    echo "[$(date)] VNC server started on :1"
}

# Function to start Websockify
start_websockify() {
    echo "[$(date)] Starting Websockify..."
    pkill -f "websockify.*6080" 2>/dev/null || true
    sleep 1
    
    websockify --web=/usr/share/novnc 6080 localhost:5901 \
        >> "$LOG_DIR/websockify.log" 2>&1 &
    
    echo "[$(date)] Websockify started on port 6080"
}

# Function to start ffmpeg screen capture
start_ffmpeg() {
    echo "[$(date)] Starting ffmpeg screen capture..."
    pkill -f "ffmpeg.*screen" 2>/dev/null || true
    sleep 1
    
    # Capture screen and stream (using X11 capture if available)
    nohup ffmpeg -f x11grab -i :1 -c:v libx264 -preset ultrafast \
        -tune zerolatency -b:v 1000k -maxrate 1000k \
        -s 1280x720 -r 30 \
        -f flv rtmp://localhost/live/screen \
        >> "$LOG_DIR/ffmpeg.log" 2>&1 &
    
    echo "[$(date)] ffmpeg started"
}

# Initial startup
start_vnc
sleep 2
start_websockify
sleep 1

# Keep running and restart services if they die
while true; do
    # Check VNC
    if ! pgrep -f "Xtigervnc :1" > /dev/null; then
        echo "[$(date)] VNC crashed, restarting..."
        start_vnc
        sleep 2
    fi
    
    # Check Websockify
    if ! pgrep -f "websockify.*6080" > /dev/null; then
        echo "[$(date)] Websockify crashed, restarting..."
        start_websockify
        sleep 1
    fi
    
    sleep 10
done