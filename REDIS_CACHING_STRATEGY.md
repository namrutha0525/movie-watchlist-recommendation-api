# 🔴 Redis Caching Strategy

## Movie Watchlist & Recommendation API

---

## 1. Why Caching External API Calls Is Important

The Movie Watchlist & Recommendation API integrates with the **OMDb API** — a third-party service for movie metadata. Every search and movie detail request triggers an outbound HTTP call. Without caching, this architecture introduces several critical risks:

| Risk | Impact |
|------|--------|
| **Rate Limiting** | OMDb's free tier caps at **1,000 requests/day**. A single user searching and browsing movies can exhaust this within minutes at scale. |
| **Latency** | Each OMDb API call adds **200–500ms** of network round-trip time. Repeated calls for the same data degrade the user experience. |
| **Availability** | External services experience outages. If OMDb is down, uncached requests fail entirely. Cached data provides **graceful degradation**. |
| **Cost** | Paid tiers charge per request. Eliminating redundant calls directly reduces operating costs. |
| **Scalability** | As user count grows, N users searching for "Inception" should not produce N external API calls — the response is identical for all of them. |

> **Principle:** External API calls should be treated as an expensive, rate-limited resource. Cache aggressively, invalidate conservatively.

---

## 2. How Redis Is Used in This Project

Redis serves as a **read-through cache layer** between the application's service layer and the OMDb API. It is **not** used as a primary data store — PostgreSQL holds that role.

### Architecture Position

```
┌──────────┐      ┌─────────────┐      ┌───────────┐      ┌──────────┐
│  Client   │─────▶│   Service   │─────▶│   Redis   │─────▶│ OMDb API │
│           │◀─────│   Layer     │◀─────│  (Cache)  │◀─────│(External)│
└──────────┘      └──────┬──────┘      └───────────┘      └──────────┘
                         │
                         ▼
                  ┌──────────────┐
                  │  PostgreSQL  │
                  │  (Permanent) │
                  └──────────────┘
```

### Cache Flow — Search Movies

```
1. User searches "Inception"
2. Service builds cache key: "omdb:search:Inception:1"
3. Redis GET → cache key
   ├── HIT  → Deserialize JSON → Return immediately (< 5ms)
   └── MISS → Call OMDb API
              ├── Store response in Redis with TTL
              └── Return response to user
```

### Cache Flow — Movie Details

```
1. User requests movie tt1375666
2. Service checks PostgreSQL first (permanent cache)
   ├── FOUND → Return from DB immediately
   └── NOT FOUND → Check Redis key "omdb:movie:tt1375666"
                    ├── HIT  → Persist to PostgreSQL + Return
                    └── MISS → Call OMDb API
                               ├── Store in Redis + PostgreSQL
                               └── Return response
```

### Implementation

The cache is accessed via a **repository interface** (`CacheRepository`), keeping the service layer decoupled from Redis:

```go
type CacheRepository interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key string, value string, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
}
```

The Redis adapter treats cache misses as empty strings (not errors), allowing clean caller logic without error-type checking:

```go
func (r *CacheRepo) Get(ctx context.Context, key string) (string, error) {
    val, err := r.client.Get(ctx, key).Result()
    if err == redis.Nil {
        return "", nil  // Cache miss — not an error
    }
    return val, err
}
```

---

## 3. Cache Expiration Strategy

We use **TTL-based (Time-To-Live) expiration** — every cached entry is written with a finite lifespan. After the TTL elapses, Redis automatically evicts the key.

### TTL Configuration

| Data Type | Cache Key | TTL | Justification |
|-----------|-----------|-----|---------------|
| **Search Results** | `omdb:search:{query}:{page}` | **24 hours** (86,400s) | Search results can change as new movies release. Daily refresh balances freshness with efficiency. |
| **Movie Details** | `omdb:movie:{imdbID}` | **7 days** (604,800s) | Movie metadata (title, director, plot) is immutable. Longer cache is safe and maximizes hit rate. |

### Why TTL-Based?

| Strategy | Pros | Cons | Used Here? |
|----------|------|------|-----------|
| **TTL-based** | Simple, automatic cleanup, predictable memory | Stale data possible within TTL window | ✅ Yes |
| **Write-through** | Always consistent | Requires change detection on external API | ✗ No |
| **Manual invalidation** | Precise control | Complex, error-prone | ✗ No |
| **LRU eviction** | Automatic under memory pressure | Unpredictable evictions | ✗ (Redis fallback) |

TTL-based expiration is optimal here because **OMDb data changes infrequently** and exact consistency is not critical — a movie's plot or rating shifting by 0.1 within a 24-hour window is acceptable.

### Environment Configuration

TTL values are configurable via environment variables, enabling tuning without code changes:

```env
CACHE_SEARCH_TTL=86400    # 24 hours in seconds
CACHE_MOVIE_TTL=604800    # 7 days in seconds
```

---

## 4. Cache Key Structure

Cache keys follow a **namespaced, hierarchical convention** to prevent collisions and enable pattern-based operations.

### Key Format

```
{service}:{resource}:{identifier}:{qualifier}
```

### Key Definitions

| Key Pattern | Example | Description |
|-------------|---------|-------------|
| `omdb:search:{query}:{page}` | `omdb:search:inception:1` | Cached search result page |
| `omdb:movie:{imdbID}` | `omdb:movie:tt1375666` | Cached movie detail |

### Design Decisions

| Decision | Rationale |
|----------|-----------|
| **`omdb:` prefix** | Namespaces all OMDb-related keys. If Redis is shared across services, this prevents key collisions. |
| **`search` vs `movie`** | Separates resource types for independent TTL policies and monitoring. |
| **`{query}` as-is** | The raw search term is used directly. Identical queries produce identical keys ensuring deduplication. |
| **`{page}` qualifier** | OMDb paginates results (10 per page). Each page is cached independently to avoid storing bloated composite responses. |
| **IMDb ID as identifier** | `tt1375666` is a globally unique, immutable identifier — ideal for cache keys. |

### Key Collision Prevention

- Different search queries produce different keys: `omdb:search:batman:1` ≠ `omdb:search:superman:1`
- Same query, different pages: `omdb:search:batman:1` ≠ `omdb:search:batman:2`
- Searches vs details: `omdb:search:inception:1` ≠ `omdb:movie:tt1375666`

---

## 5. Cache Invalidation Policy

### Primary Strategy: Passive TTL Expiration

The project uses a **passive invalidation** model — cached entries are not explicitly invalidated. Instead, they expire naturally based on their TTL.

```
Entry Written ──────────── TTL Window ──────────── Auto-Evicted
              └── Cached data served ──┘           └── Next request triggers fresh API call
```

### Justification

| Factor | Analysis |
|--------|----------|
| **Data mutability** | Movie metadata from OMDb is effectively immutable. Titles, directors, and plots do not change. |
| **Rating freshness** | IMDb ratings may shift marginally over time, but a 24h–7d delay is acceptable for a watchlist application. |
| **Consistency requirement** | This is not a financial or real-time system. Eventual consistency is sufficient. |
| **Complexity trade-off** | Active invalidation would require webhooks or polling from OMDb (which doesn't support either). TTL is the pragmatic choice. |

### Fallback Behavior

```
If Redis is unavailable:
  ├── Cache GET fails → Log warning → Proceed to OMDb API (graceful degradation)
  ├── Cache SET fails → Log warning → Response still returned to user
  └── Application continues functioning without caching
```

This ensures **Redis is never a single point of failure** — its unavailability degrades performance but not correctness.

### Manual Invalidation (If Needed)

The `CacheRepository` interface exposes a `Delete` method for administrative cache clearing:

```go
cache.Delete(ctx, "omdb:movie:tt1375666")  // Force refresh on next request
```

---

## 6. Performance Benefits

### Latency Comparison

| Scenario | Latency | Source |
|----------|---------|--------|
| Cache HIT (Redis) | **1–5 ms** | In-memory key-value lookup |
| Cache MISS → OMDb API | **200–500 ms** | Network round-trip to external API |
| Database lookup (PostgreSQL) | **5–20 ms** | Disk I/O with indexed query |

### Effective Improvement

```
Average response time improvement:

Without caching:   ~350ms (every request hits OMDb)
With caching:      ~10ms  (95%+ cache hit rate expected)

Speed improvement: ~35x faster for cached responses
```

### Hit Rate Analysis

| User Behavior | Cache Impact |
|---------------|-------------|
| User A searches "Batman" | MISS → API call → cached |
| User B searches "Batman" | HIT → served from Redis |
| User A views "The Dark Knight" (tt0468569) | MISS → API call → cached |
| User C views "The Dark Knight" | HIT → served from Redis |
| 100 users view "Inception" | **1 API call**, 99 cache hits |

### API Quota Conservation

```
Without caching (100 active users):
  ~500 API calls/hour → 1,000 daily limit exhausted in ~2 hours

With caching (100 active users):
  ~50 unique searches/hour → ~50 API calls/hour
  Daily total: ~400 calls → well within 1,000 limit
```

---

## 7. Scalability Benefits

### Horizontal Scaling

```
                    ┌──────────┐
                    │  Redis   │  ← Shared cache across all instances
                    │ (Single) │
                    └────┬─────┘
                         │
            ┌────────────┼────────────┐
            ▼            ▼            ▼
       ┌─────────┐ ┌─────────┐ ┌─────────┐
       │ API #1  │ │ API #2  │ │ API #3  │  ← Multiple Go instances
       └─────────┘ └─────────┘ └─────────┘
```

| Scaling Benefit | Explanation |
|-----------------|-------------|
| **Shared cache** | All API instances share one Redis server. A cache entry written by Instance #1 is immediately available to Instance #2. No per-instance duplication. |
| **Stateless API** | The Go API servers hold no local state. Any instance can serve any request. This enables load-balanced horizontal scaling. |
| **Reduced DB pressure** | OMDb data is served from Redis, not PostgreSQL. The database handles only user-specific data (watchlists, ratings), keeping query volume manageable. |
| **Redis clustering** | For very high throughput, Redis can be deployed as a cluster with automatic sharding across nodes. No application code changes required. |
| **Independent scaling** | Redis, PostgreSQL, and API containers scale independently. Cache layer can be enlarged without touching the application tier. |

### Memory Estimation

| Data Type | Avg. Size per Entry | 1,000 Entries | 10,000 Entries |
|-----------|---------------------|---------------|----------------|
| Search result | ~2 KB | ~2 MB | ~20 MB |
| Movie detail | ~1 KB | ~1 MB | ~10 MB |
| **Total** | | **~3 MB** | **~30 MB** |

> Even at 10,000 cached entries, Redis memory usage stays under **30 MB** — trivial for any deployment.

### Production Scaling Path

```
Phase 1 (MVP):        Single Redis instance (current)
Phase 2 (Growth):     Redis Sentinel (high availability)
Phase 3 (Scale):      Redis Cluster (horizontal sharding)
Phase 4 (Global):     Redis with read replicas per region
```

---

## Summary

| Aspect | Strategy |
|--------|----------|
| **Cache Layer** | Redis 7 (in-memory key-value store) |
| **Pattern** | Read-through cache with TTL-based expiration |
| **Search TTL** | 24 hours (configurable) |
| **Movie TTL** | 7 days (configurable) |
| **Invalidation** | Passive (TTL-based), with manual `Delete` available |
| **Failure Mode** | Graceful degradation — API works without Redis |
| **Performance** | ~35x latency improvement on cache hits |
| **Scalability** | Shared across instances, Redis Cluster-ready |
| **Memory** | < 30 MB for 10,000 cached entries |

---

*Document prepared for: Movie Watchlist & Recommendation API — Capstone Project*
