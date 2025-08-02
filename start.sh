#!/bin/bash
set -e

echo "🚀 Starting health-monitor server..."
echo "📋 Deployment Information:"
echo "  Branch: ${RAILWAY_GIT_BRANCH:-unknown}"
echo "  Commit: ${RAILWAY_GIT_COMMIT_SHA:-unknown}"
echo "  Environment: ${RAILWAY_ENVIRONMENT:-unknown}"

# Create data directory if needed
mkdir -p /data 2>/dev/null || echo "⚠️ Warning: Could not create /data directory"

# Database seeding for development branches only
CURRENT_BRANCH=${RAILWAY_GIT_BRANCH:-unknown}
echo "🌱 Checking if database seeding is needed..."
echo "  Current branch: $CURRENT_BRANCH"

if [ "$CURRENT_BRANCH" != "main" ] && [ "$CURRENT_BRANCH" != "master" ] && [ "$CURRENT_BRANCH" != "unknown" ] && [ "$CURRENT_BRANCH" != "" ]; then
    echo "  🌱 Seeding database for development branch: $CURRENT_BRANCH"
    echo "  📍 Database path: ${DB_PATH:-/data/health-monitor.db}"
    ./bin/seed -db "${DB_PATH:-/data/health-monitor.db}" || echo "  ⚠️ Seeding failed - continuing without seed data"
else
    echo "  ⏭️ Skipping seed - production branch or unknown: $CURRENT_BRANCH"
fi

echo "🎯 Starting server binary..."
exec ./bin/server