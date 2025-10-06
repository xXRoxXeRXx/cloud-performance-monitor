#!/bin/bash

# Test script for MagentaCLOUD debugging with detailed logs

echo "Setting up environment for MagentaCLOUD debugging..."

# Ensure .env exists
if [ ! -f .env ]; then
    echo "Creating .env from example..."
    cp .env.example .env
fi

# Set debug logging
echo "Setting LOG_LEVEL=DEBUG in .env..."
sed -i 's/LOG_LEVEL=.*/LOG_LEVEL=DEBUG/' .env
grep -q "LOG_LEVEL=DEBUG" .env || echo "LOG_LEVEL=DEBUG" >> .env

# Build and start the container
echo "Building and starting the monitor with debug logging..."
docker compose down
docker compose build monitor-agent
docker compose up -d monitor-agent

echo "Waiting for service to start..."
sleep 5

echo "Tailing logs with detailed debugging..."
echo "Look for the following in the logs:"
echo "  - [DEBUG] [http_request] entries showing exact HTTP requests"
echo "  - [DEBUG] [http_response] entries showing server responses"
echo "  - [INFO] [cleanup] entries showing pre-upload cleanup"
echo "  - [INFO] [debug] entries during 409 conflicts with directory listings"
echo "  - [WARN] [cleanup] entries if old upload directories are found"
echo ""

# Follow logs until interrupted
docker compose logs -f monitor-agent