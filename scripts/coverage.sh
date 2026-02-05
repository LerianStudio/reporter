#!/bin/bash

echo "Generating test coverage report..."

# Get the list of packages to test, excluding those in the ignore list
PACKAGES=$(go list ./pkg/... ./components/... | grep -v -f ./scripts/coverage_ignore.txt)

echo "Running tests on packages:"
echo "$PACKAGES"

# Run the tests and generate coverage profile
go test -cover $PACKAGES -coverprofile=coverage.out
