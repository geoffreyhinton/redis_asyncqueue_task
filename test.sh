#!/bin/bash

echo "🔧 Redis Async Queue Test Script"
echo "================================"

# Check if Redis is running
echo "📡 Checking Redis connection..."
if ! redis-cli ping > /dev/null 2>&1; then
    echo "❌ Redis is not running!"
    echo "Please start Redis first:"
    echo "  brew services start redis"
    echo "  OR"
    echo "  redis-server"
    exit 1
else
    echo "✅ Redis is running"
fi

# Clean up any existing data
echo "🧹 Cleaning up Redis data..."
redis-cli FLUSHALL > /dev/null

# Build the test program
echo "🔨 Building test program..."
cd examples
if go build main.go; then
    echo "✅ Build successful"
else
    echo "❌ Build failed"
    exit 1
fi

# Run the test
echo "🚀 Running async queue test..."
echo "Press Ctrl+C to stop the test"
echo "================================"
./main