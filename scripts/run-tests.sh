#!/bin/bash

# Import shared utilities and colors
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Check required commands
check_command() {
    command -v $1 >/dev/null 2>&1 || { echo "Error: $1 is required but not installed. $2"; exit 1; }
}

check_command go "Install Go from https://golang.org/doc/install"
check_command npm "Install Node.js and npm from https://nodejs.org/"

# Print header
echo "------------------------------------------"
echo "   📝 Running tests on all components  "
echo "------------------------------------------"

# Start timing
echo "Starting tests at $(date)"
start_time=$(date +%s)
overall_exit_code=0

# Run core package tests
echo -e "\nRunning tests on pkg kernel..."
go test -v ./pkg || overall_exit_code=1

echo -e "\nRunning component tests..."

# Test manager component
echo -e "\nTesting manager component..."
if [ -d "components/manager" ]; then
    (cd components/manager && make test) || {
        overall_exit_code=1
        echo "[error] Manager component tests failed."
    }
fi

# Test worker component
echo -e "\nTesting worker component..."
if [ -d "components/worker" ]; then
    (cd components/worker && make test) || {
        overall_exit_code=1
        echo "[error] Worker component tests failed."
    }
fi

# Calculate duration and print summary
end_time=$(date +%s)
duration=$((end_time - start_time))
echo -e "\nTest Summary:"
echo "----------------------------------------"
echo "Duration: $(printf '%dm:%02ds' $((duration / 60)) $((duration % 60)))"
echo "----------------------------------------"

# Print final status and exit with appropriate code
if [ $overall_exit_code -eq 0 ]; then
    echo "[ok] All tests passed successfully ✔️"
else
    echo "[error] Some tests failed. Please check the output above for details."
fi

exit $overall_exit_code
