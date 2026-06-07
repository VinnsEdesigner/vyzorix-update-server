#!/bin/bash
# Persistent ffmpeg streaming script
# This script keeps ffmpeg running even after the session ends

VNC_HOST="${VNC_HOST:-localhost}"
VNC_PORT="${VNC_PORT:-5900}"
STREAM_URL="${STREAM_URL:-rtmp://localhost/live/stream}"
LOG_FILE="/tmp/ffmpeg_stream.log"

echo "Starting persistent ffmpeg stream..."
echo "VNC: $VNC_HOST:$VNC_PORT -> $STREAM_URL"
echo "Log file: $LOG_FILE"

# Cleanup function
cleanup() {
    echo "Stopping ffmpeg stream..."
    pkill -f "ffmpeg.*$VNC_HOST:$VNC_PORT" 2>/dev/null
    exit 0
}

trap cleanup SIGTERM SIGINT

# Keep ffmpeg running in a loop
while true; do
    echo "$(date): Starting ffmpeg..." >> "$LOG_FILE"
    
    # Use nohup to keep running after session ends
    # Capture VNC display and stream to RTMP
    nohup ffmpeg -framerate 30 -re -f fbdev -i /dev/fb0 \
        -c:v libx264 -preset ultrafast -tune zerolatency \
        -b:v 1000k -maxrate 1000k -bufsize 2000k \
        -s 1280x720 -g 60 \
        -f flv "$STREAM_URL" >> "$LOG_FILE" 2>&1 &
    
    FFmpeg_PID=$!
    echo "ffmpeg started with PID: $FFmpeg_PID"
    
    # Wait for ffmpeg to finish
    wait $FFmpeg_PID
    EXIT_CODE=$?
    
    echo "$(date): ffmpeg exited with code $EXIT_CODE, restarting..." >> "$LOG_FILE"
    sleep 2
done