# Task Management API - A DevOps & Go Learning Project

[![Build Status](https://dev.azure.com/Your-Organization-Name/Your-Project-Name/_apis/build/status/Your-Project-Name?branchName=master)](https://dev.azure.com/Your-Organization-Name/Your-Project-Name/_build/latest?definitionId=1&branchName=master)

A hands-on project to build a simple REST API in Go and implement enterprise-grade DevOps practices. The primary goal is to master Azure, CI/CD pipelines, containerization with Docker, and comprehensive test automation.

## 🎯 Project Overview

This repository contains the source code for a Task Management REST API. The application is built using a **Clean Architecture** approach in Go, containerized with Docker, and deployed to **Microsoft Azure** using a fully automated CI/CD pipeline in **Azure DevOps**.

### Key Features & Learning Goals
- **RESTful API:** Full CRUD (Create, Read, Update, Delete) functionality for managing tasks.
- **Clean Architecture:** A well-structured, layered architecture (handlers, services, repositories) for maintainability and testability.
- **Comprehensive Testing:**
  - **Unit Tests:** Mocking dependencies to test business logic in isolation.
  - **Integration Tests:** Verifying component interactions with a real database.
  - **End-to-End Tests:** Testing the fully deployed application as a black box.
- **Containerization:** A multi-stage `Dockerfile` to produce a small, secure, and efficient final image.
- **Local Development Environment:** A `docker-compose.yml` file to easily spin up the API and a PostgreSQL database locally.
- **CI/CD Automation:** An `azure-pipelines.yml` file that defines a complete pipeline to automatically build, test, and deploy the application to Azure on every push to the `master` branch.
- **Cloud Deployment:** Infrastructure provisioned in Microsoft Azure, including Azure App Service, Azure Database for PostgreSQL, and Azure Container Registry.
- **Monitoring (In Progress):** Foundational Prometheus metrics for observing application health and performance.

## 🛠️ Tech Stack

| Component         | Technology / Service                                       |
|-------------------|------------------------------------------------------------|
| **Backend**       | Go (Golang)                                                |
| **API Framework** | `gorilla/mux`                                              |
| **Database**      | PostgreSQL                                                 |
| **Containerization**| Docker & Docker Compose                                    |
| **CI/CD**         | Azure DevOps                                               |
| **Cloud Provider**| Microsoft Azure                                            |
| **Cloud Services**| App Service, Azure DB for PostgreSQL, Container Registry   |
| **Testing**       | Go's native `testing` package, `testify` (mock, assert)    |

## 📂 Project Structure

The project follows the principles of Clean Architecture to separate concerns.

```
task-api/
├── cmd/api/
│   └── main.go                 # Application entry point
├── internal/
│   ├── handlers/               # HTTP request handlers
│   ├── metrics/                # Prometheus metrics definitions
│   ├── models/                 # Data structures (Task, Requests)
│   ├── repository/             # Data access layer (database interaction)
│   └── service/                # Business logic
├── scripts/
│   ├── setup-db.sh             # Sets up local database schema
│   └── setup-azure.sh          # (Optional) Script to create Azure resources
├── tests/
│   ├── e2e/                    # End-to-end tests
│   └── integration/            # Integration tests
├── .env.example                # Template for environment variables
├── azure-pipelines.yml         # Azure DevOps CI/CD pipeline definition
├── Dockerfile                  # Multi-stage Docker build file
├── docker-compose.yml          # Local development environment setup
├── go.mod                      # Go module dependencies
└── README.md
```

## 🚀 Getting Started

### Prerequisites

- [Go](https://go.dev/doc/install) (version 1.22 or later)
- [Docker & Docker Compose](https://www.docker.com/products/docker-desktop/)
- [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli) (optional, for manual Azure management)

### Local Development Setup

1. **Clone the repository:**
   ```bash
   git clone https://github.com/yourusername/task-api.git
   cd task-api
   ```

2. **Create your environment file:**
   Copy the example environment file. This file contains the credentials for your local database.
   ```bash
   cp .env.example .env
   ```

3. **Start the local database:**
   This command will start a PostgreSQL container in the background.
   ```bash
   # Make sure Docker Desktop is running
   docker-compose up -d postgres
   ```

4. **Set up the database schema:**
   This script connects to the running PostgreSQL container and creates the `tasks` table.
   ```bash
   chmod +x ./scripts/setup-db.sh
   ./scripts/setup-db.sh
   ```

### Running the Application

You can run the application in two ways:

1. **Directly with Go (for quick development):**
   ```bash
   go run ./cmd/api/
   ```
   The API will be available at `http://localhost:8080`.

2. **With Docker Compose (recommended):**
   This method runs the entire application within containers, exactly as it would be in production.
   ```bash
   docker-compose up --build
   ```
   The API will be available at `http://localhost:8080`.

## 🧪 Running Tests

The project includes a comprehensive test suite.

- **Run Unit Tests:**
  ```bash
  go test ./internal/... -v
  ```

- **Run Integration Tests:**
  Requires the PostgreSQL Docker container to be running.
  ```bash
  # Make sure the DB is running
  docker-compose up -d postgres
  
  go test ./tests/integration/... -v
  ```

- **Run End-to-End (E2E) Tests:**
  This requires two separate terminal windows.
  
  **Terminal 1:** Start the application.
  ```bash
  go run ./cmd/api/
  ```
  
  **Terminal 2:** Run the E2E tests against the live application.
  ```bash
  go test ./tests/e2e/... -v
  ```

## 📋 API Endpoints

The following endpoints are available:

| Method | Endpoint          | Description                      |
|--------|-------------------|----------------------------------|
| POST   | /api/tasks        | Creates a new task.              |
| GET    | /api/tasks        | Retrieves all tasks.             |
| GET    | /api/tasks/{id}   | Retrieves a single task by ID.   |
| PUT    | /api/tasks/{id}   | Updates an existing task.        |
| DELETE | /api/tasks/{id}   | Deletes a task by ID.            |
| GET    | /health           | Health check endpoint.           |

### Example: Create a Task with curl

```bash
curl -X POST \
  http://localhost:8080/api/tasks \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "My New Task",
    "description": "This is a test task."
  }'
```

## ⚙️ CI/CD Pipeline

The CI/CD pipeline is defined in `azure-pipelines.yml` and managed by Azure DevOps. It automates the following process on every push to the `master` branch:

1. **Build Stage:**
   - A Microsoft-hosted agent is provisioned.
   - The source code is checked out.
   - The Dockerfile is used to build a new Docker image of the application.
   - The newly built image is tagged and pushed to our private Azure Container Registry (ACR).

2. **Deploy Stage:**
   - If the Build stage succeeds, the Deploy stage begins.
   - The pipeline connects securely to our Azure subscription.
   - It instructs the Azure App Service to pull the newly pushed image from ACR and restart, completing the deployment.

## ☁️ Azure Infrastructure

All cloud resources are provisioned in a single resource group (`task-api-rg`) for easy management:

- **Azure App Service:** A PaaS offering to host and run our containerized web application.
- **Azure Database for PostgreSQL:** A fully managed, enterprise-ready PostgreSQL database service.
- **Azure Container Registry (ACR):** A private, secure registry for storing and managing our Docker images.

## 🚀 Future Improvements

- Deploy to Azure Kubernetes Service (AKS) for better scalability.
- Use Terraform to manage infrastructure as code (IaC).
- Add performance tests using a tool like k6.
- Integrate Application Insights for advanced observability and logging.
- Implement user authentication and authorization (e.g., JWT).
