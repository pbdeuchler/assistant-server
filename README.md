# Assistant Server

A comprehensive personal assistant web service that provides both REST API and Model Context Protocol (MCP) interfaces for managing todos, notes, recipes, user preferences, and household data.

## Features

### Core Functionality

- **Todo Management**: Create, list, update, and complete tasks with priority levels and due dates
- **Notes System**: Save and retrieve structured notes with key-based lookup
- **Recipe Management**: Store and search recipes with detailed metadata (prep time, difficulty, ratings)
- **User Preferences**: Flexible key-value preference storage system
- **Household Management**: Support for multi-user households with shared data
- **User Authentication**: OAuth integration with Google for secure authentication

### Dual Interface Support

- **REST API**: Traditional HTTP endpoints for all features
- **MCP Server**: Model Context Protocol implementation for AI assistant integration

## Architecture

```
assistant-server/
├── cmd/                    # Application configuration and server setup
├── dao/postgres/           # PostgreSQL data access layer
├── service/                # HTTP handlers and business logic
├── integration_test/       # Comprehensive integration tests
├── migrations/             # Database schema migrations
└── mocks/                  # Mock implementations for testing
```

## Prerequisites

- Go 1.24.3+
- PostgreSQL 16+
- Docker and Docker Compose (for testing)

## Installation

1. Clone the repository:

```bash
git clone https://github.com/pbdeuchler/assistant-server.git
cd assistant-server
```

2. Install dependencies:

```bash
go mod download
```

3. Set up environment variables:

```bash
export DATABASE_URL="postgres://username:password@localhost:5432/assistant?sslmode=disable"
export PORT="8080"
export BASE_URL="http://localhost:8080"

# Optional: For Google OAuth
export GCLOUD_CLIENT_ID="your-client-id"
export GCLOUD_CLIENT_SECRET="your-client-secret"
export GCLOUD_PROJECT_ID="your-project-id"
```

4. Run database migrations:

```bash
# Migrations will be automatically applied on server startup
```

5. Start the server:

```bash
go run main.go
```

## API Documentation

### REST API Endpoints

#### Todos

- `GET /todos` - List todos with optional filters
- `POST /todos` - Create a new todo
- `GET /todos/{id}` - Get a specific todo
- `PUT /todos/{id}` - Update a todo
- `DELETE /todos/{id}` - Delete a todo

#### Notes

- `GET /notes` - List notes with optional filters
- `POST /notes` - Create a new note
- `GET /notes/{id}` - Get a specific note
- `PUT /notes/{id}` - Update a note
- `DELETE /notes/{id}` - Delete a note

#### Recipes

- `GET /recipes` - Search recipes with filters
- `POST /recipes` - Create a new recipe
- `GET /recipes/{id}` - Get a specific recipe
- `PUT /recipes/{id}` - Update a recipe
- `DELETE /recipes/{id}` - Delete a recipe

#### Preferences

- `GET /preferences` - List preferences
- `POST /preferences` - Set a preference
- `GET /preferences/{key}/{specifier}` - Get a specific preference
- `DELETE /preferences/{key}/{specifier}` - Delete a preference

#### Bootstrap

- `GET /bootstrap` - Get initial data for all entities

#### Authentication

- `GET /oauth/login` - Initiate OAuth flow
- `GET /oauth/callback` - OAuth callback handler

### MCP Tools

The server implements 13 MCP tools for AI assistant integration:

#### Todo Tools

- `create_todo` - Create a new todo task
- `list_todos` - List todos with optional filtering
- `complete_todo` - Mark a todo as completed

#### Note Tools

- `save_note` - Save a note with a key for later retrieval
- `recall_note` - Retrieve a saved note by key
- `list_notes` - List notes with optional filtering

#### Recipe Tools

- `save_recipe` - Save a recipe with metadata
- `find_recipes` - Search recipes by criteria
- `get_recipe` - Get a specific recipe by ID

#### Preference Tools

- `set_preference` - Set a user preference
- `get_preference` - Get a user preference

#### User/Household Tools

- `update_user_description` - Update a user's description
- `update_household_description` - Update a household's description

## Configuration

Environment variables:

- `PORT` - Server port (default: 8080)
- `DATABASE_URL` - PostgreSQL connection string (required)
- `BASE_URL` - Base URL for OAuth callbacks (default: http://localhost:8080)
- `GCLOUD_CLIENT_ID` - Google OAuth client ID (optional)
- `GCLOUD_CLIENT_SECRET` - Google OAuth client secret (optional)
- `GCLOUD_PROJECT_ID` - Google Cloud project ID (optional)

## Testing

### Run All Tests

```bash
make test-integration
```

### Run Specific Test Suites

```bash
# Unit tests
go test ./...

# Integration tests
cd integration_test
go test -v .

# MCP tests only
go test -v . -run TestMCP_

# HTTP API tests only
go test -v . -run TestAPI_
```

### Test Database

Tests use a Docker PostgreSQL instance on port 5433 with:

- Database: `assistant_test`
- User: `test_user`
- Password: `test_password`

## Database Schema

The application uses PostgreSQL with the following main tables:

- `users` - User accounts with OAuth integration
- `households` - Household groups for shared data
- `todos` - Task management
- `notes` - Structured note storage
- `recipes` - Recipe storage with metadata
- `preferences` - Key-value preference storage
- `credentials` - OAuth credential storage

All tables use UUIDs for primary keys and include proper foreign key relationships for data integrity.

## Development

### Building

```bash
go build -o assistant-server
```

### Running with Docker

```bash
# Start test database
make docker-up

# Run server
go run main.go

# Stop test database
make docker-down
```

### Code Generation

```bash
# Generate mocks for testing
go generate ./...
```

## MCP Integration

To use this server with an AI assistant that supports MCP:

1. Start the server
2. Configure your AI assistant to connect to the MCP endpoint at `/mcp`
3. The server will handle protocol negotiation and tool registration automatically

## License

This project is licensed under the GNU GPLv3 License with the [Commons Clause License Condition v1.0](https://commonsclause.com/).

## Contributing

[Add contribution guidelines here]
