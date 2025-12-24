#!/bin/bash

# AutoZap Demo CLI Helper
# Runs CLI commands with the correct database path

DB_PATH="./autozap.db"

case "$1" in
    history)
        shift
        ./autozap history --db "$DB_PATH" "$@"
        ;;
    stats)
        shift
        ./autozap stats "$@" --db "$DB_PATH"
        ;;
    failures)
        shift
        ./autozap failures --db "$DB_PATH" "$@"
        ;;
    *)
        echo "Usage: $0 {history|stats|failures} [options]"
        echo ""
        echo "Examples:"
        echo "  $0 history --limit 10"
        echo "  $0 stats demo-quick"
        echo "  $0 failures --hours 24"
        exit 1
        ;;
esac
