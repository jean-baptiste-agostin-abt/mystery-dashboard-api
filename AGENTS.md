# Mystery Factory API - Agent Documentation

## Project Context

Mystery Factory API is a multi-tenant video management and publishing platform built with Go and the Gin framework. It enables users to upload, process, and publish videos across multiple social media platforms while providing comprehensive analytics and AI-powered content insights.

### Key Features

- **Multi-tenant Architecture**: Secure tenant isolation with role-based access control
- **Video Management**: Upload, process, and manage video content with metadata
- **Multi-platform Publishing**: Automated publishing to YouTube, TikTok, Instagram, Facebook, Twitter, and LinkedIn
- **AI Campaigns**: Orchestrate parallel AI agents (research, ideation, validation) to generate video ideas
- **AI Magic Brushes**: On-demand prompts to auto-generate titles & descriptions via LLM
- **Analytics & Statistics**: Comprehensive performance tracking and insights (ROI, engagement)
- **Real-time Monitoring**: OpenTelemetry integration with Jaeger tracing
- **Webhook Support**: Platform-specific webhook handling for status updates

### Technology Stack

- **Backend**: Go 1.24+ with Gin web framework
- **Database**: MySQL 8.0 with migration support
- **Authentication**: JWT-based authentication with role-based access control
- **Observability**: OpenTelemetry with Jaeger tracing
- **Documentation**: Swagger/OpenAPI 3.0
- **Testing**: Comprehensive unit and integration tests
- **Containerization**: Docker and Docker Compose for development and testing

## Architecture Overview

### Project Structure

```
mystery-dashboard-api/
├── cmd/server/                 # Application entry point
│   └── main.go                # Server initialization and configuration
├── internal/                  # Private application code
│   ├── config/               # Configuration management
│   ├── handlers/             # HTTP request handlers
│   ├── middleware/           # HTTP middleware (auth, logging, etc.)
│   ├── models/              # Data models and business logic
│   └── router/              # Route definitions and setup
├── pkg/                     # Public packages
│   ├── db/                  # Database connection and utilities
│   └── logger/              # Structured logging wrapper
├── test/                    # Integration tests
├── migrations/              # Database migration files
├── docs/                    # API documentation
└── .junie/                  # AI agent guidelines and standards
```

### Core Components

#### 1. Multi-tenant System
- Tenant isolation at the database and application level
- User management with role-based permissions (admin, editor, viewer, publisher)
- Secure JWT-based authentication with tenant context

#### 2. Video Management
- Video upload and processing pipeline
- Metadata extraction and storage
- Status tracking (uploading, processing, ready, failed)
- File storage integration (S3-compatible)

#### 3. Publication System
- Multi-platform publishing jobs with scheduling
- Platform-specific configuration and authentication
- Retry mechanisms and error handling
- Webhook integration for status updates

#### 4. AI Processing
- Batch processing system for AI operations
- Support for video analysis, content generation, thumbnail creation
- Cost estimation and tracking
- Progress monitoring and error handling

#### 5. Analytics & Statistics
- Real-time performance metrics
- Historical data tracking with snapshots
- Cross-platform aggregation
- Demographic and engagement analytics

## Development Guidelines

### Code Standards

1. **Language**: All code, comments, and documentation must be in English
2. **Naming Conventions**: Use camelCase for variables/functions, PascalCase for types/structs
3. **Error Handling**: Always handle errors explicitly, use structured error responses
4. **Logging**: Use structured logging with appropriate log levels and context
5. **Testing**: Maintain high test coverage with both unit and integration tests

### API Design Principles

1. **RESTful Design**: Follow REST conventions for resource naming and HTTP methods
2. **Consistent Responses**: Use standardized response formats for success and error cases
3. **Versioning**: API versioning through URL path (`/api/v1/`)
4. **Authentication**: JWT-based authentication for all protected endpoints
5. **Validation**: Comprehensive input validation with clear error messages

### Database Guidelines

1. **Migrations**: All schema changes must be done through migration files
2. **Soft Deletes**: Use soft deletes for important entities (users, videos, etc.)
3. **Indexing**: Proper indexing for performance-critical queries
4. **Transactions**: Use database transactions for multi-table operations

## Development Setup

### Prerequisites

- Go 1.24 or higher
- Docker and Docker Compose
- MySQL 8.0 (for local development)
- Git

### Local Development

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd mystery-dashboard-api
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Set up environment variables**:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Start development services**:
   ```bash
   docker-compose up -d mysql redis jaeger
   ```

5. **Run database migrations**:
   ```bash
   go run cmd/migrate/main.go up
   ```

6. **Start the application**:
   ```bash
   go run cmd/server/main.go
   ```

### Testing

#### Unit Tests
```bash
# Run all unit tests
go test ./... -v

# Run tests with coverage
go test ./... -cover

# Run specific package tests
go test ./internal/handlers -v
```

#### Integration Tests
```bash
# Start test environment
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
go test ./test/integration -v

# Clean up test environment
docker-compose -f docker-compose.test.yml down -v
```

### Code Quality

#### Formatting and Linting
```bash
# Format code
go fmt ./...

# Vet code for issues
go vet ./...

# Run golangci-lint (if installed)
golangci-lint run
```

#### Pre-commit Checks
Before committing code, ensure:
1. All tests pass
2. Code is properly formatted
3. No linting errors
4. Documentation is updated

## API Documentation

### Swagger/OpenAPI

The API is documented using OpenAPI 3.0 specifications. Access the interactive documentation at:
- Development: `http://localhost:8080/swagger/index.html`
- The swagger documentation is automatically generated from code annotations

### Key Endpoints

#### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration
- `GET /api/v1/auth/me` - Get current user profile
- `POST /api/v1/auth/logout` - User logout

#### Video Management
- `GET /api/v1/videos` - List videos
- `POST /api/v1/videos` - Create video
- `GET /api/v1/videos/{id}` - Get video details
- `PUT /api/v1/videos/{id}` - Update video
- `DELETE /api/v1/videos/{id}` - Delete video
- `POST /api/v1/videos/{id}/upload` - Upload video file
- `POST /api/v1/videos/{id}/publish` - Publish video to platforms

#### Platform Integration
- `POST /webhooks/{platform}` - Platform webhooks
- `GET /api/v1/platforms/{platform}/auth` - Initiate OAuth
- `POST /api/v1/platforms/{platform}/auth/callback` - OAuth callback

#### Analytics
- `GET /api/v1/stats/dashboard` - Dashboard overview
- `GET /api/v1/stats/videos/{id}` - Video statistics
- `GET /api/v1/stats/performance` - Performance analytics

#### AI Processing
- `GET /api/v1/ai/batches` - List AI batches
- `POST /api/v1/ai/batches` - Create AI batch
- `POST /api/v1/ai/batches/{id}/start` - Start batch processing

## Deployment

### Environment Configuration

Required environment variables:
- `DATABASE_DSN`: MySQL connection string
- `JWT_SECRET`: Secret key for JWT token signing
- `ENVIRONMENT`: Application environment (development, staging, production)
- `LOG_LEVEL`: Logging level (debug, info, warn, error)
- `JAEGER_ENDPOINT`: Jaeger tracing endpoint
- `AWS_*`: AWS credentials for S3 storage

### Production Deployment

1. **Build the application**:
   ```bash
   go build -o mysteryfactory cmd/server/main.go
   ```

2. **Create production Docker image**:
   ```bash
   docker build -t mysteryfactory:latest .
   ```

3. **Deploy with Docker Compose**:
   ```bash
   docker-compose -f docker-compose.prod.yml up -d
   ```

## Monitoring and Observability

### Logging
- Structured logging with JSON format in production
- Request/response logging with correlation IDs
- Error tracking with stack traces
- Performance metrics logging

### Tracing
- OpenTelemetry integration with Jaeger
- Distributed tracing across services
- Database query tracing
- HTTP request tracing

### Health Checks
- `/health` - Basic health check
- `/ready` - Readiness probe for Kubernetes
- Database connectivity checks
- External service dependency checks

## Contributing

### Workflow

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**:
   - Follow coding standards
   - Add tests for new functionality
   - Update documentation as needed

3. **Test your changes**:
   ```bash
   go test ./...
   go fmt ./...
   go vet ./...
   ```

4. **Commit your changes**:
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

5. **Push and create a pull request**:
   ```bash
   git push origin feature/your-feature-name
   ```

### Commit Message Format

Use conventional commits format:
- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `test:` - Test additions or modifications
- `refactor:` - Code refactoring
- `chore:` - Maintenance tasks

## Support and Resources

### Documentation
- API Documentation: `/swagger/index.html`
- Architecture Decision Records: `/docs/adr/`
- Database Schema: `/docs/schema/`

### Development Tools
- Go: https://golang.org/
- Gin Framework: https://gin-gonic.com/
- OpenTelemetry: https://opentelemetry.io/
- Swagger: https://swagger.io/

### Getting Help
- Check existing issues and documentation
- Create detailed bug reports with reproduction steps
- Include relevant logs and error messages
- Provide system information and Go version

---

This documentation serves as a comprehensive guide for developers working on the Mystery Factory API project. Keep it updated as the project evolves and new features are added.