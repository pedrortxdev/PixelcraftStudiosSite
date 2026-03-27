# Pixelcraft API

Backend API for the Pixelcraft platform, built with Go (Golang) and Gin framework. This API powers the Pixelcraft frontend, handling user authentication, product management, checkout processes, and user dashboards.

## 🚀 Technologies

-   **Language**: Go (1.21+)
-   **Framework**: Gin Web Framework
-   **Database**: PostgreSQL
-   **Authentication**: JWT (JSON Web Tokens)
-   **Documentation**: Swagger/OpenAPI (planned)

## 🛠️ Prerequisites

-   **Go**: Version 1.21 or higher
-   **PostgreSQL**: A running PostgreSQL instance
-   **Make** (optional, for running Makefile commands)

## ⚙️ Configuration

1.  Clone the repository.
2.  Copy `.env.example` to `.env`:
    ```bash
    cp .env.example .env
    ```
3.  Update the `.env` file with your database credentials and other configuration:
    ```env
    DB_HOST=localhost
    DB_PORT=5432
    DB_USER=postgres
    DB_PASSWORD=your_password
    DB_NAME=pixelcraft_db
    JWT_SECRET=your_jwt_secret_key
    CPF_ENCRYPTION_KEY=your_32_byte_key_here
    ```

## 📦 Database Setup

The project includes a `dump_schema.sql` file in `database/` (or similar) to initialize the database schema.
You can also use the provided `setup.ps1` script on Windows or `Makefile` commands if available.

## ▶️ Running the Server

To run the API server locally:

```bash
go run cmd/api/main.go
```

The server will start on port `8080` (default).

## 🏗️ Project Structure

```
backend/
├── cmd/
│   └── api/            # Application entry point
├── internal/
│   ├── config/         # Configuration loading
│   ├── database/       # Database connection
│   ├── handlers/       # HTTP request handlers
│   ├── middleware/     # HTTP middleware (Auth, CORS)
│   ├── models/         # Data structures
│   ├── repository/     # Data access layer
│   └── service/        # Business logic
├── .env                # Environment variables
└── go.mod              # Go module definition
```

## 🧪 Testing

To run tests:

```bash
go test ./...
```

---

Developed by Pixelcraft Team.
