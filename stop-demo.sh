#!/bin/bash

# Stop AutoZap Demo Script

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() {
    echo -e "${BLUE}ℹ ${NC}$1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

if [ ! -f "autozap.pid" ]; then
    print_error "No autozap.pid file found. AutoZap may not be running."
    exit 1
fi

PID=$(cat autozap.pid)

if ! kill -0 "$PID" 2>/dev/null; then
    print_error "Process $PID is not running."
    rm -f autozap.pid
    exit 1
fi

print_info "Stopping AutoZap (PID: $PID)..."
kill "$PID"

# Wait for process to stop
for i in {1..10}; do
    if ! kill -0 "$PID" 2>/dev/null; then
        print_success "AutoZap stopped successfully"
        rm -f autozap.pid
        exit 0
    fi
    sleep 1
done

# Force kill if still running
print_info "Force stopping..."
kill -9 "$PID" 2>/dev/null || true
rm -f autozap.pid
print_success "AutoZap stopped"
