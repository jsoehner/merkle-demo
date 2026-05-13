#!/bin/bash
set -e

echo "🔄 Generating fresh MTC artifacts on startup..."
cd /app/demo
./setup.sh

echo ""
echo "🎮 Launching MTC Playground server on https://localhost:8444..."
cd /app/playground
PLAYGROUND_BIN_DIR=/usr/local/bin ./playground-server &

echo ""
echo "🚀 Launching MTC Demo website on :8443..."
cd /app/demo
exec ./website-server
