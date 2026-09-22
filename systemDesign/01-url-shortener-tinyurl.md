# URL Shortener (TinyURL) System Design

## 1. Clarify Requirements

### Functional Requirements
- Generate a short URL for a given long URL
- Redirect short URLs to original URLs
- Ensure uniqueness of short URLs
- Support optional custom aliases (e.g., `example.com/launch`)
- Support optional expiration time
- Track click analytics (count, country, referer, time of day)
- Support user accounts and link management

### Non-Functional Requirements
- **High Availability**: 99.99% (four nines) for redirects
- **Low Latency**: P99 redirect under 50ms, cache hits in single-digit milliseconds
- **Read-Heavy**: 100:1 read-to-write ratio (sometimes 1000:1)
- **Scalability**: Handle millions of URLs and billions of redirects
- **Durability**: Never lose URL mappings

### Clarifying Questions
- What is the expected traffic volume?
- Do we need custom aliases or auto-generated only?
- What is the required URL lifetime?
- Do we need analytics or just basic redirect functionality?
- Should short URLs be permanent or have expiration?
- What is the expected growth rate?

## 2. Estimate Scale

### Traffic Assumptions
- 100M new URLs per day
- 1000:1 read-to-write ratio = 10B redirects per day
- 115K redirects per second (peak)
- 1.2K new URLs per second (peak)

### Storage Requirements
- Each mapping: long URL (2KB) + short code (7 bytes) + metadata (100 bytes) ≈ 2.1KB
- 100M URLs × 2.1KB = 210GB per day
- 5 years = 210GB × 365 × 5 = 383TB

### Short Code Length Calculation
- Base62 encoding (a-z, A-Z, 0-9)
- 6 characters: 62⁶ = 56.8B combinations (not enough)
- 7 characters: 62⁷ = 3.5T combinations (sufficient for 96 years at 100M/day)
- 8 characters: 62⁸ = 218T (with reserve)

**Decision: 7 characters using Base62**

### Bandwidth Requirements
- Redirect requests: 115K req/s × 1KB response = 115MB/s
- Create requests: 1.2K req/s × 2KB request = 2.4MB/s
- Total bandwidth: ~120MB/s

## 3. APIs

### Create Short URL
```
POST /api/v1/shorten
Request: {
  "long_url": "https://example.com/very/long/url",
  "custom_alias": "mylink",  // optional
  "expiration_date": "2024-12-31",  // optional
  "user_id": "user123"  // optional
}
Response: {
  "short_url": "https://tinyurl.com/aB3xK9",
  "short_code": "aB3xK9",
  "created_at": "2024-01-01T00:00:00Z",
  "expires_at": "2024-12-31T23:59:59Z"
}
```

### Redirect to Long URL
```
GET /:code
Response: 301 Redirect to long URL
```

### Get URL Analytics
```
GET /api/v1/analytics/:code
Response: {
  "short_code": "aB3xK9",
  "total_clicks": 15234,
  "clicks_by_country": {"US": 8000, "UK": 2000, ...},
  "clicks_by_date": {"2024-01-01": 500, "2024-01-02": 600, ...},
  "last_clicked": "2024-01-15T10:30:00Z"
}
```

### Delete URL
```
DELETE /api/v1/urls/:code
Response: 204 No Content
```

## 4. Data Model

### URL Mappings Table
```sql
CREATE TABLE url_mappings (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(7) UNIQUE NOT NULL,
    long_url TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    user_id BIGINT,
    custom_alias VARCHAR(50),
    INDEX idx_short_code (short_code),
    INDEX idx_user_id (user_id),
    INDEX idx_expires_at (expires_at)
);
```

### Click Analytics Table
```sql
CREATE TABLE click_analytics (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(7) NOT NULL,
    clicked_at TIMESTAMP DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT,
    referer TEXT,
    country VARCHAR(2),
    INDEX idx_short_code_clicked (short_code, clicked_at),
    INDEX idx_clicked_at (clicked_at)
);
```

### User Accounts Table
```sql
CREATE TABLE users (
    user_id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    storage_quota INT DEFAULT 1000
);
```

### Redis Cache Schema
```
Key: short_code
Value: {
  "long_url": "https://example.com/...",
  "expires_at": "2024-12-31T23:59:59Z"
}
TTL: 24 hours (or based on expiration date)
```

## 5. High Level Design

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   CDN/Edge  │  (Cache hot redirects)
└──────┬──────┘
       │
       ▼
┌─────────────┐
│Load Balancer│
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ API Gateway │  (Rate limiting, auth)
└──────┬──────┘
       │
       ▼
┌─────────────────────────────┐
│   URL Shortener Service     │
│  (Stateless, scalable)      │
└──────┬──────────────────────┘
       │
       ├──────────────┬──────────────┐
       ▼              ▼              ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│  Redis      │ │ PostgreSQL  │ │   Analytics │
│  (Cache)    │ │ (Primary DB)│ │  Service    │
└─────────────┘ └─────────────┘ └─────────────┘
                      │
                      ▼
              ┌─────────────┐
              │   Read      │
              │  Replicas   │
              └─────────────┘
```

### Key Design Decisions
- **Stateless service layer** for horizontal scaling
- **Cache-first architecture** for read-heavy workload
- **Separate analytics service** to not impact critical redirect path
- **CDN edge caching** for extremely hot URLs

## 6. Deep Dive Components

### 6.1 Short Code Generation

#### Strategy 1: Hash + Collision Resolution
- Compute MD5/SHA256 of long URL
- Take first 7 characters
- Check for collision in database
- If collision, append random characters or use different hash

**Pros:**
- Simple to implement
- No coordination needed across servers
- Deterministic (same URL always generates same code)

**Cons:**
- Collision handling adds complexity
- Cannot guarantee sequential allocation
- Hash computation overhead

#### Strategy 2: Unique ID Generator + Base62 (Recommended)
- Use distributed unique ID generator (Snowflake)
- Convert numeric ID to Base62
- Store mapping in database

**Pros:**
- Guaranteed uniqueness
- No collision handling needed
- Sequential allocation possible

**Cons:**
- Requires coordination for ID generation
- Single point of failure for ID generator
- Same URL can generate different codes

**Implementation:**
```python
def generate_short_code():
    unique_id = snowflake_generator.next_id()
    return base62_encode(unique_id)[:7]
```

### 6.2 Cache Layer Design

#### Cache-Aside Pattern
```python
def get_long_url(short_code):
    # Check cache
    long_url = redis.get(short_code)
    if long_url:
        return long_url

    # Cache miss - query database
    url_mapping = db.query("SELECT long_url FROM url_mappings WHERE short_code = ?", short_code)
    if url_mapping:
        # Populate cache
        redis.setex(short_code, 3600, url_mapping.long_url)
        return url_mapping.long_url

    return None
```

#### Cache Configuration
- **TTL**: 24 hours (or based on URL expiration)
- **Eviction Policy**: LRU
- **Memory**: 100GB for hot URLs (≈50M mappings)
- **Replication**: Master-slave for high availability

#### Cache Warming
- Pre-load frequently accessed URLs based on analytics
- Use background jobs to warm cache during low-traffic periods
- Implement proactive refresh for expiring hot keys

### 6.3 Database Sharding Strategy

#### Sharding Key Options
1. **By short_code hash**: Even distribution, good for redirect queries
2. **By user_id**: User-specific queries optimized, uneven distribution
3. **By created_at**: Time-based queries optimized, hot spot on recent data

**Recommendation: Shard by short_code hash**

#### Shard Count
- 100M URLs, 10 shards = 10M URLs per shard
- Each shard: 10M × 2.1KB = 21GB
- Add 50% buffer = 32GB per shard

### 6.4 Analytics Pipeline

#### Asynchronous Logging
```python
def log_click(short_code, metadata):
    # Non-blocking async write
    analytics_queue.enqueue({
        "short_code": short_code,
        "clicked_at": datetime.now(),
        "ip_address": metadata.ip,
        "user_agent": metadata.user_agent,
        "referer": metadata.referer
    })
```

#### Batch Processing
- Buffer click events in memory
- Flush to database in batches (1000 events)
- Use separate database to avoid impacting redirect performance

## 7. Scaling Strategy

### 7.1 Horizontal Scaling

#### Service Layer
- Stateless services can be scaled horizontally
- Auto-scaling based on CPU/memory metrics
- Load balancer distributes traffic evenly

#### Database Scaling
- **Read Replicas**: Multiple read replicas for analytics queries
- **Sharding**: Distribute data across multiple database shards
- **Connection Pooling**: Reuse database connections

#### Cache Scaling
- **Redis Cluster**: Distributed cache with automatic sharding
- **Consistent Hashing**: Minimal data movement on topology changes
- **Read Replicas**: Cache replicas for read-heavy workloads

### 7.2 Vertical Scaling
- Increase database server resources (CPU, RAM, storage)
- Optimize database queries and indexes
- Use SSD storage for faster I/O

### 7.3 CDN Integration
- Cache redirects at edge locations
- Reduce latency for global users
- Handle DDoS attacks at edge level

## 8. Reliability Strategy

### 8.1 Redundancy
- **Multi-region deployment**: Deploy in multiple AWS regions
- **Database replication**: Master-slave with automatic failover
- **Cache replication**: Redis cluster with replicas
- **Load balancer redundancy**: Multiple load balancer instances

### 8.2 Failure Detection
- **Health checks**: Regular health checks for all components
- **Circuit breakers**: Prevent cascade failures
- **Monitoring**: Real-time monitoring of all metrics

### 8.3 Data Backup
- **Database backups**: Daily snapshots with point-in-time recovery
- **Cache persistence**: Redis AOF/RDB for durability
- **Cross-region replication**: Replicate data to backup region

### 8.4 Disaster Recovery
- **RTO (Recovery Time Objective)**: 1 hour
- **RPO (Recovery Point Objective)**: 5 minutes
- **Failover testing**: Regular failover drills

## 9. Tradeoffs

### 9.1 Hash vs Unique ID Generation
| Aspect | Hash | Unique ID |
|--------|------|-----------|
| Uniqueness | Collision handling needed | Guaranteed |
| Coordination | None needed | Required |
| Deterministic | Yes | No |
| Complexity | Collision handling | ID generator |

**Decision: Unique ID for production systems**

### 9.2 SQL vs NoSQL Database
| Aspect | SQL | NoSQL |
|--------|-----|-------|
| Consistency | Strong | Eventual |
| Query Complexity | Rich queries | Limited |
| Scaling | Vertical | Horizontal |
| Schema | Fixed | Flexible |

**Decision: SQL for strong consistency, NoSQL for write-heavy workloads**

### 9.3 Cache TTL
| TTL | Pros | Cons |
|-----|------|------|
| Short (1 hour) | Fresh data | High cache miss rate |
| Medium (24 hours) | Balanced | Moderate staleness |
| Long (7 days) | High hit rate | Stale data risk |

**Decision: Medium TTL with proactive refresh for hot keys**

### 9.4 Single Region vs Multi-Region
| Aspect | Single Region | Multi-Region |
|--------|---------------|---------------|
| Latency | Higher for global users | Lower globally |
| Cost | Lower | Higher |
| Complexity | Simple | Complex |
| Disaster Recovery | Limited | Better |

**Decision: Multi-region for global services**

## 10. Bottlenecks

### 10.1 Current Bottlenecks
1. **Database reads**: Analytics queries can overload database
2. **Cache misses**: Cold start after cache eviction
3. **ID generator**: Can become single point of failure
4. **Network bandwidth**: High redirect volume

### 10.2 Mitigation Strategies

#### Database Read Bottleneck
- Use read replicas for analytics queries
- Implement query optimization and indexing
- Use materialized views for common aggregations

#### Cache Miss Bottleneck
- Implement cache warming strategies
- Use CDN edge caching for hot URLs
- Implement request coalescing for cache stampede

#### ID Generator Bottleneck
- Use distributed ID generator (Snowflake)
- Implement multiple ID generator instances
- Use ZooKeeper/etcd for coordination

#### Network Bottleneck
- Use CDN for global distribution
- Implement compression for responses
- Optimize TLS handshake

## 11. Follow-ups

### 11.1 Custom Aliases
**Question**: How do you handle custom aliases?
**Answer**:
- Check availability in database before reserving
- Use separate namespace for custom aliases
- Implement conflict resolution (suggest alternatives)
- Rate limit custom alias creation per user

### 11.2 URL Expiration
**Question**: How do you handle expired URLs?
**Answer**:
- Store expiration date in database
- Implement background cleanup job
- Return 410 Gone for expired URLs
- Cache expiration dates in Redis

### 11.3 Malicious URLs
**Question**: How do you prevent abuse with malicious URLs?
**Answer**:
- Implement URL blacklist (phishing, malware domains)
- Use third-party security APIs (Google Safe Browsing)
- Rate limit per user/IP
- Implement CAPTCHA for bulk operations
- User reporting mechanism

### 11.4 Analytics Scale
**Question**: How do you handle analytics at scale?
**Answer**:
- Separate analytics database
- Use time-series database for click data
- Implement data aggregation and rollups
- Use stream processing (Kafka + Spark)
- Implement data retention policies

### 11.5 Short Code Exhaustion
**Question**: What happens when we run out of 7-character codes?
**Answer**:
- Monitor code usage and set alerts
- Increase to 8 characters before exhaustion
- Implement code recycling for expired URLs
- Use different encoding schemes if needed

### 11.6 Database Migration
**Question**: How do you migrate to a new database schema?
**Answer**:
- Implement dual-write strategy
- Migrate historical data in batches
- Validate data consistency
- Gradual traffic shift
- Rollback plan

### 11.7 Multi-tenancy
**Question**: How do you support multiple tenants/customers?
**Answer**:
- Add tenant_id to all tables
- Implement tenant isolation at API level
- Per-tenant rate limiting
- Tenant-specific analytics
- Separate database shards per tenant

### 11.8 Real-time Analytics
**Question**: How do you provide real-time click analytics?
**Answer**:
- Use stream processing (Kafka + Flink/Spark)
- Implement real-time aggregation
- Use WebSocket for live updates
- Pre-compute common aggregations
- Use in-memory analytics database

### 11.9 Rate Limiting
**Question**: How do you implement rate limiting?
**Answer**:
- Token bucket algorithm
- Redis-based distributed rate limiting
- Per-user and per-IP limits
- Different limits for different operations
- Sliding window for accurate limiting

### 11.10 SEO Considerations
**Question**: How do you handle SEO for short URLs?
**Answer**:
- Implement 301 redirects for SEO
- Allow custom domain names
- Provide canonical URL meta tags
- Implement URL preview services
- Support robots.txt

---

## Summary

The URL shortener system is designed as a read-heavy, cache-first architecture with:
- **7-character Base62 codes** for sufficient uniqueness
- **Cache-aside pattern** for optimal redirect performance
- **Distributed ID generation** for guaranteed uniqueness
- **Separate analytics pipeline** to not impact critical path
- **Multi-region deployment** for global availability
- **Horizontal scaling** at all layers

The key insight is that redirects are 1000x more frequent than URL creation, so the architecture is optimized for fast reads with caching, CDN integration, and database read replicas.