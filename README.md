# Personal Health Monitor

A personal health monitoring application built with Go, HTMX, and DaisyUI. Track weekly health metrics with visual gauges and historical trends.

## Features
- Weekly health metrics dashboard with target-based gauges
- Admin interface for managing metrics and targets
- Historical trends visualization (monthly and yearly)
- Visual indicators for above/below target metrics

## Tech Stack
- Backend: Go with Chi router
- Database: Turso DB with Atlas migrations
- Frontend: HTMX + DaisyUI
- Data Visualization: Chart.js

## Project Structure
```
health-monitor/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── db/
│   ├── handlers/
│   ├── models/
│   └── templates/
├── migrations/
├── static/
│   ├── css/
│   └── js/
└── templates/
```

## Setup and Running
1. Install dependencies
2. Set up Turso DB
3. Run migrations
4. Start the server: `go run cmd/server/main.go`
