# movie-watchlist-recommendation-api
A RESTful Movie Watchlist &amp; Recommendation API built with Go, Gin, PostgreSQL, and Redis. Integrates with OMDB API to allow users to search movies, manage watchlists, rate films, and receive personalized recommendations based on genres and ratings.
<div align="center">

# 🎬 Movie Watchlist & Recommendation API

**A production-grade REST API for managing movie watchlists, ratings, and personalized recommendations.**

Built with Go · PostgreSQL · Redis · OMDb API

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=for-the-badge&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=for-the-badge&logo=redis&logoColor=white)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)

</div>

---

## 📋 Table of Contents

- [Project Overview](#-project-overview)
- [Features](#-features)
- [Tech Stack](#-tech-stack)
- [System Architecture](#-system-architecture)
- [Database Schema](#-database-schema)
- [API Endpoints](#-api-endpoints)
- [Getting Started](#-getting-started)
- [Environment Variables](#-environment-variables)
- [OMDb API Key Setup](#-omdb-api-key-setup)
- [Redis Setup](#-redis-setup)
- [Recommendation Logic](#-recommendation-logic)
- [Caching Strategy](#-caching-strategy)
- [Example curl Requests](#-example-curl-requests)
- [Future Improvements](#-future-improvements)
- [License](#-license)

---

## 🎯 Project Overview

The **Movie Watchlist & Recommendation API** is a backend capstone project that demonstrates industry-standard Go API development practices. It allows users to search for movies, build personal watchlists, rate films, and receive intelligent content-based recommendations — all powered by the [OMDb API](https://www.omdbapi.com/) with smart caching via Redis.

The project follows **Clean Architecture** principles (Ports & Adapters) to ensure separation of concerns, testability, and maintainability at scale.

```
📌 Key Highlights:
• Clean Architecture with dependency injection
• JWT-based stateless authentication
• Content-based recommendation engine
• Multi-layer caching (Redis + PostgreSQL)
• Rate limiting, CORS, structured logging
• Docker-ready with health checks
• Graceful server shutdown
```

---

## ✨ Features

| Feature | Description |
|---------|-------------|
| 🔐 **Authentication** | Register/login with bcrypt password hashing and JWT tokens |
| 🔍 **Movie Search** | Search 280,000+ movies via OMDb API with auto-caching |
| 📋 **Watchlist Management** | Add, update status (plan_to_watch / watching / watched), remove |
| ⭐ **Movie Ratings** | Rate movies 1–10 with optional text reviews |
| 🤖 **Smart Recommendations** | Content-based engine analyzing your highly-rated genres |
| 🛡️ **Security Middleware** | JWT auth, CORS, IP-based rate limiting (100 req/min) |
| 📊 **Structured Logging** | Production JSON / development colored logs via Zap |
| 🐳 **Docker Support** | One-command setup with Postgres, Redis, and API containers |
| ⚡ **Graceful Shutdown** | Clean resource cleanup on SIGINT/SIGTERM |

---

## 🛠 Tech Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| **Language** | Go 1.22 | High-performance, statically typed backend |
| **Framework** | [Gin](https://github.com/gin-gonic/gin) | Fast HTTP router with middleware support |
| **Database** | PostgreSQL 16 | Primary data store with UUID keys |
| **Cache** | Redis 7 | External API response caching |
| **Auth** | [golang-jwt](https://github.com/golang-jwt/jwt) | Stateless JWT authentication |
| **DB Driver** | [pgx v5](https://github.com/jackc/pgx) | Native PostgreSQL driver with connection pooling |
| **Redis Client** | [go-redis v9](https://github.com/redis/go-redis) | Type-safe Redis operations |
| **Config** | [Viper](https://github.com/spf13/viper) | 12-factor app configuration |
| **Logging** | [Zap](https://go.uber.org/zap) | Blazing-fast structured logging |
| **Validation** | [validator v10](https://github.com/go-playground/validator) | Struct-level input validation |
| **Migrations** | [golang-migrate](https://github.com/golang-migrate/migrate) | Version-controlled schema migrations |
| **External API** | [OMDb API](https://www.omdbapi.com/) | Movie metadata source |
| **Containers** | Docker + Docker Compose | Reproducible development environment |

---

## 🏗 System Architecture

### High-Level Flow

```
┌──────────┐     HTTP      ┌────────────────────────────────────────────┐
│          │ ──────────────▶│              GIN ROUTER                   │
│  Client  │               │  ┌──────┐ ┌─────┐ ┌──────────┐ ┌──────┐  │
│ (Postman/│               │  │ CORS │→│ Log │→│RateLimit │→│ Auth │  │
│  curl)   │               │  └──────┘ └─────┘ └──────────┘ └──────┘  │
│          │ ◀──────────────│                                          │
└──────────┘    JSON        └────────────────┬───────────────────────────┘
                                             │
                                             ▼
                                    ┌────────────────┐
                                    │    HANDLERS     │
                                    │ (Parse & Reply) │
                                    └───────┬────────┘
                                            │
                                            ▼
                                    ┌────────────────┐
                                    │    SERVICES     │
                                    │(Business Logic) │
                                    └──┬─────┬─────┬─┘
                                       │     │     │
                               ┌───────┘     │     └────────┐
                               ▼             ▼              ▼
                      ┌──────────────┐ ┌──────────┐ ┌──────────────┐
                      │  PostgreSQL  │ │  Redis   │ │   OMDb API   │
                      │  (pgx Pool) │ │ (Cache)  │ │  (External)  │
                      └──────────────┘ └──────────┘ └──────────────┘
```

### Clean Architecture Layers

```
movie_recommend/
│
├── cmd/api/                    # Application entry-point & dependency injection
│   └── main.go
│
├── internal/                   # Private application code (Go convention)
│   ├── config/                 # Environment & configuration loading
│   ├── domain/                 # Core entities — ZERO external dependencies
│   │   ├── user.go             #   User entity + auth DTOs
│   │   ├── movie.go            #   Movie entity + OMDb response types
│   │   ├── watchlist.go        #   Watchlist entity + request DTOs
│   │   └── rating.go           #   Rating entity + request DTOs
│   ├── repository/             # Data access layer (Ports & Adapters)
│   │   ├── interfaces.go       #   Port interfaces (contracts)
│   │   ├── postgres/           #   PostgreSQL adapters
│   │   └── redis/              #   Redis cache adapter
│   ├── service/                # Business logic / use-cases
│   │   ├── auth_service.go     #   Register, login, JWT generation
│   │   ├── movie_service.go    #   OMDb search, caching, persistence
│   │   ├── watchlist_service.go#   Watchlist CRUD with ownership checks
│   │   ├── rating_service.go   #   Rating CRUD with ownership checks
│   │   └── recommendation_service.go  # Content-based recommendation engine
│   ├── handler/                # HTTP handlers (request/response layer)
│   ├── middleware/             # Auth, logging, rate-limiting, CORS
│   ├── router/                # Route registration & middleware wiring
│   └── errors/                # Custom error types & HTTP status mapping
│
├── pkg/                       # Shared, reusable packages
│   ├── logger/                # Zap logger initialization
│   ├── response/              # Standardized JSON response builders
│   └── validator/             # Input validation helpers
│
├── migrations/                # SQL migration files
├── docker/                    # Dockerfile & docker-compose.yml
├── .env.example               # Environment variable template
├── Makefile                   # Build, run, test, migrate commands
└── go.mod                     # Go module definition
```

---

## 🗄 Database Schema

### Entity Relationship Diagram

```
┌─────────────────────┐       ┌─────────────────────────┐
│       USERS         │       │         MOVIES           │
├─────────────────────┤       ├─────────────────────────┤
│ id (UUID, PK)       │       │ id (UUID, PK)            │
│ username (UNIQUE)   │       │ imdb_id (UNIQUE)         │
│ email (UNIQUE)      │       │ title                    │
│ password_hash       │       │ year                     │
│ created_at          │       │ genre                    │
│ updated_at          │       │ director                 │
└────────┬────────────┘       │ actors                   │
         │                    │ plot                     │
         │ 1:N                │ poster_url               │
         │                    │ imdb_rating              │
         ▼                    │ created_at               │
┌─────────────────────┐       └────────────┬─────────────┘
│     WATCHLISTS      │                    │
├─────────────────────┤                    │ 1:N
│ id (UUID, PK)       │                    │
│ user_id (FK→users)  │◀───────────────────┘
│ movie_id (FK→movies)│
│ status (ENUM)       │       ┌─────────────────────────┐
│ added_at            │       │        RATINGS           │
├─────────────────────┤       ├─────────────────────────┤
│ UQ(user_id,movie_id)│       │ id (UUID, PK)            │
│ CHK(status IN ...)  │       │ user_id (FK→users)       │
└─────────────────────┘       │ movie_id (FK→movies)     │
                              │ score (1-10)             │
                              │ review (TEXT, nullable)   │
                              │ created_at               │
                              │ updated_at               │
                              ├─────────────────────────┤
                              │ UQ(user_id,movie_id)     │
                              │ CHK(score >= 1 AND <= 10)│
                              └─────────────────────────┘
```

### Table Details

| Table | Records | Key Constraints |
|-------|---------|-----------------|
| **users** | Registered accounts | Unique `username` + `email`, bcrypt hashed passwords |
| **movies** | OMDb movie cache | Unique `imdb_id`, auto-persisted on first access |
| **watchlists** | User → Movie links | One entry per user/movie pair, status enum validation |
| **ratings** | User reviews | One rating per user/movie pair, score 1–10 CHECK constraint |

### Indexes

```sql
-- Users: fast auth lookups
idx_users_email, idx_users_username

-- Movies: search and filter
idx_movies_imdb_id, idx_movies_genre, idx_movies_title

-- Watchlists: user dashboard queries
idx_watchlists_user_id, idx_watchlists_movie_id, idx_watchlists_status

-- Ratings: recommendation engine queries
idx_ratings_user_id, idx_ratings_movie_id, idx_ratings_score
```

---

## 📡 API Endpoints

### Authentication (Public)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/auth/register` | Create a new user account |
| `POST` | `/api/v1/auth/login` | Authenticate and receive JWT |

### Movies (Protected 🔒)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/movies/search?q={title}&page={n}` | Search movies via OMDb |
| `GET` | `/api/v1/movies/:imdbID` | Get full movie details |

### Watchlist (Protected 🔒)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/watchlist` | List all watchlist entries |
| `POST` | `/api/v1/watchlist` | Add a movie to watchlist |
| `PATCH` | `/api/v1/watchlist/:id` | Update entry status |
| `DELETE` | `/api/v1/watchlist/:id` | Remove from watchlist |

### Ratings (Protected 🔒)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/ratings` | Rate a movie (1–10) |
| `GET` | `/api/v1/ratings` | List all your ratings |
| `PUT` | `/api/v1/ratings/:id` | Update a rating |
| `DELETE` | `/api/v1/ratings/:id` | Delete a rating |

### Recommendations (Protected 🔒)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/recommendations` | Get personalized recommendations |

### Health (Public)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/health` | Liveness/readiness check |

---

## 🚀 Getting Started

### Prerequisites

- **Go** 1.22+ → [Download](https://go.dev/dl/)
- **Docker** & **Docker Compose** → [Download](https://www.docker.com/products/docker-desktop)
- **OMDb API Key** (free) → [Get Key](https://www.omdbapi.com/apikey.aspx)

### Step 1: Clone the Repository

```bash
git clone https://github.com/namru/movie-recommend.git
cd movie-recommend
```

### Step 2: Configure Environment

```bash
cp .env.example .env
```

Edit `.env` and set your OMDb API key:

```env
OMDB_API_KEY=your_api_key_here
```

### Step 3: Start Infrastructure (PostgreSQL + Redis)

```bash
# Start Postgres and Redis containers
docker-compose -f docker/docker-compose.yml up -d postgres redis
```

Verify containers are running:

```bash
docker ps
# Should show docker-postgres-1 and docker-redis-1
```

### Step 4: Apply Database Migrations

```powershell
# PowerShell (Windows)
Get-Content migrations/schema.sql | docker exec -i docker-postgres-1 psql -U postgres -d movie_recommend
```

```bash
# Bash (Linux/macOS)
docker exec -i docker-postgres-1 psql -U postgres -d movie_recommend < migrations/schema.sql
```

### Step 5: Install Dependencies & Run

```bash
go mod tidy
go run ./cmd/api
```

You should see:

```
INFO    logger initialized    {"mode": "debug", "pid": 12345}
INFO    connected to PostgreSQL    {"host": "localhost"}
INFO    connected to Redis    {"addr": "localhost:6379"}
INFO    server starting    {"addr": ":8080"}
```

### Step 6: Run with Docker (Full Stack)

```bash
docker-compose -f docker/docker-compose.yml up --build -d
```

---

## ⚙️ Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | HTTP server port |
| `GIN_MODE` | `debug` | `debug` / `release` (affects logging format) |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | PostgreSQL username |
| `DB_PASSWORD` | `postgres` | PostgreSQL password |
| `DB_NAME` | `movie_recommend` | PostgreSQL database name |
| `DB_SSLMODE` | `disable` | PostgreSQL SSL mode |
| `REDIS_HOST` | `localhost` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `REDIS_PASSWORD` | *(empty)* | Redis password |
| `REDIS_DB` | `0` | Redis database number |
| `JWT_SECRET` | *(required)* | Secret key for JWT signing |
| `JWT_EXPIRY_HOURS` | `24` | JWT token validity (hours) |
| `OMDB_API_KEY` | *(required)* | OMDb API key |
| `OMDB_BASE_URL` | `http://www.omdbapi.com` | OMDb API base URL |
| `CACHE_SEARCH_TTL` | `86400` | Search cache TTL (seconds) = 24h |
| `CACHE_MOVIE_TTL` | `604800` | Movie detail cache TTL (seconds) = 7d |

---

## 🔑 OMDb API Key Setup

1. Visit [https://www.omdbapi.com/apikey.aspx](https://www.omdbapi.com/apikey.aspx)
2. Select **FREE! (1,000 daily limit)** plan
3. Enter your email and submit
4. Check your email for the activation link
5. Copy your API key
6. Add it to your `.env` file:

```env
OMDB_API_KEY=your_key_here
```

**Verify your key works:**

```bash
curl "http://www.omdbapi.com/?apikey=your_key_here&s=Inception"
```

> ⚠️ The free tier allows **1,000 requests/day**. Our Redis caching layer minimizes API calls to stay well within this limit.

---

## 🔴 Redis Setup

Redis is used as a caching layer to store OMDb API responses and reduce external API calls.

### With Docker (Recommended)

```bash
docker-compose -f docker/docker-compose.yml up -d redis
```

### Standalone Installation

- **Windows**: [Download Redis for Windows](https://github.com/microsoftarchive/redis/releases)
- **macOS**: `brew install redis && brew services start redis`
- **Linux**: `sudo apt install redis-server && sudo systemctl start redis`

### Verify Connection

```bash
docker exec -it docker-redis-1 redis-cli ping
# Expected: PONG
```

---

## 🧠 Recommendation Logic

The API uses a **content-based filtering** algorithm that analyzes the user's taste profile to suggest new movies.

### Algorithm Steps

```
Step 1: Analyze User Taste
  └── Query user's ratings where score >= 7 (highly rated)
  └── Extract genres from those movies
  └── Rank genres by frequency → Top 3 genres

Step 2: Find Candidates
  └── Search local PostgreSQL database by top genres
  └── If insufficient results, query OMDb API for each genre
  └── Persist new movies to database for future recommendations

Step 3: Filter & Deduplicate
  └── Exclude movies the user has already rated
  └── Exclude duplicates across genres
  └── Return up to 10 recommendations

Step 4: Cold Start Fallback
  └── If user has no ratings (new user)
  └── Use default popular genres: Action, Drama, Comedy
```

### Example Flow

```
User has rated highly:
  "The Dark Knight"     (Action, Crime, Drama)     → Score: 9
  "Inception"           (Action, Adventure, Sci-Fi) → Score: 10
  "Interstellar"        (Adventure, Drama, Sci-Fi)  → Score: 8

Extracted top genres: Action (3), Drama (2), Sci-Fi (2)

Recommendations → movies matching Action/Drama/Sci-Fi
                   that user hasn't already rated
```

---

## ⚡ Caching Strategy

Redis caching is implemented at the **service layer** to minimize OMDb API calls and improve response times.

### Cache Architecture

```
Request Flow:
                                    ┌─ HIT ──▶ Return cached JSON
Client → Handler → Service → Redis ─┤
                                    └─ MISS ──▶ OMDb API → Cache → Return
```

### TTL Configuration

| Cache Key Pattern | TTL | Rationale |
|-------------------|-----|-----------|
| `omdb:search:{query}:{page}` | **24 hours** | Search results change frequently as new movies release |
| `omdb:movie:{imdbID}` | **7 days** | Movie details rarely change; longer cache is safe |

### Benefits

| Metric | Without Cache | With Cache |
|--------|---------------|------------|
| **OMDb API calls** | Every request | Only on cache miss |
| **Response time** | 200–500ms | < 5ms (cache hit) |
| **Daily API quota** | Exhausted quickly | Stays within 1,000/day free tier |
| **Rate limit risk** | High | Near zero |

### Cache Invalidation

- **Automatic**: Redis TTL expiration handles invalidation
- **Manual**: Delete specific keys if immediate refresh is needed

---

## 📝 Example curl Requests

### 1. Register a New User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "moviefan",
    "email": "moviefan@example.com",
    "password": "securepass123"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "user registered successfully",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "username": "moviefan",
      "email": "moviefan@example.com",
      "created_at": "2026-02-20T22:30:00Z",
      "updated_at": "2026-02-20T22:30:00Z"
    }
  }
}
```

### 2. Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "moviefan@example.com",
    "password": "securepass123"
  }'
```

### 3. Search Movies

```bash
curl -X GET "http://localhost:8080/api/v1/movies/search?q=Inception&page=1" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Response:**
```json
{
  "success": true,
  "message": "movies found",
  "data": {
    "Search": [
      {
        "Title": "Inception",
        "Year": "2010",
        "imdbID": "tt1375666",
        "Type": "movie",
        "Poster": "https://m.media-amazon.com/images/..."
      }
    ],
    "totalResults": "8"
  }
}
```

### 4. Get Movie Details

```bash
curl -X GET http://localhost:8080/api/v1/movies/tt1375666 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 5. Add to Watchlist

```bash
curl -X POST http://localhost:8080/api/v1/watchlist \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "imdb_id": "tt1375666",
    "status": "plan_to_watch"
  }'
```

### 6. Update Watchlist Status

```bash
curl -X PATCH http://localhost:8080/api/v1/watchlist/ENTRY_UUID \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "status": "watched"
  }'
```

### 7. Rate a Movie

```bash
curl -X POST http://localhost:8080/api/v1/ratings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "imdb_id": "tt1375666",
    "score": 9,
    "review": "A mind-bending masterpiece by Christopher Nolan."
  }'
```

### 8. Get Recommendations

```bash
curl -X GET http://localhost:8080/api/v1/recommendations \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Response:**
```json
{
  "success": true,
  "message": "recommendations generated",
  "data": [
    {
      "id": "...",
      "imdb_id": "tt0816692",
      "title": "Interstellar",
      "year": "2014",
      "genre": "Adventure, Drama, Sci-Fi",
      "director": "Christopher Nolan",
      "imdb_rating": "8.7"
    }
  ]
}
```

### 9. Health Check

```bash
curl http://localhost:8080/api/v1/health
# {"status":"ok"}
```

---

## 🔮 Future Improvements

| Category | Enhancement | Description |
|----------|-------------|-------------|
| 🔐 **Auth** | Refresh tokens | Add refresh token rotation for better security |
| 🔐 **Auth** | OAuth 2.0 | Google/GitHub social login |
| 🧠 **Recommendations** | Collaborative filtering | Recommend based on similar users' tastes |
| 🧠 **Recommendations** | ML integration | Train a model on user ratings for better predictions |
| 📊 **Analytics** | User statistics | Viewing history, genre distribution charts |
| 🔍 **Search** | Advanced filters | Filter by year, genre, rating range, director |
| 📄 **Documentation** | Swagger UI | Auto-generated interactive API docs |
| 🧪 **Testing** | Unit & integration tests | Repository mocks, handler tests, E2E tests |
| ⚡ **Performance** | Pagination | Cursor-based pagination for large watchlists |
| ⚡ **Performance** | Connection pooling | pgBouncer for high-concurrency deployments |
| 🔔 **Notifications** | Webhooks | New movie alerts for watchlisted genres |
| 📦 **DevOps** | CI/CD pipeline | GitHub Actions for lint, test, build, deploy |
| 📦 **DevOps** | Kubernetes | Helm charts for production deployment |

---

## 📄 License

This project is licensed under the **MIT License** — see the [LICENSE](LICENSE) file for details.

---

<div align="center">

**Built with ❤️ in Go**

⭐ Star this repo if you found it useful!

</div>
