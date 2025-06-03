#!/bin/bash

# Ensure cleanup on exit
trap 'echo "Stopping servers..."; kill $(jobs -p) 2>/dev/null; exit' SIGINT SIGTERM

echo "Starting Go backend server on http://localhost:5888 ..."
SERVER_PORT=5888 DATABASE_PATH=./community_watch.db DEV_MODE=true JWT_SECRET=dev-jwt-secret go run ./cmd/server/main.go &
GO_PID=$!
# echo "Go backend server PID: $GO_PID" # Optional: for debugging

# Wait a moment for the Go server to potentially start up before SvelteKit tries to connect via proxy
sleep 2

echo "Starting SvelteKit dev server on http://localhost:5173 (from ./app directory)..."
(cd app && pnpm dev) &
SVELTE_PID=$!
# echo "SvelteKit dev server PID: $SVELTE_PID" # Optional: for debugging

echo
echo "---------------------------------------------------------------------"
echo "Go backend (PID $GO_PID) and SvelteKit dev server (PID $SVELTE_PID) started."
echo "SvelteKit dev (frontend): http://localhost:5173"
echo "Go backend (API):       http://localhost:5888"
echo "Frontend API calls (/api/*) will be proxied from 5173 to 5888."
echo "Press Ctrl+C to stop both servers."
echo "---------------------------------------------------------------------"
echo

# Wait for any of the background jobs to exit
# If one exits (e.g., crashes), the script will then exit, triggering the trap for cleanup.
wait -n
# Fallback wait in case -n isn't supported or one process exits very quickly
wait $GO_PID 2>/dev/null
wait $SVELTE_PID 2>/dev/null