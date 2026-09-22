# Distributed Cache System Design

## 1. Clarify Requirements

### Functional Requirements
- Store key-value pairs with configurable TTL
- Support basic operations: GET, SET, DELETE
- Support complex data structures (lists, sets, sorted sets, hashes)
- Automatic eviction when memory is full
- Replication for high availability
- Horizontal scaling
- Cluster management (add/remove nodes)
- Data partitioning across nodes

### Non-Functional Requirements
- **Low Latency**: Sub-millisecond response time
- **High Throughput**: Millions of operations per second
- **Scalability**: Add/remove nodes with minimal data movement
- **Availability**: Survive node failures without data loss
- **Consistency**: Configurable consistency levels
- **Durability**: Optional persistence for critical data

### Clarifying Questions
- What is the expected data size and growth rate?
- What is the required read/write ratio?
- What are the latency requirements (P50, P99)?
- Do we need persistence or is in-memory only acceptable?
- What are the consistency requirements (strong vs eventual)?
- Do we need to support complex data structures?
- What is the expected failure rate and recovery time?

## 2. Estimate Scale

### Traffic Assumptions
- 10M operations per second
- 80% reads, 20% writes
- Average key size: 50 bytes
- Average value size: 1KB
- 100M unique keys
- 70% hit rate target

### Storage Requirements
- Total data: 100M × (50 + 1024) bytes = 107.5GB
- With metadata overhead: 150GB
- With replication (factor 3): 450GB
- Memory per node: 100GB → 5 nodes minimum

### Bandwidth Requirements
- Read operations: 8M ops/s × 1KB = 8GB/s
- Write operations: 2M ops/s × 1KB = 2GB/s
- Total bandwidth: 10GB/s
- Replication traffic: 2M × 1KB × 2 = 4GB/s
- Total network: 14GB/s

### Latency Requirements
- P50: < 1ms
- P99: < 5ms
- P99.9: < 10ms

### Node Count Estimation
- Single node capacity: 200K ops/s, 100GB memory
- Required ops/s: 10M
- Nodes for throughput: 10M / 200K = 50 nodes
- Nodes for storage: 450GB / 100GB = 5 nodes
- **Total: 50 nodes (throughput limited)**

## 3. APIs

### Basic Operations
```
GET /key
Response: value or null

SET /key
Request: { "value": "data", "ttl": 3600 }
Response: "OK"

DELETE /key
Response: "OK" or "NOT_FOUND"
```

### Data Structure Operations
```
# List operations
LPUSH /listkey value
RPUSH /listkey value
LPOP /listkey
RPOP /listkey
LRANGE /listkey 0 -1

# Set operations
SADD /setkey member
SREM /setkey member
SMEMBERS /setkey

# Sorted Set operations
ZADD /zsetkey score member
ZREM /zsetkey member
ZRANGE /zsetkey 0 -1
```

### Cluster Management
```
POST /cluster/nodes
Request: { "node_address": "cache-node-1:6379" }
Response: { "node_id": "node1", "slots": [0, 16383] }

DELETE /cluster/nodes/:node_id
Response: "OK"

GET /cluster/info
Response: {
  "nodes": 5,
  "slots_assigned": 16384,
  "memory_used": "450GB",
  "ops_per_second": 10M
}
```

### Configuration
```
POST /config
Request: {
  "eviction_policy": "allkeys-lru",
  "max_memory": "100GB",
  "replication_factor": 3
}
Response: "OK"
```

## 4. Data Model

### Key-Value Storage
```
Key: string (max 250MB)
Value: string, list, set, sorted set, hash, stream
TTL: seconds (optional)
```

### Cluster Metadata
```sql
CREATE TABLE cluster_nodes (
    node_id VARCHAR(50) PRIMARY KEY,
    address VARCHAR(255) NOT NULL,
    slots INT[] NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    last_heartbeat TIMESTAMP DEFAULT NOW()
);

CREATE TABLE slot_mapping (
    slot INT PRIMARY KEY,
    primary_node VARCHAR(50) NOT NULL,
    replica_nodes VARCHAR(50)[],
    FOREIGN KEY (primary_node) REFERENCES cluster_nodes(node_id)
);
```

### Replication Metadata
```sql
CREATE TABLE replication_info (
    key_hash VARCHAR(64) PRIMARY KEY,
    primary_node VARCHAR(50) NOT NULL,
    replica_nodes VARCHAR(50)[],
    version BIGINT NOT NULL,
    last_updated TIMESTAMP DEFAULT NOW()
);
```

### Persistence Metadata (if using RDB/AOF)
```sql
CREATE TABLE persistence_info (
    node_id VARCHAR(50) PRIMARY KEY,
    last_save_timestamp TIMESTAMP,
    save_type VARCHAR(10), -- 'RDB' or 'AOF'
    file_location VARCHAR(255),
    file_size BIGINT
);
```

## 5. High Level Design

```
┌─────────────────────────────────────────────────────────────┐
│                    Client Applications                       │
│              (with client library for routing)               │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    Client Library                            │
│              (Consistent hashing, connection pool)           │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                   Cache Cluster                              │
│                                                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │  Node 1  │  │  Node 2  │  │  Node 3  │  │  Node 4  │  │
│  │ Slots:   │  │ Slots:   │  │ Slots:   │  │ Slots:   │  │
│  │ 0-4095   │  │ 4096-8191│  │ 8192-12287│ │12288-16383│ │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘  │
│       │             │             │             │           │
│       └─────────────┴─────────────┴─────────────┘           │
│                      Replication                             │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                  Optional Persistence                         │
│              (RDB snapshots, AOF logs)                        │
└─────────────────────────────────────────────────────────────┘
```

### Key Design Decisions
- **Client-side routing**: Client library handles consistent hashing
- **Consistent hashing**: Minimal data movement on topology changes
- **Asynchronous replication**: Low latency, eventual consistency
- **Configurable eviction**: Multiple eviction policies
- **Optional persistence**: For critical data

## 6. Deep Dive Components

### 6.1 Consistent Hashing

#### Ring-based Approach
- Place nodes on a ring (0 to 2^32)
- Hash key to ring position
- Key belongs to next node clockwise

#### Virtual Nodes
- Each physical node has multiple virtual nodes on ring
- Improves load distribution
- Easier rebalancing

#### Implementation
```python
def get_node(key, ring):
    key_hash = hash(key) % 2^32
    # Find next node clockwise
    for node_hash in sorted(ring.keys()):
        if node_hash >= key_hash:
            return ring[node_hash]
    return ring[min(ring.keys())]  # Wrap around
```

#### Benefits
- Minimal data movement on node addition/removal
- Only ~1/N keys move when N changes
- Better load distribution with virtual nodes

### 6.2 Replication Strategy

#### Asynchronous Replication (Recommended)
- Write acknowledged immediately
- Replicas updated in background
- Eventual consistency
- Lower latency
- Higher availability

#### Synchronous Replication
- Write acknowledged after all replicas confirm
- Strong consistency
- Higher latency
- Lower availability

#### Replication Flow
1. Client writes to primary node
2. Primary acknowledges immediately (async) or waits for replicas (sync)
3. Primary propagates write to replicas
4. Replicas acknowledge write
5. Primary tracks replication lag

#### Failure Detection
- Regular heartbeat between nodes
- Failure detector identifies failed nodes
- Automatic promotion of replica to primary
- Re-replication to maintain replication factor

### 6.3 Eviction Policies

#### LRU (Least Recently Used)
- Evict least recently accessed items
- Good for temporal locality
- Implementation: Linked list + hash map

#### LFU (Least Frequently Used)
- Evict least frequently accessed items
- Good for stable access patterns
- Implementation: Frequency counter

#### W-TinyLFU (Recommended)
- Combination of LRU and LFU
- Uses frequency sketch for recent history
- Excellent for real-world workloads
- Adapts to changing patterns

#### TTL-based Eviction
- Evict expired keys
- Can be combined with LRU/LFU
- Passive vs active expiration

#### Implementation Choice
```python
def evict_when_full():
    if memory_used > max_memory:
        if policy == 'lru':
            evict_lru()
        elif policy == 'lfu':
            evict_lfu()
        elif policy == 'w-tinylfu':
            evict_w_tinylfu()
```

### 6.4 Cache Patterns

#### Cache-Aside (Lazy Loading)
```python
def get(key):
    value = cache.get(key)
    if value is None:
        value = database.get(key)
        cache.set(key, value)
    return value

def set(key, value):
    database.set(key, value)
    cache.delete(key)  # or cache.set(key, value)
```

**Pros**: Only cache requested data, simple
**Cons**: Cache miss penalty, stale data possible

#### Write-Through
```python
def set(key, value):
    cache.set(key, value)
    database.set(key, value)
    return value
```

**Pros**: Strong consistency, data always in cache
**Cons**: Higher write latency, unnecessary cache writes

#### Write-Behind (Write-Back)
```python
def set(key, value):
    cache.set(key, value)
    async_write_to_database(key, value)
    return value
```

**Pros**: Lowest write latency, can batch writes
**Cons**: Data loss risk on cache failure

**Recommendation: Cache-Aside for most scenarios**

### 6.5 Hot Key Mitigation

#### Problem
Single key receives disproportionate traffic, overwhelming single node

#### Solutions

##### Replication
- Replicate hot key across multiple nodes
- Client-side load balancing
- Read from any replica

##### Request Coalescing
- Single flight pattern
- Multiple requests share same computation
```python
def get_with_coalescing(key):
    if key in flight:
        return flight[key].result()
    flight[key] = Future()
    result = fetch_from_cache(key)
    flight[key].set_result(result)
    return result
```

##### Local Caching
- Client-side cache for hot keys
- Reduce load on distributed cache
- Short TTL to maintain consistency

##### Partition Splitting
- Split hot key into multiple keys
- Aggregate on read
- Use hash suffix for partitioning

### 6.6 Cache Stampede Prevention

#### Problem
Many requests miss cache simultaneously, overwhelming backend

#### Solutions

##### Lease-Based Locking
- First request gets lease to populate cache
- Other requests wait or return stale data
```python
def get_with_lease(key):
    value = cache.get(key)
    if value is None:
        lease = cache.acquire_lease(key)
        if lease:
            value = database.get(key)
            cache.set(key, value)
            cache.release_lease(key)
        else:
            # Wait for lease holder or return stale
            value = wait_for_lease(key)
    return value
```

##### Probabilistic Early Expiration
- Expire hot keys slightly before TTL
- Proactive refresh
- Reduces thundering herd

##### Request Batching
- Combine multiple cache misses
- Single backend query
- Distribute results

## 7. Scaling Strategy

### 7.1 Horizontal Scaling

#### Adding Nodes
1. Add new node to cluster
2. Redistribute slots using consistent hashing
3. Migrate data to new node
4. Update client routing tables
5. Remove migration status

#### Removing Nodes
1. Mark node for removal
2. Redistribute slots to remaining nodes
3. Migrate data from leaving node
4. Update client routing tables
5. Shut down node

#### Rebalancing
- Automatic rebalancing based on load
- Minimal data movement
- Gradual migration to avoid performance impact

### 7.2 Vertical Scaling
- Increase node memory capacity
- Upgrade CPU for better performance
- Use SSD for persistence
- Optimize network configuration

### 7.3 Geographic Distribution
- **Multi-region deployment**: Deploy in multiple regions
- **Data locality**: Cache data close to users
- **Cross-region replication**: For disaster recovery
- **Regional read replicas**: For low-latency reads

## 8. Reliability Strategy

### 8.1 Replication
- **Replication factor**: 3 for production
- **Async replication**: For low latency
- **Automatic failover**: Promote replica on primary failure
- **Re-replication**: Maintain replication factor

### 8.2 Failure Detection
- **Heartbeat**: Regular health checks between nodes
- **Failure detector**: Identify failed nodes
- **Quorum**: Require majority for decisions
- **Timeouts**: Configurable failure detection timeouts

### 8.3 Data Persistence
- **RDB snapshots**: Periodic point-in-time snapshots
- **AOF logs**: Append-only file for durability
- **Configurable persistence**: Based on data criticality
- **Recovery**: Replay AOF or load RDB on restart

### 8.4 Monitoring
- **Metrics**: Ops/sec, hit rate, memory usage, latency
- **Alerts**: High memory, low hit rate, node failures
- **Health checks**: Node availability, replication lag
- **Logging**: Detailed logging for troubleshooting

## 9. Tradeoffs

### 9.1 Redis vs Memcached
| Aspect | Redis | Memcached |
|--------|-------|-----------|
| Data Structures | Rich (lists, sets, etc.) | Strings only |
| Persistence | Yes (RDB, AOF) | No |
| Replication | Yes | No |
| Multi-threaded | Limited (I/O threads) | Yes |
| Memory Efficiency | Good | Excellent |
| Features | Pub/sub, Lua, transactions | Basic |

**Decision: Redis for most use cases, Memcached for simple high-throughput caching**

### 9.2 Consistency vs Availability
| Consistency | Availability | Use Case |
|-------------|---------------|----------|
| Strong | Lower | Critical data |
| Eventual | Higher | Most caching scenarios |

**Decision: Eventual consistency for caches**

### 9.3 Memory vs Disk
| Storage | Latency | Capacity | Use Case |
|---------|---------|----------|----------|
| Memory-only | Lowest | Limited | Hot data |
| Memory + Disk | Higher | Larger | Warm data |
| Disk-only | Highest | Largest | Cold data |

**Decision: Memory-only with optional persistence**

### 9.4 Client-side vs Server-side Routing
| Routing | Complexity | Scalability | Failure Handling |
|---------|------------|-------------|------------------|
| Client-side | Client library | Excellent | Client updates needed |
| Server-side | Proxy | Good | Centralized management |

**Decision: Client-side for performance, server-side for simplicity**

## 10. Bottlenecks

### 10.1 Current Bottlenecks
1. **Network bandwidth**: High throughput requirements
2. **Single node capacity**: Memory and CPU limits
3. **Hot keys**: Disproportionate load on single node
4. **Replication lag**: Async replication delay

### 10.2 Mitigation Strategies

#### Network Bandwidth Bottleneck
- Data compression
- Batch operations
- Connection pooling
- Optimized serialization

#### Single Node Capacity Bottleneck
- Horizontal scaling with sharding
- Vertical scaling with better hardware
- Data partitioning
- Load balancing

#### Hot Key Bottleneck
- Hot key replication
- Request coalescing
- Local caching
- Partition splitting

#### Replication Lag Bottleneck
- Monitor replication lag
- Optimize network between nodes
- Increase replication bandwidth
- Use synchronous replication for critical data

## 11. Follow-ups

### 11.1 Data Persistence
**Question**: How do you handle data persistence?
**Answer**:
- RDB snapshots for point-in-time recovery
- AOF logs for durability
- Configurable persistence strategies
- Recovery process on restart
- Trade-off between performance and durability

### 11.2 Cache Invalidation
**Question**: How do you handle cache invalidation?
**Answer**:
- TTL-based expiration
- Manual invalidation on data changes
- Write-through for critical data
- Cache invalidation patterns
- Handling stale data

### 11.3 Multi-tenancy
**Question**: How do you support multiple tenants?
**Answer**:
- Namespace isolation
- Per-tenant resource limits
- Tenant-specific configurations
- Authentication and authorization
- Billing and metering

### 11.4 Security
**Question**: How do you secure the cache?
**Answer**:
- TLS for network encryption
- Authentication (SASL, ACL)
- Authorization (per-key permissions)
- Network isolation (VPC)
- Encryption at rest

### 11.5 Monitoring
**Question**: How do you monitor cache performance?
**Answer**:
- Metrics collection (ops/sec, hit rate, latency)
- Distributed tracing
- Alerting on anomalies
- Performance profiling
- Capacity planning

### 11.6 Disaster Recovery
**Question**: How do you handle cache cluster failures?
**Answer**:
- Multi-region deployment
- Cross-region replication
- Automatic failover
- Data recovery from persistence
- Testing and drills

### 11.7 Client Library
**Question**: What features should the client library provide?
**Answer**:
- Consistent hashing
- Connection pooling
- Retry logic
- Failover handling
- Load balancing
- Metrics collection

### 11.8 Cold Start
**Question**: How do you handle cache cold start?
**Answer**:
- Cache warming strategies
- Gradual traffic ramp-up
- Fallback to database
- Pre-loading critical data
- Monitoring hit rate during warm-up

### 11.9 Memory Fragmentation
**Question**: How do you handle memory fragmentation?
**Answer**:
- Memory allocators (jemalloc, tcmalloc)
- Slab allocation (Memcached-style)
- Regular memory defragmentation
- Monitoring fragmentation ratio
- Restart to reclaim memory

### 11.10 Large Values
**Question**: How do you handle large values?
**Answer**:
- Value size limits
- Chunking large values
- Compression for large values
- Separate storage for large values
- Memory vs disk storage for large values

---

## Summary

The distributed cache system is designed for high throughput and low latency:
- **Consistent hashing** for minimal data movement during scaling
- **Asynchronous replication** for low latency and high availability
- **Multiple eviction policies** for different workload patterns
- **Hot key mitigation** to prevent single-node overload
- **Cache stampede prevention** to protect backend systems
- **Horizontal scaling** through adding nodes with automatic rebalancing

The system balances performance, scalability, and reliability through careful design choices around consistency, replication, and data distribution.