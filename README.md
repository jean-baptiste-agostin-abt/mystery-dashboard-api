# Mystery Factory API

Mystery Factory API is a modern, multi-tenant video management and publishing platform built with Go and the Gin framework. It provides AI-powered content generation, comprehensive analytics, and seamless multi-platform publishing capabilities.

## Key Features

- **Multi-tenant Architecture**: Secure tenant isolation with role-based access control
- **Video Management**: Upload, process, and manage video content with GORM-based persistence
- **AI Magic Brush**: Real-time AI content generation using AWS Bedrock Claude 4
- **Prompt Catalog**: Centralized YAML-based prompt management system
- **Analytics & Statistics**: Comprehensive performance tracking with ROI and engagement insights
- **Prometheus Monitoring**: Full observability with metrics, tracing, and dashboards
- **Service Architecture**: Clean interfaces with dependency injection and testability
- **Multi-platform Publishing**: Automated publishing to YouTube, TikTok, Instagram, and more
- **Real-time Processing**: Event-driven architecture with immediate AI responses

## Technology Stack

- **Backend**: Go 1.24+ with Gin framework
- **Database**: MySQL 8.0+ with GORM ORM
- **AI Processing**: AWS Bedrock with Claude 4
- **Monitoring**: Prometheus + Grafana + Jaeger
- **Authentication**: JWT-based with RBAC
- **Caching**: Redis for performance optimization
- **Documentation**: Swagger/OpenAPI 3.0
- **Testing**: Unit & integration tests with mocking
- **Containerization**: Docker & Docker Compose

## Architecture Overview

### Service Layer Architecture

The application follows a clean architecture pattern with well-defined service interfaces:

```
internal/
├── services/
│   ├── interfaces.go          # Service interface definitions
│   ├── ai_service.go          # AI processing with AWS Bedrock
│   ├── video_service.go       # Video management operations
│   ├── analytics_service.go   # Statistics and analytics
│   ├── campaign_service.go    # Campaign management
│   └── prompt_service.go      # Prompt catalog management
├── handlers/                  # HTTP request handlers
├── models/                    # GORM data models
└── middleware/                # HTTP middleware
```

### Key Service Interfaces

- **AIService**: Real-time AI content generation using AWS Bedrock Claude 4
- **VideoService**: Video CRUD operations, upload handling, and publishing
- **AnalyticsService**: Performance metrics, ROI analysis, and engagement tracking
- **PromptService**: Centralized prompt catalog management with YAML configuration
- **CampaignService**: Campaign lifecycle management and scheduling

## Prompt Catalog System

The prompt catalog provides centralized management of AI prompts with template variables:

```yaml
# prompts/catalog.yaml
prompts:
  magic_brush/title_gen:
    name: "Video Title Generator"
    template: "Generate 5 compelling titles for {{topic}} targeting {{audience}}..."
    variables:
      - name: "topic"
        type: "string"
        required: true
      - name: "audience"
        type: "string"
        default: "general audience"
```

### Prompt Management Features

- **Template Variables**: Dynamic prompt rendering with validation
- **Hot Reloading**: Automatic catalog updates during development
- **Version Control**: Track prompt changes and performance
- **Testing**: Built-in prompt testing with mock data

## AI Processing with AWS Bedrock

All AI features use AWS Bedrock with Claude 4 for content generation:

```http
POST /api/v1/ai/magic-brush
Content-Type: application/json

{
  "video_id": "video_123",
  "brush_type": "title",
  "context": {
    "topic": "Go programming tutorial",
    "platform": "youtube",
    "audience": "developers"
  }
}
```

### AI Features

- **Magic Brush**: Real-time title, description, and tag generation
- **Prompt Testing**: Test prompts with custom data
- **Token Tracking**: Monitor usage and costs
- **Error Handling**: Comprehensive retry logic and fallbacks

## Monitoring and Observability

### Prometheus Metrics

The application exposes comprehensive metrics at `/metrics`:

- **HTTP Metrics**: Request counts, duration, status codes
- **AI Metrics**: Processing time, token usage, success rates
- **Database Metrics**: Query performance, connection pool status
- **Business Metrics**: Video counts, campaign success rates

### Monitoring Stack

- **Prometheus**: Metrics collection and alerting
- **Grafana**: Dashboards and visualization
- **Jaeger**: Distributed tracing
- **Health Checks**: Application and dependency health

## Quick Start

### Prerequisites

- Go 1.24 or higher
- Docker and Docker Compose
- AWS credentials for Bedrock access
- MySQL 8.0+

### Using Makefile (Recommended)

1. Clone and setup:
```bash
git clone <repository-url>
cd mystery-dashboard-api
make setup          # Install development tools
make env-example    # Create .env.example
```

2. Configure environment:
```bash
cp .env.example .env
# Edit .env with your AWS credentials and database settings
```

3. Start development environment:
```bash
make docker-compose-up    # Start all services
make migrate-up          # Run database migrations
make dev                 # Start development server with hot reload
```

### Manual Installation

1. Install dependencies:
```bash
make deps
```

2. Start services:
```bash
docker-compose up -d mysql redis prometheus grafana jaeger
```

3. Run migrations:
```bash
make migrate-up
```

4. Start the application:
```bash
make run
```

The API will be available at `http://localhost:8080`

### Monitoring URLs

- **API**: http://localhost:8080
- **Swagger Docs**: http://localhost:8080/swagger/index.html
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **Jaeger**: http://localhost:16686

## API Documentation

Interactive API documentation is available at:
- Development: `http://localhost:8080/swagger/index.html`

### Key Endpoints

#### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration
- `GET /api/v1/auth/me` - Get current user profile
- `PUT /api/v1/auth/me` - Update user profile
- `POST /api/v1/auth/change-password` - Change password

#### Video Management
- `GET /api/v1/videos` - List videos with pagination
- `POST /api/v1/videos` - Create video metadata
- `GET /api/v1/videos/{id}` - Get video details
- `PUT /api/v1/videos/{id}` - Update video metadata
- `DELETE /api/v1/videos/{id}` - Delete video
- `POST /api/v1/videos/{id}/upload` - Upload video file
- `POST /api/v1/videos/{id}/publish` - Publish video to platforms

#### AI Magic Brush
- `POST /api/v1/ai/magic-brush` - Generate titles, descriptions, or tags
- `GET /api/v1/ai/prompts` - List available prompts
- `POST /api/v1/ai/test-prompt` - Test prompt with custom data

#### Analytics & Statistics
- `GET /api/v1/stats/dashboard` - Dashboard overview
- `GET /api/v1/stats/videos` - Video performance statistics
- `GET /api/v1/stats/videos/{id}` - Individual video statistics
- `GET /api/v1/stats/roi` - ROI analytics and financial performance
- `GET /api/v1/stats/engagement` - Engagement metrics and audience insights
- `POST /api/v1/stats/sync` - Sync statistics from platforms

#### Platform Integration
- `POST /api/v1/platforms/webhook/{platform}` - Platform webhook handler
- `GET /api/v1/platforms/{platform}/auth` - Initiate platform authentication
- `POST /api/v1/platforms/{platform}/auth/callback` - Handle auth callback

#### Monitoring
- `GET /health` - Application health check
- `GET /ready` - Readiness check
- `GET /metrics` - Prometheus metrics

## Development

### Using the Makefile

The project includes a comprehensive Makefile with all common development tasks:

```bash
# Setup development environment
make setup

# Install dependencies
make deps

# Run tests
make test                # Full test suite with coverage
make test-short         # Quick tests without race detection
make test-integration   # Integration tests

# Code quality
make lint               # Run linters
make format            # Format code
make vet               # Run go vet
make security          # Security checks

# Database operations
make migrate-up        # Apply migrations
make migrate-down      # Rollback migration
make migrate-create NAME=migration_name  # Create new migration

# Development server
make dev               # Hot reload development server
make run               # Build and run locally

# Docker operations
make docker-build      # Build Docker image
make docker-run        # Run in container
make docker-compose-up # Start all services

# Monitoring
make metrics           # Start monitoring stack
make metrics-down      # Stop monitoring stack

# Utilities
make clean             # Clean build artifacts
make help              # Show all available targets
```

### Testing Strategy

```bash
# Run all tests with coverage
make test

# Run specific test packages
go test ./internal/services/... -v

# Run tests with race detection
go test -race ./...

# Generate mocks for testing
make generate
```

### Code Quality Standards

```bash
# Comprehensive code quality check
make check  # Runs lint + test

# Individual quality checks
make lint     # golangci-lint + go vet + gofmt
make format   # Auto-format code
make security # gosec security analysis
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.