#!/bin/bash

# Kill any existing processes
echo "Killing existing processes..."
lsof -i :3000 | awk 'NR!=1 {print $2}' | xargs kill -9 2>/dev/null || true

# Clean temporary directories
echo "Cleaning temp directories..."
rm -rf tmp/

# Regenerate templates
echo "Regenerating templates..."
templ generate

# Start server
echo "Starting server with Air..."
air
