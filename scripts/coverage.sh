#!/bin/bash

echo "Generating test coverage report..."

PACKAGES=$(go list ./... | grep -v -f ./scripts/coverage_ignore.txt)

echo "Running tests on packages:"
echo "$PACKAGES"

go test -cover $PACKAGES -coverprofile=coverage.raw.out

# Filter out mock files and datasource-config from coverage report
echo "Filtering mock files and connection files from coverage..."
grep -v -E "mock\.go|datasource-config\.go" coverage.raw.out > coverage.out
rm coverage.raw.out

printf "\nCoverage Summary:\n"
go tool cover -func=coverage.out

printf "\nGenerating HTML coverage report...\n"
go tool cover -html=coverage.out -o coverage.html
echo "HTML coverage report generated at: coverage.html"
