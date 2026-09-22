# Metrics System Design

## 1. Clarify Requirements

### Functional Requirements
- Multi-format metric collection (pull and push)
- Support standard metric types (Counter, Gauge, Histogram)
- Dynamic filtering and grouping by labels
- Query engine for dashboards and alerting
- Scheduled alert evaluation
- Data retention and downsampling
- Multi-tenant isolation
- Service discovery for target management

### Non-Functional Requirements
- **Ingestion Rate**: Millions of samples per second
- **Query Latency**: Sub-second for typical dashboard queries
- **Storage Efficiency**: High compression ratio
- **Scalability**: Horizontal scaling for ingestion and query
- **Reliability**: No data loss during ingestion
- **Cardinality Management**: Handle high-dimensional metrics

### Clarifying Questions
- What is the expected number of metrics and series?
- What is the required retention period?
- What are the query latency requirements (P50, P99)?
- Do we need multi-tenancy support?
- What are the alerting requirements?
- Do we need to support custom metric types?
- What is the expected growth rate?

## 2. Estimate Scale

### Traffic Assumptions
- 10K services monitored
- 100 metrics per service average
- 10 label combinations per metric
- 15-second scrape interval
- 5 samples per metric per scrape (for histograms)

### Series Calculation
- Total series: 10K × 100 × 10 = 10M active series
- Peak series: 15M (with churn)
- Historical series: 50M (over time)

### Ingestion Rate
- Scrapes per second: 10K / 15 = 667 scrapes/sec
- Samples per scrape: 100 × 5 = 500 samples
- Total samples/sec: 667 × 500 = 333K samples/sec
- Peak samples/sec: 5M samples/sec (with burst)

### Storage Requirements
- Raw samples: 5M × 1.3 bytes = 6.5MB/sec compressed
- Daily storage: 6.5MB × 86400 = 562GB/day
- 15-day retention: 8.4TB
- 30-day retention: 16.8TB
- With downsampling: 10TB total

### Query Load
- Dashboard queries: 1K queries/sec average
- Alert evaluations: 5K rules every 15 seconds = 333 evaluations/sec
- Ad-hoc queries: 100 queries/sec
- Total queries: 1.4K queries/sec

### Node Count Estimation
- Single node capacity: 1M series, 100K samples/sec
- Required series: 10M
- Nodes for series: 10M / 1M = 10 nodes
- Required ingestion: 5M samples/sec
- Nodes for ingestion: 5M / 100K = 50 nodes
- **Total: 50 nodes (ingestion limited)**

## 3. APIs

### Scrape API (Pull)
```
GET /metrics
Response:
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",path="/api/users",status="200"} 12345
http_requests_total{method="POST",path="/api/users",status="201"} 678

# HELP http_request_duration_seconds HTTP request latency
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.1"} 1000
http_request_duration_seconds_bucket{le="0.5"} 2500
http_request_duration_seconds_bucket{le="+Inf"} 5000
http_request_duration_seconds_sum 123.45
http_request_duration_seconds_count 5000
```

### Write API (Push)
```
POST /api/v1/write
Request: {
  "series": [
    {
      "metric": "http_requests_total",
      "labels": {
        "method": "GET",
        "path": "/api/users",
        "status": "200"
      },
      "value": 12345,
      "timestamp": 1704067200000
    }
  ]
}
Response: 200 OK
```

### Query API
```
POST /api/v1/query
Request: {
  "query": "rate(http_requests_total[5m])",
  "time": "2024-01-01T00:00:00Z"
}
Response: {
  "status": "success",
  "data": {
    "resultType": "vector",
    "result": [
      {
        "metric": {
          "method": "GET",
          "path": "/api/users",
          "status": "200"
        },
        "value": [1704067200, "123.45"]
      }
    ]
  }
}
```

### Range Query API
```
POST /api/v1/query_range
Request: {
  "query": "rate(http_requests_total[5m])",
  "start": "2024-01-01T00:00:00Z",
  "end": "2024-01-01T01:00:00Z",
  "step": "15s"
}
Response: {
  "status": "success",
  "data": {
    "resultType": "matrix",
    "result": [
      {
        "metric": {
          "method": "GET",
          "path": "/api/users",
          "status": "200"
        },
        "values": [
          [1704067200, "123.45"],
          [1704067215, "125.67"]
        ]
      }
    ]
  }
}
```

### Alert Rules API
```
POST /api/v1/rules
Request: {
  "name": "high_error_rate",
  "query": "rate(http_requests_total{status=~"5.."}[5m]) > 0.1",
  "duration": "5m",
  "labels": {
    "severity": "critical"
  },
  "annotations": {
    "summary": "High error rate detected",
    "description": "Error rate is {{ $value }} errors/sec"
  }
}
Response: {
  "name": "high_error_rate",
  "status": "active"
}
```

### Targets API
```
GET /api/v1/targets
Response: {
  "activeTargets": [
    {
      "discoveryLabels": {
        "job": "api-server",
        "env": "production"
      },
      "labels": {
        "instance": "api-server-1:8080",
        "job": "api-server"
      },
      "scrapeUrl": "http://api-server-1:8080/metrics",
      "health": "up",
      "lastError": "",
      "lastScrape": "2024-01-01T00:00:00Z",
      "lastScrapeDuration": 0.5
    }
  ]
}
```

## 4. Data Model

### Time Series Data Model
```
Series: metric_name + label set
Example: http_requests_total{method="GET",path="/api/users",status="200"}

Sample: timestamp + value
Example: (1704067200, 12345)
```

### Metric Types

#### Counter
- Monotonically increasing value
- Use case: Request counts, error counts
- Operations: increment, rate calculation

#### Gauge
- Value that can go up or down
- Use case: Current memory usage, queue size
- Operations: set, increment, decrement

#### Histogram
- Distribution of values
- Use case: Request latency distribution
- Operations: observe, calculate quantiles

#### Summary
- Similar to histogram, client-side calculation
- Use case: Request latency (client-calculated)
- Trade-off: Less accurate, no aggregation

### Metadata Schema
```sql
CREATE TABLE metrics_metadata (
    series_id BIGSERIAL PRIMARY KEY,
    metric_name VARCHAR(255) NOT NULL,
    labels JSONB NOT NULL,
    fingerprint VARCHAR(64) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    last_seen TIMESTAMP DEFAULT NOW(),
    INDEX idx_metric_name (metric_name),
    INDEX idx_fingerprint (fingerprint)
);
```

### Label Index
```sql
CREATE TABLE label_index (
    label_name VARCHAR(255) NOT NULL,
    label_value VARCHAR(255) NOT NULL,
    series_id BIGINT NOT NULL,
    PRIMARY KEY (label_name, label_value, series_id),
    INDEX idx_series (series_id)
);
```

### Alert Rules Schema
```sql
CREATE TABLE alert_rules (
    rule_id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    query TEXT NOT NULL,
    duration INTERVAL NOT NULL,
    labels JSONB,
    annotations JSONB,
    evaluation_interval INT DEFAULT 15,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### Alert State Schema
```sql
CREATE TABLE alert_states (
    rule_id BIGINT NOT NULL,
    fingerprint VARCHAR(64) NOT NULL,
    state VARCHAR(20) NOT NULL, -- firing, pending, inactive
    value FLOAT,
    last_evaluated TIMESTAMP DEFAULT NOW(),
    fired_at TIMESTAMP,
    resolved_at TIMESTAMP,
    PRIMARY KEY (rule_id, fingerprint)
);
```

### TSDB Storage Schema (on disk)
```
/data/prometheus/
├── wal/
│   └── 00000001
├── head_chunks/
│   ├── 000001
│   └── 000002
└── blocks/
    ├── 01GKZ...
    │   ├── chunks/
    │   ├── index/
    │   └── meta.json
    └── 01GKZ...
```

## 5. High Level Design

```
┌─────────────────────────────────────────────────────────────┐
│                      Services                                 │
│              (exposing /metrics endpoint)                      │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                  Service Discovery                            │
│              (Consul, Kubernetes, DNS)                        │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                   Scrapers / Ingesters                        │
│                                                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │ Scraper 1   │  │ Scraper 2   │  │  Ingesters  │          │
│  │ (Pull)      │  │ (Pull)      │  │  (Push)     │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                   TSDB Cluster                                │
│                                                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │ TSDB Node 1 │  │ TSDB Node 2 │  │ TSDB Node 3 │          │
│  │             │  │             │  │             │          │
│  │ WAL + Head  │  │ WAL + Head  │  │ WAL + Head  │          │
│  │ + Blocks    │  │ + Blocks    │  │ + Blocks    │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                   Query Engine                               │
│                                                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │ PromQL      │  │ Query       │  │ Query       │          │
│  │ Evaluator   │  │ Parallelizer│  │ Optimizer   │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
└──────────────────────────┬──────────────────────────────────┘
                           │
           ┌───────────────┴───────────────┐
           ▼                               ▼
┌──────────────────┐            ┌──────────────────┐
│  Alert Manager   │            │  Dashboards      │
│                  │            │  (Grafana)       │
│  - Rule Eval     │            │                  │
│  - Alerting      │            │  - Visualization │
│  - Grouping      │            │  - Query UI      │
└──────────────────┘            └──────────────────┘
```

### Key Design Decisions
- **Pull-based collection**: Simpler target management, implicit health checks
- **Time-series database**: Optimized storage format for metrics
- **Label-based data model**: Multi-dimensional querying
- **Separate alerting**: Decoupled evaluation and notification
- **Horizontal scaling**: Multiple TSDB nodes for scale

## 6. Deep Dive Components

### 6.1 Collection Models

#### Pull Model (Prometheus-style)
- Scraper pulls metrics from targets
- Targets expose `/metrics` endpoint
- Service discovery feeds target list

**Pros:**
- Simple target management
- Failed scrape = target down (implicit health check)
- Back-pressure flows naturally
- No push complexity

**Cons:**
- Targets must be reachable from scraper
- Short-lived jobs need Pushgateway
- Higher latency for metric availability

**Use Case**: Long-running services, Kubernetes environments

#### Push Model
- Services push metrics to ingester
- Ingester accepts and stores metrics

**Pros:**
- Works for ephemeral jobs
- Immediate metric availability
- No network restrictions

**Cons:**
- Need to handle push overload
- Complex target management
- Need separate health checking

**Use Case**: Short-lived jobs, edge computing, serverless

**Recommendation**: Support both models in production

### 6.2 TSDB Storage Design

#### Storage Format

##### WAL (Write-Ahead Log)
- New samples appended to WAL
- Flushed to disk periodically
- Replayed on restart for crash recovery

##### Head Block
- In-memory representation of recent data (2-hour window)
- All recent reads served from here
- Periodically flushed to persistent blocks

##### Persistent Blocks
- Immutable directories named by time range
- Contains chunks (compressed samples), index, metadata
- Compaction merges small blocks into larger ones

#### Compression

##### Gorilla XOR Encoding (Float Values)
- Exploits that consecutive values are similar
- XOR gives mostly-zero deltas
- ~1.3 bytes per sample (100x better than raw float64)

##### Delta-of-Delta Encoding (Timestamps)
- Most timestamps are exactly scrape interval apart
- Second derivative usually zero
- Highly compressible

#### Compression Ratio
- Raw: 8 bytes per sample (float64)
- Compressed: 1.3 bytes per sample
- Ratio: 6.15x compression

### 6.3 Query Engine

#### PromQL (Prometheus Query Language)
- Range queries: `rate(http_requests_total[5m])`
- Aggregation: `sum(http_requests_total) by (service)`
- Functions: `rate()`, `irate()`, `increase()`
- Operators: +, -, *, /, comparisons

#### Query Optimization
- **Label index**: Fast filtering by labels
- **Chunk pruning**: Skip irrelevant time ranges
- **Query parallelization**: Distribute across nodes
- **Result caching**: Cache frequent queries

#### Query Execution
```python
def execute_query(query, time_range):
    # Parse query
    ast = parse_query(query)

    # Optimize query
    optimized = optimize_query(ast)

    # Execute in parallel
    results = parallel_execute(optimized, time_range)

    # Aggregate results
    return aggregate(results)
```

### 6.4 Alerting

#### Alert Rules
- PromQL expressions evaluated at intervals
- Example: `rate(http_requests_total[5m]) > 1000`
- Evaluation interval: 15-60 seconds

#### Alert Manager
- Receives alerts from TSDB
- Deduplicates and groups alerts
- Routes to notification channels
- Handles silencing and inhibition

#### Alert Lifecycle
1. Rule evaluation
2. Alert firing (threshold exceeded for duration)
3. Alert grouping
4. Notification routing
5. Alert resolution
6. Notification of resolution

#### Notification Channels
- Email
- Slack/PagerDuty
- Webhooks
- SMS (for critical alerts)

### 6.5 Label Cardinality Management

#### Cardinality
- Number of unique label combinations
- High cardinality = high memory and CPU usage
- Practical limit: ~10M active series per instance

#### Cardinality Sources
- High-cardinality labels: user_id, request_id, ip_address
- Label combinations: product of label value counts
- Churn: new series created and destroyed

#### Mitigation Strategies
- **Label whitelisting**: Only allow specific labels
- **Label value limits**: Limit number of unique values
- **Relabeling**: Drop high-cardinality labels
- **Cardinality limits**: Hard limits per metric
- **Monitoring**: Alert on high cardinality

#### Implementation
```python
def check_cardinality(series_id, labels):
    cardinality = calculate_cardinality(labels)
    if cardinality > MAX_CARDINALITY:
        log_warning(f"High cardinality: {cardinality}")
        if cardinality > HARD_LIMIT:
            raise CardinalityError("Cardinality limit exceeded")
    return True
```

### 6.6 Downsampling

#### Purpose
- Reduce storage cost for long-term data
- Maintain query performance for historical data
- Balance accuracy and cost

#### Downsampling Strategy
- **Raw data**: 15-second intervals (15 days)
- **5-minute data**: 5-minute intervals (90 days)
- **1-hour data**: 1-hour intervals (1 year)
- **1-day data**: 1-day intervals (5 years)

#### Downsampling Operations
- **Gauges**: Average over interval
- **Counters**: Sum over interval
- **Histograms**: Merge buckets

#### Implementation
```python
def downsample(series, interval, aggregation):
    # Group samples by interval
    grouped = group_by_interval(series, interval)

    # Apply aggregation
    downsampled = {}
    for timestamp, samples in grouped.items():
        if aggregation == 'avg':
            downsampled[timestamp] = sum(samples) / len(samples)
        elif aggregation == 'sum':
            downsampled[timestamp] = sum(samples)
        elif aggregation == 'max':
            downsampled[timestamp] = max(samples)

    return downsampled
```

## 7. Scaling Strategy

### 7.1 Horizontal Scaling

#### TSDB Scaling
- **Federation**: Hierarchical Prometheus setup
- **Remote Write**: Send data to long-term storage
- **Sharding**: Partition by metric or label
- **Query Federation**: Query multiple TSDB instances

#### Ingestion Scaling
- **Multiple scrapers**: Distribute scrape targets
- **Load balancing**: Distribute push traffic
- **Queue buffering**: Kafka for push-based ingestion
- **Auto-scaling**: Scale based on ingestion rate

#### Query Scaling
- **Query parallelization**: Distribute across nodes
- **Result caching**: Cache frequent queries
- **Query routing**: Route to appropriate TSDB
- **Read replicas**: Separate query nodes

### 7.2 Vertical Scaling
- Increase TSDB node resources (CPU, RAM, disk)
- Use SSD storage for faster I/O
- Optimize compression settings
- Tune query engine parameters

### 7.3 Geographic Distribution
- **Multi-region deployment**: Deploy in multiple regions
- **Regional scraping**: Scrape services regionally
- **Data aggregation**: Aggregate regional data
- **Global query layer**: Query across regions

## 8. Reliability Strategy

### 8.1 Data Durability
- **WAL**: Crash recovery
- **Replication**: Multiple TSDB instances
- **Backups**: Regular snapshots
- **Remote storage**: Long-term storage in object storage

### 8.2 High Availability
- **Redundant scrapers**: Multiple scraper instances
- **TSDB replication**: Multiple TSDB instances
- **Alert manager HA**: Multiple alert manager instances
- **Load balancing**: Distribute traffic

### 8.3 Monitoring
- **TSDB health**: Monitor ingestion rate, query latency
- **Scrapers**: Monitor scrape success rate
- **Alert manager**: Monitor alert evaluation
- **Cardinality**: Monitor series count

### 8.4 Disaster Recovery
- **Backups**: Regular TSDB backups
- **Remote storage**: Data in object storage
- **Failover testing**: Regular disaster recovery drills
- **Documentation**: Clear runbooks

## 9. Tradeoffs

### 9.1 Pull vs Push
| Aspect | Pull | Push |
|--------|------|------|
| Simplicity | Simple | Complex |
| Health Checks | Implicit | Explicit |
| Ephemeral Jobs | Poor | Good |
| Network Restrictions | Yes | No |

**Decision: Pull for long-running services, Push for ephemeral jobs**

### 9.2 Resolution vs Storage
| Resolution | Accuracy | Storage Cost | Use Case |
|------------|----------|--------------|----------|
| High (15s) | Excellent | High | Debugging |
| Medium (5m) | Good | Medium | Standard |
| Low (1h) | Fair | Low | Trends |

**Decision: Multi-resolution with downsampling**

### 9.3 Immediate vs Delayed Alerting
| Latency | Response Time | Noise | Use Case |
|---------|---------------|-------|----------|
| Immediate | Fast | High | Critical |
| Delayed | Slow | Low | Standard |

**Decision: Immediate for critical, delayed for standard**

### 9.4 Single Node vs Cluster
| Architecture | Complexity | Scale | Cost |
|--------------|------------|-------|------|
| Single Node | Simple | Limited | Low |
| Cluster | Complex | High | High |

**Decision: Start single, scale to cluster as needed**

## 10. Bottlenecks

### 10.1 Current Bottlenecks
1. **High cardinality**: Memory and CPU exhaustion
2. **Ingestion overload**: Dropped samples during spikes
3. **Query performance**: Slow queries on large datasets
4. **Storage growth**: Disk space exhaustion

### 10.2 Mitigation Strategies

#### High Cardinality Bottleneck
- Cardinality limits and alerts
- Label value whitelisting
- Relabeling to drop high-cardinality labels
- Series limits per metric

#### Ingestion Overload Bottleneck
- Rate limiting at scraper
- Back-pressure handling
- Queue buffering for push ingestion
- Auto-scaling scrapers

#### Query Performance Bottleneck
- Query optimization
- Result caching
- Query parallelization
- Downsampling for historical queries

#### Storage Growth Bottleneck
- Retention policies
- Automated downsampling
- Storage monitoring and alerts
- Tiered storage (hot/warm/cold)

## 11. Follow-ups

### 11.1 Custom Metrics
**Question**: How do you handle custom metric types?
**Answer**:
- Extensible metric type system
- Custom aggregation functions
- Plugin architecture
- Validation and sanitization
- Documentation and examples

### 11.2 Multi-tenancy
**Question**: How do you support multiple tenants?
**Answer**:
- Tenant isolation at ingestion
- Per-tenant resource limits
- Tenant-specific query routing
- Billing and metering
- Authentication and authorization

### 11.3 Exemplars
**Question**: How do you support exemplars for distributed tracing?
**Answer**:
- Exemplar storage with samples
- Trace ID integration
- Query support for exemplars
- Visualization in dashboards
- Integration with tracing systems

### 11.4 Recording Rules
**Question**: How do you optimize frequently used queries?
**Answer**:
- Recording rules for pre-computation
- Rule evaluation intervals
- Rule dependency management
- Storage of pre-computed results
- Query rewriting

### 11.5 Remote Storage
**Question**: How do you integrate with long-term storage?
**Answer**:
- Remote write protocol
- Object storage integration (S3, GCS)
- Compression and encoding
- Query federation
- Cost optimization

### 11.6 Service Discovery
**Question**: How do you discover targets dynamically?
**Answer**:
- Integration with service discovery (Consul, Kubernetes)
- DNS-based discovery
- File-based discovery
- API-based discovery
- Relabeling for target configuration

### 11.7 Metric Relabeling
**Question**: How do you transform metrics at ingestion?
**Answer**:
- Relabeling configuration
- Label manipulation
- Metric name changes
- Sample dropping
- Label value normalization

### 11.8 Query Optimization
**Question**: How do you optimize slow queries?
**Answer**:
- Query analysis and profiling
- Index optimization
- Query rewriting
- Result caching
- Pre-computation with recording rules

### 11.9 Alert Deduplication
**Question**: How do you prevent alert storms?
**Answer**:
- Alert grouping
- Inhibition rules
- Silencing mechanisms
- Rate limiting
- Alert routing optimization

### 11.10 Visualization
**Question**: How do you integrate with visualization tools?
**Answer**:
- Grafana integration
- Query API compatibility
- Dashboard templates
- Annotation support
- Export and sharing

---

## Summary

The metrics system is designed as a high-performance time-series database:
- **Pull-based collection** for simplicity and implicit health checks
- **TSDB storage** with Gorilla compression (100x compression ratio)
- **Label-based data model** for multi-dimensional querying
- **Separate alerting** with rule evaluation and notification routing
- **Cardinality management** to prevent resource exhaustion
- **Horizontal scaling** through federation and remote write

The system balances ingestion rate, query performance, and storage efficiency through careful design choices around collection models, compression, and downsampling. The key challenge is managing label cardinality while providing flexible querying capabilities.