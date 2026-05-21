# Chirpy

A RESTful HTTP server built with Go, created as part of the [Learn HTTP Servers in Go](https://www.boot.dev/courses/learn-http-servers-golang) course on [boot.dev](https://www.boot.dev).

## Overview

Chirpy is a microblogging API — think Twitter, but tiny. It exposes REST endpoints for creating users, posting short messages ("chirps"), and serving a static frontend. The server uses PostgreSQL for persistence and [sqlc](https://sqlc.dev/) for type-safe database access.

## Tech Stack

- **Language:** Go 1.24.1
- **Database:** PostgreSQL (via `lib/pq`)
- **Query generation:** sqlc
- **Environment:** godotenv
- **ID generation:** google/uuid

## Project Structure

```
chirpy/
├── main.go               # Entry point, router setup
├── handlers/             # HTTP handler functions
├── helpers/              # Utility/helper functions
├── internal/
│   └── database/         # sqlc-generated DB layer
├── sql/                  # SQL schema and queries
├── assets/               # Static assets
├── index.html            # Frontend entry point
├── sqlc.yaml             # sqlc configuration
├── go.mod
└── go.sum
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/healthz` | Health check |
| `POST` | `/api/users` | Create a new user |
| `POST` | `/api/chirps` | Create a chirp |
| `GET` | `/api/chirps` | Get all chirps |
| `GET` | `/api/chirps/{chirpID}` | Get a single chirp by ID |
| `GET` | `/admin/metrics` | View request count (admin) |
| `POST` | `/admin/reset` | Reset request count (admin) |
| `GET` | `/app/` | Serves the static frontend |

## Getting Started

### Prerequisites

- Go 1.24.1+
- PostgreSQL running locally
- [sqlc](https://sqlc.dev/) (if regenerating DB code)

### Setup

1. **Clone the repository**

   ```bash
   git clone https://github.com/olimeme/chirpy.git
   cd chirpy
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Configure environment variables**

   Create a `.env` file in the project root:

   ```env
   DB_URL=postgres://username:password@localhost:5432/chirpy?sslmode=disable
   PLATFORM=dev
   ```

4. **Set up the database**

   Run the SQL migrations found in the `sql/` directory against your PostgreSQL instance.

5. **Run the server**

   ```bash
   go run main.go
   ```

   The server starts on **port 8080**. Visit [http://localhost:8080/app/](http://localhost:8080/app/) in your browser.

## Environment Variables

| Variable | Description |
|----------|-------------|
| `DB_URL` | PostgreSQL connection string |
| `PLATFORM` | Runtime platform (`dev` or `prod`) |

## License

This project is for educational purposes as part of the boot.dev curriculum.
