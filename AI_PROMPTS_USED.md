# 🤖 AI Prompts Used to Build This Project

> **Project:** Movie Watchlist & Recommendation API
> **AI Assistant:** Gemini (Senior Go Backend Architect role)
> **Date:** February 2026

This document lists every AI prompt used during the development of this project, in chronological order. Each prompt contributed to a specific phase of the software development lifecycle.

---

## Table of Contents

- [Prompt 1 — Architecture Design](#prompt-1--architecture-design)
- [Prompt 2 — Source Code Generation](#prompt-2--source-code-generation)
- [Prompt 3 — SQL Schema Generation](#prompt-3--sql-schema-generation)
- [Prompt 4 — Database Setup via Docker](#prompt-4--database-setup-via-docker)
- [Prompt 5 — Professional README Generation](#prompt-5--professional-readme-generation)
- [Prompt 6 — Redis Caching Strategy Document](#prompt-6--redis-caching-strategy-document)
- [Summary](#summary)

---

## Prompt 1 — Architecture Design

**Phase:** Planning & Design
**Output:** Folder structure, package list, database schema, API endpoints, architecture diagrams

```
You are a senior Go backend architect.

I am building a Capstone Project: Movie Watchlist & Recommendation API.

Requirements:
- REST API in Go
- Integrate with OMDB API (external movie API)
- Users can:
  - Search movies
  - Add movies to watchlist
  - Rate movies
  - Get personalized recommendations
- Use PostgreSQL for database
- Use Redis for caching external API responses
- Follow clean architecture structure
- Include proper folder structure
- Include middleware
- Include environment config
- Include error handling
- Include logging

Generate:
1. Full production-ready folder structure
2. Short explanation of each folder
3. Required Go packages
4. Database schema design
5. API endpoint list

Do NOT generate code yet.
Only design architecture first.
```

---

## Prompt 2 — Source Code Generation

**Phase:** Implementation
**Output:** 44 production-ready Go source files across all clean architecture layers

```
(User approved the architecture design document)
```

> **Note:** After reviewing and approving the architecture design, the AI proceeded to generate the full source code across all layers — config, domain entities, repository interfaces and adapters (PostgreSQL + Redis), service layer (auth, movie, watchlist, rating, recommendation), HTTP handlers, middleware (JWT auth, CORS, rate-limiting, logging), router, main entry-point with dependency injection, Docker files, and project tooling (Makefile, .env, .gitignore).

---

## Prompt 3 — SQL Schema Generation

**Phase:** Database Design
**Output:** Consolidated `migrations/schema.sql` with all tables, indexes, constraints, and triggers

```
OMDb API: http://www.omdbapi.com/?i=tt3896198&apikey=efeb080

Generate PostgreSQL schema SQL file for this project.

Include:
- users table
- movies table
- watchlist table
- ratings table
- Proper indexing
- Foreign keys
- Constraints
```

---

## Prompt 4 — Database Setup via Docker

**Phase:** Infrastructure Setup
**Output:** Running PostgreSQL and Redis containers with schema applied

```
docker-compose -f docker/docker-compose.yml up -d postgres
# Wait a few seconds for Postgres to be ready, then:
docker exec -i movie_recommend-postgres-1 psql -U postgres -d movie_recommend -f - < migrations/schema.sql

u run this now
```

---

## Prompt 5 — Professional README Generation

**Phase:** Documentation
**Output:** Industry-level `README.md` with 14 sections

```
Generate a professional README.md for my GitHub repository.

Project Name:
Movie Watchlist & Recommendation API

Include:

1. Project overview
2. Features
3. Tech stack
4. System architecture diagram (in markdown)
5. Database schema explanation
6. API endpoints documentation (with example requests)
7. How to run locally
8. Environment variables setup
9. OMDB API key setup instructions
10. Redis setup
11. Recommendation logic explanation
12. Caching strategy explanation
13. Future improvements
14. Example curl requests

Make it professional and industry-level.
```

---

## Prompt 6 — Redis Caching Strategy Document

**Phase:** Design Documentation
**Output:** Submission-ready `docs/REDIS_CACHING_STRATEGY.md` with 7 sections

```
Generate a markdown document explaining my Redis caching strategy.

Explain:

1. Why caching external API calls is important
2. How Redis is used in this project
3. Cache expiration strategy
4. Cache key structure
5. Cache invalidation policy
6. Performance benefits
7. Scalability benefits

Make it submission-ready.
```

---

## Summary

| # | Prompt | Phase | Key Output |
|---|--------|-------|------------|
| 1 | Architecture Design | Planning | Folder structure, packages, schema, endpoints |
| 2 | Source Code Generation | Implementation | 44 Go source files (clean architecture) |
| 3 | SQL Schema | Database | `schema.sql` with tables, FKs, indexes, triggers |
| 4 | Docker Setup | Infrastructure | PostgreSQL + Redis containers running |
| 5 | README Generation | Documentation | Professional 14-section README.md |
| 6 | Caching Strategy | Documentation | 7-section Redis caching strategy document |

### Development Approach

```
Prompt 1 (Architecture)
    │
    ▼
Prompt 2 (Code Generation) ──▶ 44 files across 12 packages
    │
    ▼
Prompt 3 (SQL Schema) ──▶ Database tables, indexes, triggers
    │
    ▼
Prompt 4 (Docker Setup) ──▶ Infrastructure running
    │
    ▼
Prompt 5 (README) ──▶ GitHub-ready documentation
    │
    ▼
Prompt 6 (Caching Doc) ──▶ Design document for submission
```

### Key Takeaways

1. **Architecture-first approach** — Design was reviewed and approved before any code was written
2. **Clean Architecture** — Domain entities have zero external dependencies; layers communicate via interfaces
3. **Iterative prompting** — Each prompt built upon the output of the previous one
4. **Infrastructure as Code** — Docker Compose for reproducible environments
5. **Documentation-driven** — README and design docs generated alongside code

---

*Generated as part of the Movie Watchlist & Recommendation API Capstone Project*
