#!/bin/bash
# Scan internal/ws/ WebSocket hub

cd "$(dirname "$0")/.."

echo ""
echo "                    SCANNING: internal/ws/ WEBSOCKET HUB                                          "
echo ""
echo ""

echo ""
echo "1. internal/ws/ FILES"
echo ""
echo ""

if [ -d "internal/ws" ]; then
    echo "Files:"
    for f in internal/ws/*.go; do
        if [ -f "$f" ] && ! echo "$f" | grep -q "_test"; then
            funcs=$(grep -c "^func " "$f" 2>/dev/null || echo "0")
            types=$(grep -c "^type " "$f" 2>/dev/null || echo "0")
            echo "  $(basename $f): $types types, $funcs funcs"
        fi
    done
    
    echo ""
    echo "Key Logic:"
    grep -E "^type |^func " internal/ws/*.go 2>/dev/null | grep -v "_test"
fi

echo ""
echo ""
echo "2. MIGRATION STATUS"
echo ""
echo ""

echo ""
echo " Component                                         Status                                  "
echo ""

# Check if WebSocket exists in infrastructure
if [ -d "internal/infrastructure/websocket" ]; then
    echo " WebSocket Hub                                   In infrastructure/websocket/   "
else
    echo " WebSocket Hub                                    internal/ws/ (needs migration) "
fi

echo ""

echo ""
echo ""
echo "3. FILES TO MIGRATE"
echo ""
echo ""

echo "Files needing migration:"
if [ -d "internal/ws" ]; then
    for f in internal/ws/*.go; do
        if [ -f "$f" ] && ! echo "$f" | grep -q "_test"; then
            echo "    $(basename $f)"
        fi
    done
fi

echo ""
echo ""
echo "4. KEY LOGIC SUMMARY"
echo ""
echo ""

echo "Hub functionality:"
grep "^type " internal/ws/hub.go 2>/dev/null
grep "^func " internal/ws/hub.go 2>/dev/null | head -10

echo ""
echo "Client functionality:"
grep "^type " internal/ws/client.go 2>/dev/null
grep "^func " internal/ws/client.go 2>/dev/null | head -10

echo ""
echo ""
