#!/bin/bash

# Import shared utilities and colors
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Print header
echo "---------------------------------------"
echo "   📝 Cleaning all build artifacts  "
echo "---------------------------------------"

# Define components
COMPONENTS=$(find "${PROJECT_ROOT:?}/components" -mindepth 1 -maxdepth 1 -type d | sort)

# Clean each component
for dir in $COMPONENTS; do
    component_name=$(basename "$dir")
    echo "Cleaning in ./$component_name..."

    # Run component's clean target if Makefile exists
    if [ -f "$dir/Makefile" ]; then
        (cd "$dir" && make clean) || true
    fi

    echo "Ensuring thorough cleanup in ./$component_name..."

    # Remove common build artifacts
    (cd "$dir" &&
        for item in bin dist coverage.out coverage.html artifacts *.tmp node_modules; do
            if [ -e "$item" ]; then
                echo "Removing ./$component_name/$item"
                rm -rf "$item"
            fi
        done
    ) || true
done

# Clean root-level build artifacts
echo "Cleaning root-level build artifacts..."
for item in bin dist coverage.out coverage.html *.tmp node_modules .bin artifacts; do
    if [ -e "${PROJECT_ROOT:?}/$item" ]; then
        echo "Removing $item"
        rm -rf "${PROJECT_ROOT:?}/$item"
    fi
done

# Deep cleaning
echo "Deep cleaning project..."

echo "Finding and removing coverage.out files..."
find "${PROJECT_ROOT:?}" -name "coverage.out" -type f -delete -print || true

echo "Finding and removing coverage.html files..."
find "${PROJECT_ROOT:?}" -name "coverage.html" -type f -delete -print || true

echo "Finding and removing bin directories..."
find "${PROJECT_ROOT:?}" -name "bin" -type d -prune -exec rm -rf {} \; 2>/dev/null || true

echo "Finding and removing dist directories..."
find "${PROJECT_ROOT:?}" -name "dist" -type d -prune -exec rm -rf {} \; 2>/dev/null || true

echo "Finding and removing node_modules directories..."
find "${PROJECT_ROOT:?}" -name "node_modules" -type d -prune -exec rm -rf {} \; 2>/dev/null || true

echo "[ok] All artifacts cleaned successfully ✔️"
