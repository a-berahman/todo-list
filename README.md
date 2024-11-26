# Todo List API

A RESTful API for managing todo items with file attachment capabilities.

## Getting Started

### Prerequisites
- Docker
- Docker Compose


### Running the Application
Using Docker Compose:

```bash
make run
```

This will:
1. Start PostgreSQL, LocalStack (S3, SQS), and the API service
2. Run database migrations
3. Create required S3 bucket and SQS queue

The API will be available at [http://localhost:8080](http://localhost:8080)

### Running Tests
```bash
make test
```

### API Endpoints

#### Create Todo

```
curl --location 'http://localhost:8080/api/v1/upload' \
--form 'description="Buy groceries"' \
--form 'dueDate="2024-12-29T15:04:05Z"' \
--form 'file=@/path/to/file.png'
```

Supported file types: .txt, .png, .jpg


## Project Review Guide

### Architecture

The project follows a hexagonal (ports and adapters) architecture pattern:

1. **Domain Layer**: Contains core business logic and entities
2. **Application Layer**: Orchestrates the flow of data and implements use cases
3. **Infrastructure Layer**: Implements external service integrations
4. **Ports Layer**: Defines interfaces for input/output operations
5. **Handlers Layer**: Manages HTTP request/response lifecycle



### Request Workflow

Here's how a typical create todo request flows through the system:

1. **HTTP Request**
   - Client sends POST request to `/api/v1/upload`
   - Request includes todo description, due date, and optional file attachment

2. **Handler Processing** 
    - Handler validates request parameters
    - Handler converts HTTP request to domain commands/queries
    - Handler invokes the application layer

3. **Application Layer Processing**
    - Application layer validates input parameters
    - Application layer invokes the domain layer
    - Application layer invokes the infrastructure layer

4. **Domain Layer Processing**
    - Domain layer validates input parameters
    - Domain layer invokes the infrastructure layer

5. **Infrastructure Layer Processing**
    - Infrastructure layer handles file storage (S3) and message queue (SQS) operations
    - Infrastructure layer invokes external services (AWS)

6. **External Service Processing**
    - External services handle file storage (S3) and message queue (SQS) operations

        