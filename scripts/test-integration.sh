#!/bin/bash

set -e

echo "Starting integration tests..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "Error: Docker is not running. Please start Docker and try again."
    exit 1
fi

# Stop any existing test containers
echo "Cleaning up any existing test containers..."
docker-compose -f docker-compose.test.yml down -v > /dev/null 2>&1 || true

# Start test database
echo "Starting test database..."
docker-compose -f docker-compose.test.yml up -d

# Wait for database to be ready
echo "Waiting for database to be ready..."
timeout=30
while ! docker-compose -f docker-compose.test.yml exec -T postgres-test pg_isready -U test_user -d assistant_test > /dev/null 2>&1; do
    timeout=$((timeout - 1))
    if [ $timeout -eq 0 ]; then
        echo "Error: Database did not become ready in time"
        docker-compose -f docker-compose.test.yml logs postgres-test
        docker-compose -f docker-compose.test.yml down -v
        exit 1
    fi
    sleep 1
done

echo "Database is ready!"

# Set test environment variables
export TEST_DB_URL="postgres://test_user:test_password@localhost:5433/assistant_test?sslmode=disable"

# Run integration tests
echo "Running integration tests..."
cd integration_test && go test -v . -timeout=5m

# Get test exit code
test_exit_code=$?

# Clean up
echo "Cleaning up test containers..."
docker-compose -f docker-compose.test.yml down -v

if [ $test_exit_code -eq 0 ]; then
    echo "✅ All integration tests passed!"
else
    echo "❌ Integration tests failed!"
    exit $test_exit_code
fi