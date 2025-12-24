#!/bin/bash

# AutoZap Demo Script
# Works on both Linux and macOS
# Builds AutoZap and runs sample workflows to populate the dashboard

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored messages
print_info() {
    echo -e "${BLUE}ℹ ${NC}$1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_header() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.21 or higher."
        exit 1
    fi

    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_success "Found Go version: $GO_VERSION"
}

# Build AutoZap
build_autozap() {
    print_info "Building AutoZap..."

    if go build -mod=mod -o autozap .; then
        print_success "AutoZap built successfully"
    else
        print_error "Failed to build AutoZap"
        exit 1
    fi
}

# Clean up old data
cleanup_old_data() {
    print_info "Cleaning up old demo data..."

    if [ -f "autozap.db" ]; then
        rm -f autozap.db
        print_success "Removed old database"
    fi

    if [ -f "autozap.pid" ]; then
        OLD_PID=$(cat autozap.pid)
        if kill -0 "$OLD_PID" 2>/dev/null; then
            kill "$OLD_PID" 2>/dev/null || true
            sleep 2
            print_success "Stopped old AutoZap process"
        fi
        rm -f autozap.pid
    fi
}

# Start AutoZap in agent mode
start_autozap() {
    print_info "Starting AutoZap in agent mode..."

    # Start in background and save PID
    nohup ./autozap agent ./workflows \
        --db ./autozap.db \
        --http-port 8080 \
        > autozap.log 2>&1 &

    echo $! > autozap.pid
    print_success "AutoZap started (PID: $(cat autozap.pid))"

    # Wait for server to be ready
    print_info "Waiting for server to be ready..."
    for i in {1..30}; do
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            print_success "Server is ready!"
            return 0
        fi
        sleep 1
    done

    print_error "Server failed to start within 30 seconds"
    print_info "Check autozap.log for details:"
    tail -n 20 autozap.log
    exit 1
}

# Display dashboard info
show_dashboard_info() {
    print_header "AutoZap Demo Running"

    echo -e "${GREEN}Dashboard URLs:${NC}"
    echo -e "  🎨 Dashboard:    ${BLUE}http://localhost:8080/dashboard${NC}"
    echo -e "  📊 Metrics:      ${BLUE}http://localhost:8080/metrics${NC}"
    echo -e "  ❤️  Health:       ${BLUE}http://localhost:8080/health${NC}"
    echo -e "  📈 Status:       ${BLUE}http://localhost:8080/status${NC}"
    echo ""

    echo -e "${GREEN}API Endpoints:${NC}"
    echo -e "  📋 Active Workflows: ${BLUE}http://localhost:8080/api/workflows/active${NC}"
    echo -e "  📜 History:          ${BLUE}http://localhost:8080/api/workflows/history${NC}"
    echo -e "  📊 Stats:            ${BLUE}http://localhost:8080/api/workflows/stats${NC}"
    echo -e "  ⚠️  Failures:         ${BLUE}http://localhost:8080/api/workflows/failures${NC}"
    echo ""

    echo -e "${GREEN}Loaded Workflows:${NC}"
    sleep 2
    WORKFLOW_COUNT=$(curl -s http://localhost:8080/api/workflows/active | grep -o '"name"' | wc -l | tr -d ' ')
    echo -e "  Total: ${YELLOW}${WORKFLOW_COUNT}${NC} workflows"
    echo ""

    echo -e "${YELLOW}Quick Demo Workflow:${NC}"
    echo -e "  • 'demo-quick' runs every ${GREEN}1 minute${NC}"
    echo -e "  • Watch the dashboard update in real-time!"
    echo -e "  • Success rates and execution counts will populate quickly"
    echo ""

    echo -e "${GREEN}Command Line Tools:${NC}"
    echo -e "  ${BLUE}./demo-cli.sh history${NC}          # View execution history"
    echo -e "  ${BLUE}./demo-cli.sh stats demo-quick${NC} # View workflow statistics"
    echo -e "  ${BLUE}./demo-cli.sh failures${NC}         # View recent failures"
    echo ""

    echo -e "${YELLOW}Logs:${NC}"
    echo -e "  ${BLUE}tail -f autozap.log${NC}        # Watch live logs"
    echo ""

    print_info "To stop AutoZap: ${BLUE}kill \$(cat autozap.pid)${NC}"
    echo ""
}

# Try to open browser
open_browser() {
    print_info "Attempting to open dashboard in browser..."

    # Detect OS and open browser
    case "$(uname -s)" in
        Darwin*)
            open http://localhost:8080/dashboard 2>/dev/null || true
            ;;
        Linux*)
            if command -v xdg-open &> /dev/null; then
                xdg-open http://localhost:8080/dashboard 2>/dev/null || true
            elif command -v gnome-open &> /dev/null; then
                gnome-open http://localhost:8080/dashboard 2>/dev/null || true
            fi
            ;;
    esac
}

# Main execution
main() {
    print_header "AutoZap Demo Setup"

    check_go
    build_autozap
    cleanup_old_data
    start_autozap
    show_dashboard_info
    open_browser

    print_success "Demo is now running!"
    print_warning "The demo-quick workflow will execute every minute."
    print_info "Watch the dashboard at: ${BLUE}http://localhost:8080/dashboard${NC}"
    echo ""
}

# Run main function
main
