# Integration Tests

This directory contains comprehensive integration tests for both the HTTP API and MCP server functionality.

## Overview

The integration tests verify end-to-end functionality using a real PostgreSQL database running in Docker. Tests are designed to use shared infrastructure and utilities to maximize code reuse between HTTP API and MCP testing.

## Features Tested

### ✅ MCP Server (FULLY WORKING)
- **Initialization**: Protocol negotiation and capability exchange ✅
- **Tools List**: All 13 MCP tools are properly registered ✅
- **Todo Management**: Create, list, and complete todos via MCP ✅
- **Notes Management**: Save, recall, and list notes via MCP ✅
- **Recipe Management**: Save, search, and retrieve recipes via MCP ✅
- **Preferences**: Set and get user preferences via MCP ✅
- **User/Household Updates**: Update descriptions via MCP ✅
- **Error Handling**: Invalid methods, tools, and parameters ✅

### ✅ Database Infrastructure (FULLY WORKING)  
- **Migrations**: All 16 database migrations execute successfully ✅
- **Schema**: Proper table creation with correct column types and constraints ✅
- **UUID Support**: Proper UUID generation and foreign key relationships ✅
- **Data Cleanup**: Automatic test isolation between runs ✅

### ✅ HTTP API (MOSTLY WORKING)
- **Todos API**: Full CRUD operations working ✅
- **Database Connection**: Successfully connects and applies migrations ✅
- **Endpoint Routing**: All endpoints are properly mounted ✅
- **Request Parsing**: JSON parsing and validation works ✅
- **Timestamp Handling**: Fixed scanning issues with timestamps ✅
- **Remaining Issues**: Some endpoints (notes, recipes, preferences) need similar fixes ⚠️

## Test Structure

```
integration_test/
├── testutil/
│   └── testutil.go          # Shared database setup and test utilities
├── http_api_test.go         # HTTP API integration tests
├── mcp_test.go             # MCP server integration tests
├── migration_test.go       # Database migration verification
└── go.mod                  # Test module dependencies
```

## Running Tests

### Prerequisites
- Docker and Docker Compose
- Go 1.24.3+

### Quick Start
```bash
# Run all integration tests
make test-integration

# Or use the script directly
./scripts/test-integration.sh

# Run specific test suites
cd integration_test
go test -v . -run TestMCP_
go test -v . -run TestMigrations
```

### Manual Database Setup
```bash
# Start test database
make docker-up

# Run tests against existing database
cd integration_test && go test -v .

# Clean up
make docker-down
```

## Test Database

- **Image**: PostgreSQL 16 Alpine
- **Port**: 5433 (to avoid conflicts with local PostgreSQL)
- **Database**: `assistant_test`
- **User**: `test_user` / `test_password`
- **Storage**: In-memory tmpfs for fast test execution

## Shared Test Utilities

The `testutil` package provides:

- **Database Setup**: Automatic migration running and cleanup
- **Test Fixtures**: Helper functions to create test users, households, todos, notes, recipes, and preferences
- **UUID Generation**: Proper UUID generation for database compatibility
- **Assertion Helpers**: Specialized assertion functions for comparing database entities

## Key Accomplishments

1. **MCP Server Verification**: ✅ **CONFIRMED WORKING** - All MCP functionality passes integration tests
2. **Database Migrations**: ✅ All 16 migrations execute successfully, creating proper schema
3. **Todos API Fixed**: ✅ Resolved timestamp scanning issues - full CRUD operations working
4. **Shared Infrastructure**: ✅ Reusable test utilities minimize code duplication
5. **Test Isolation**: ✅ Each test gets a clean database state
6. **Error Coverage**: ✅ Both success and failure scenarios are tested

## Issues Fixed

1. **Timestamp Scanning**: Fixed `UpdateTodo` struct to use `*time.Time` instead of `*string`
2. **SQL Column Ordering**: Replaced `SELECT *` with explicit column lists to ensure consistent scanning
3. **MCP Completion**: Fixed marked_complete field to use `time.Now()` instead of string "true"
4. **UUID Generation**: Added proper UUID generation for all test fixtures
5. **Migration Parser**: Created goose migration parser to extract SQL from migration files

## Remaining Issues

- Some HTTP API endpoints (notes, recipes, preferences) may need similar timestamp fixes
- Some error message assertions need refinement  
- Test execution could be optimized to share database instances

## Architecture Benefits

- **Docker Integration**: No need for developers to install PostgreSQL locally
- **Migration Verification**: Tests ensure database schema matches application expectations  
- **Real Database Testing**: Uses actual PostgreSQL instead of mocks for higher confidence
- **Comprehensive Coverage**: Tests both HTTP and MCP interfaces with shared backend