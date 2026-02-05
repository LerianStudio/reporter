#!/bin/bash

# Clean build artifacts script

BIN_DIR="./.bin"
ARTIFACTS_DIR="./artifacts"

echo "Cleaning build artifacts..."

# Safety checks
if [ -z "$BIN_DIR" ] || [ -z "$ARTIFACTS_DIR" ]; then
    echo "[error] BIN_DIR or ARTIFACTS_DIR is not set. Aborting to prevent accidental deletion."
    exit 1
fi

if [ "$BIN_DIR" = "/" ] || [ "$ARTIFACTS_DIR" = "/" ]; then
    echo "[error] BIN_DIR or ARTIFACTS_DIR cannot be root directory. Aborting."
    exit 1
fi

# Clean bin directory
if [ -d "$BIN_DIR" ]; then
    echo "Cleaning $BIN_DIR..."
    rm -rf "$BIN_DIR"/*
fi

# Clean artifacts directory
if [ -d "$ARTIFACTS_DIR" ]; then
    echo "Cleaning $ARTIFACTS_DIR..."
    rm -rf "$ARTIFACTS_DIR"/*
fi

# Clean coverage files
if [ -f "coverage.out" ]; then
    echo "Cleaning coverage.out..."
    rm -f coverage.out
fi

if [ -f "coverage.html" ]; then
    echo "Cleaning coverage.html..."
    rm -f coverage.html
fi

echo "[ok] Artifacts cleaned successfully"
