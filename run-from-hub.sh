#!/bin/bash
set -e

# Default Docker Hub username
export DOCKER_USER="${1:-jsoehner}"
export TAG="${2:-latest}"

echo "=========================================================================="
echo "🚀 Running MTC Demo & Playground from Docker Hub"
echo "   Docker Hub User : $DOCKER_USER"
echo "   Image Tag       : $TAG"
echo "=========================================================================="

# 1. Pull the latest images from Docker Hub
echo "📥 Pulling latest images..."
docker compose -f docker-compose.prod.yml pull

# 2. Stop and clean any existing volumes/containers
echo "🧹 Cleaning up old containers..."
docker compose -f docker-compose.prod.yml down -v --remove-orphans

# 3. Start containers in detached mode
echo "⚙️ Starting containers..."
docker compose -f docker-compose.prod.yml up -d

# 4. Wait for Demo to become healthy
echo "⏳ Waiting for mtc-demo-website to initialize and become healthy..."
docker compose -f docker-compose.prod.yml exec -T demo /bin/bash -c "echo 'Container is up.'" >/dev/null 2>&1 || true

# Check health loop
for i in {1..30}; do
  STATUS=$(docker inspect -f '{{.State.Health.Status}}' mtc-demo-website 2>/dev/null || echo "starting")
  if [ "$STATUS" = "healthy" ]; then
    echo "✅ mtc-demo-website is healthy!"
    break
  fi
  if [ "$STATUS" = "unhealthy" ]; then
    echo "❌ mtc-demo-website is unhealthy. Check logs with: docker logs mtc-demo-website"
    exit 1
  fi
  sleep 2
done

echo "=========================================================================="
echo "🎉 Setup complete! The containers are running successfully."
echo "   • MTC Demo Website : https://localhost:8443"
echo "   • MTC Playground   : https://localhost:8444"
echo "=========================================================================="
