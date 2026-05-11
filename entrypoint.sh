#!/bin/bash
set -e

echo "🔄 Generating fresh MTC artifacts on startup..."
cd /app/demo
./setup.sh

echo ""
echo "🚀 Launching website server..."
exec ./website-server
