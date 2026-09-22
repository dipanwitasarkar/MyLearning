# Data Partitioning

## Overview
Data partitioning is the process of dividing a large dataset into smaller, more manageable parts called partitions. This technique is fundamental for scaling databases and distributed systems, enabling horizontal scaling, improved performance, and better resource utilization.

## Why Data Partitioning?

### Benefits
- **Scalability**: Distribute data across multiple servers
- **Performance**: Reduce query latency by working with smaller datasets
- **Availability**: Partition failures don't affect entire system
- **Manageability**: Easier to maintain and backup smaller datasets
- **Cost Efficiency**: Optimize resource usage

### When to Use
- Large datasets that don't fit on a single server
- High query throughput requirements
- Geographic distribution needs
- Multi-tenant applications
- Real-time analytics

## Partitioning Strategies

### 1. Horizontal Partitioning (Sharding)
Split rows of a table across multiple servers.

**Example:**
```
Users Table (1 billion rows)
├── Shard 1: Users 1-100M (Server A)
├── Shard 2: Users 100M-200M (Server B)
├── Shard 3: Users 200M-300M (Server C)
└── Shard 4: Users 300M-400M (Server D)
```

**Pros:**
- Scales write throughput
- Natural load distribution
- Simple to understand

**Cons:**
- Cross-shard queries are complex
- Rebalancing is challenging
- Hot spots possible

**Best For:**
- Time-series data
- User-specific data
- Write-heavy workloads

### 2. Vertical Partitioning
Split columns of a table across multiple servers.

**Example:**
```
Users Table
├── Server A: user_id, name, email (frequently accessed)
├── Server B: user_id, bio, preferences (less frequently accessed)
└── Server C: user_id, logs, history (rarely accessed)
```

**Pros:**
- Optimizes for access patterns
- Reduces I/O for queries
- Better memory utilization

**Cons:**
- Cross-partition joins needed
- More complex schema
- Rebalancing complexity

**Best For:**
- Tables with many columns
- Varying access patterns
- Memory-constrained environments

### 3. Directory-Based Partitioning
Maintain a lookup service to track partition locations.

**Architecture:**
```
Application → Directory Service → Partition Location → Data Shard
```

**Implementation:**
```python
class DirectoryService:
    def __init__(self):
        self.partition_map = {}
    
    def get_partition(self, key):
        return self.partition_map.get(key)
    
    def add_partition(self, key, location):
        self.partition_map[key] = location
    
    def remove_partition(self, key):
        del self.partition_map[key]
```

**Pros:**
- Flexible partition mapping
- Easy to rebalance
- Can handle complex routing

**Cons:**
- Additional lookup overhead
- Directory service becomes bottleneck
- Single point of failure

**Best For:**
- Complex routing requirements
- Dynamic partitioning
- Multi-tenant systems

### 4. Hash Partitioning
Apply a hash function to the partitioning key to determine partition.

**Algorithm:**
```python
def get_partition(key, num_partitions):
    hash_value = hash(key)
    return hash_value % num_partitions

# Example
partition = get_partition("user123", 4)  # Returns 0-3
```

**Pros:**
- Even data distribution
- No hot spots (with good hash function)
- Simple to implement
- Predictable routing

**Cons:**
- Difficult to do range queries
- Rebalancing requires remapping
- No locality preservation

**Best For:**
- Even distribution requirements
- Key-based access patterns
- Distributed caches

### 5. Range Partitioning
Partition based on value ranges of the partitioning key.

**Example:**
```
Users Table partitioned by user_id range:
├── Partition 1: user_id 1-1000
├── Partition 2: user_id 1001-2000
├── Partition 3: user_id 2001-3000
└── Partition 4: user_id 3001-4000
```

**Implementation:**
```python
def get_range_partition(key, ranges):
    for partition_id, (start, end) in ranges.items():
        if start <= key <= end:
            return partition_id
    return None  # Key not in any range

ranges = {
    1: (1, 1000),
    2: (1001, 2000),
    3: (2001, 3000),
    4: (3001, 4000)
}
```

**Pros:**
- Range queries are efficient
- Data locality preserved
- Simple to understand

**Cons:**
- Uneven distribution possible
- Hot spots in popular ranges
- Rebalancing complexity

**Best For:**
- Time-series data
- Range query workloads
- Data with natural ordering

### 6. Consistent Hashing
Distribute data across a changing set of servers using a hash ring.

**Algorithm:**
```python
import hashlib

class ConsistentHash:
    def __init__(self, replicas=100):
        self.replicas = replicas
        self.ring = {}
        self.sorted_keys = []
    
    def add_node(self, node):
        for i in range(self.replicas):
            key = f"{node}:{i}"
            hash_value = int(hashlib.md5(key.encode()).hexdigest(), 16)
            self.ring[hash_value] = node
            self.sorted_keys.append(hash_value)
        self.sorted_keys.sort()
    
    def remove_node(self, node):
        for i in range(self.replicas):
            key = f"{node}:{i}"
            hash_value = int(hashlib.md5(key.encode()).hexdigest(), 16)
            del self.ring[hash_value]
            self.sorted_keys.remove(hash_value)
    
    def get_node(self, key):
        if not self.ring:
            return None
        
        hash_value = int(hashlib.md5(key.encode()).hexdigest(), 16)
        
        # Find first node with hash >= key hash
        for node_hash in self.sorted_keys:
            if node_hash >= hash_value:
                return self.ring[node_hash]
        
        # Wrap around to first node
        return self.ring[self.sorted_keys[0]]
```

**Pros:**
- Minimal data movement on topology changes
- Even distribution with virtual nodes
- Handles node additions/removals gracefully

**Cons:**
- More complex implementation
- Requires virtual nodes for even distribution
- Slightly higher computational overhead

**Best For:**
- Dynamic environments
- Distributed caches
- Systems with frequent topology changes

## Partitioning Key Selection

### Guidelines for Choosing Partition Keys

#### 1. Even Distribution
Choose keys that distribute data evenly.

**Good:**
- User ID (random distribution)
- Order ID (sequential but spread across users)
- Random UUIDs

**Bad:**
- Status (few distinct values)
- Country (uneven population)
- Zip code (clustered)

#### 2. Query Patterns
Align partitioning with common query patterns.

**Example:**
```sql
-- If most queries are by user_id, partition by user_id
SELECT * FROM orders WHERE user_id = 123;

-- If most queries are by date, partition by date
SELECT * FROM orders WHERE created_at > '2024-01-01';
```

#### 3. Data Locality
Group related data together.

**Example:**
```python
# Partition by user_id to keep all user data together
user_data = {
    'profile': partition_by(user_id),
    'orders': partition_by(user_id),
    'preferences': partition_by(user_id)
}
```

#### 4. Future Growth
Consider how data will grow over time.

**Example:**
```python
# Bad: Fixed ranges that may become uneven
partition = user_id // 1000000

# Good: Hash-based that scales
partition = hash(user_id) % num_partitions
```

## Cross-Partition Queries

### Challenges
- Queries spanning multiple partitions
- Join operations across partitions
- Aggregation across partitions
- Transaction management

### Solutions

#### 1. Distributed Query Processing
Execute queries across partitions and combine results.

**Implementation:**
```python
def distributed_query(query, partitions):
    results = []
    for partition in partitions:
        result = execute_query(query, partition)
        results.append(result)
    return combine_results(results)
```

#### 2. Two-Phase Query
First phase: scatter queries to partitions
Second phase: gather and combine results

**Example:**
```python
def scatter_gather_query(query, partitions):
    # Scatter phase
    futures = [execute_async(query, partition) for partition in partitions]
    
    # Gather phase
    results = [future.result() for future in futures]
    return aggregate_results(results)
```

#### 3. Query Routing Service
Intelligent routing based on query analysis.

**Implementation:**
```python
class QueryRouter:
    def route_query(self, query):
        if is_single_partition(query):
            return get_target_partition(query)
        else:
            return get_all_partitions()
```

#### 4. Denormalization
Duplicate data to avoid cross-partition joins.

**Example:**
```sql
-- Instead of joining across partitions
SELECT u.name, o.total 
FROM users u 
JOIN orders o ON u.id = o.user_id 
WHERE o.id = 123;

-- Denormalize order data
SELECT name, total 
FROM orders 
WHERE id = 123;
-- (orders table includes user_name)
```

## Partition Rebalancing

### When to Rebalance
- Add/remove nodes
- Uneven data distribution
- Hot spot mitigation
- Performance optimization

### Rebalancing Strategies

#### 1. Offline Rebalancing
Stop writes, move data, restart system.

**Pros:**
- Simple to implement
- Consistent state
- No data duplication

**Cons:**
- Downtime required
- Not suitable for 24/7 systems
- User impact

**Best For:**
- Planned maintenance
- Non-critical systems
- Small datasets

#### 2. Online Rebalancing
Move data while system continues operating.

**Implementation:**
```python
def online_rebalance(from_partition, to_partition, key_range):
    # 1. Start dual-write to both partitions
    enable_dual_write(from_partition, to_partition, key_range)
    
    # 2. Migrate existing data
    migrate_data(from_partition, to_partition, key_range)
    
    # 3. Update routing
    update_routing(key_range, to_partition)
    
    # 4. Stop dual-write
    disable_dual_write(from_partition, to_partition, key_range)
```

**Pros:**
- No downtime
- Minimal user impact
- Continuous availability

**Cons:**
- Complex implementation
- Temporary data duplication
- Performance impact during migration

**Best For:**
- 24/7 systems
- Large datasets
- Production environments

#### 3. Consistent Hashing Rebalancing
Add/remove nodes with minimal data movement.

**Benefits:**
- Only ~1/N data moves when N changes
- Automatic rebalancing
- Minimal disruption

**Implementation:**
```python
# Adding a node
consistent_hash.add_node("new_node")
# Only ~1/N keys need to move

# Removing a node
consistent_hash.remove_node("old_node")
# Only ~1/N keys need to move
```

## Partitioning in Different Systems

### 1. Database Partitioning

#### MySQL Partitioning
```sql
CREATE TABLE orders (
    id INT,
    user_id INT,
    created_at DATETIME,
    amount DECIMAL(10,2)
)
PARTITION BY RANGE (YEAR(created_at)) (
    PARTITION p2023 VALUES LESS THAN (2024),
    PARTITION p2024 VALUES LESS THAN (2025),
    PARTITION p2025 VALUES LESS THAN (2026)
);
```

#### PostgreSQL Partitioning
```sql
CREATE TABLE orders (
    id SERIAL,
    user_id INT,
    created_at TIMESTAMP,
    amount DECIMAL(10,2)
) PARTITION BY RANGE (created_at);

CREATE TABLE orders_2023 PARTITION OF orders
    FOR VALUES FROM ('2023-01-01') TO ('2024-01-01');

CREATE TABLE orders_2024 PARTITION OF orders
    FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');
```

#### MongoDB Sharding
```javascript
sh.shardCollection("mydb.orders", {
    "user_id": "hashed"  // Hash-based sharding
})
```

### 2. Distributed File Systems

#### HDFS Partitioning
```python
# HDFS automatically partitions data into blocks
# Default block size: 128MB
# Blocks are distributed across DataNodes
```

#### Cassandra Partitioning
```sql
-- Partition key determines data distribution
CREATE TABLE users (
    user_id UUID PRIMARY KEY,
    name TEXT,
    email TEXT
);
-- Data partitioned by user_id
```

### 3. Message Queue Partitioning

#### Kafka Partitioning
```python
# Messages with same key go to same partition
producer.send(
    topic='orders',
    key='user123',  # Partitioning key
    value=order_data
)
```

## Partitioning Patterns

### 1. Geo-Partitioning
Partition data based on geographic location.

**Example:**
```
Users Table
├── US Partition: Users in United States
├── EU Partition: Users in Europe
└── Asia Partition: Users in Asia
```

**Benefits:**
- Data residency compliance
- Reduced latency for local users
- Regulatory compliance

**Implementation:**
```python
def get_geo_partition(user):
    country = get_user_country(user.id)
    if country in ['US', 'CA', 'MX']:
        return 'us_partition'
    elif country in ['GB', 'DE', 'FR']:
        return 'eu_partition'
    else:
        return 'asia_partition'
```

### 2. Time-Based Partitioning
Partition data based on time ranges.

**Example:**
```
Orders Table
├── 2023_Q1: Jan-Mar 2023
├── 2023_Q2: Apr-Jun 2023
├── 2023_Q3: Jul-Sep 2023
└── 2023_Q4: Oct-Dec 2023
```

**Benefits:**
- Efficient time-range queries
- Easy data archival
- Natural data lifecycle management

**Implementation:**
```python
def get_time_partition(timestamp):
    quarter = (timestamp.month - 1) // 3 + 1
    return f"{timestamp.year}_Q{quarter}"
```

### 3. Tenant-Based Partitioning
Partition data by tenant in multi-tenant systems.

**Example:**
```
Multi-tenant SaaS
├── Tenant A Partition: All data for Tenant A
├── Tenant B Partition: All data for Tenant B
└── Tenant C Partition: All data for Tenant C
```

**Benefits:**
- Tenant isolation
- Per-tenant performance optimization
- Easy tenant migration

**Implementation:**
```python
def get_tenant_partition(tenant_id):
    return f"tenant_{tenant_id}"
```

## Monitoring Partitioned Systems

### Key Metrics
- **Partition Size**: Monitor data distribution
- **Query Latency**: Per-partition query performance
- **Hot Spots**: Identify uneven load distribution
- **Rebalancing Progress**: Track migration status
- **Cross-Partition Queries**: Monitor distributed query performance

### Implementation
```python
class PartitionMonitor:
    def __init__(self):
        self.metrics = {}
    
    def record_partition_size(self, partition_id, size):
        self.metrics[f'partition_{partition_id}_size'] = size
    
    def record_query_latency(self, partition_id, latency):
        self.metrics[f'partition_{partition_id}_latency'] = latency
    
    def detect_hot_spots(self):
        sizes = {k: v for k, v in self.metrics.items() if 'size' in k}
        avg_size = sum(sizes.values()) / len(sizes)
        hot_spots = [k for k, v in sizes.items() if v > avg_size * 2]
        return hot_spots
```

## Common Pitfalls

### 1. Uneven Distribution
Poor partition key choice leads to hot spots.

**Solution:**
- Monitor data distribution
- Use hash-based partitioning
- Implement rebalancing

### 2. Cross-Partition Queries
Frequent cross-partition queries hurt performance.

**Solution:**
- Design schema to minimize cross-partition queries
- Use denormalization
- Implement query routing

### 3. Rebalancing Complexity
Rebalancing is complex and error-prone.

**Solution:**
- Use consistent hashing
- Plan for rebalancing in design
- Use automated tools

### 4. Partition Skew Over Time
Data distribution changes over time.

**Solution:**
- Monitor partition sizes
- Implement periodic rebalancing
- Design for growth

### 5. Transaction Complexity
Cross-partition transactions are complex.

**Solution:**
- Design for single-partition operations
- Use eventual consistency where appropriate
- Implement distributed transactions

## Best Practices

### 1. Choose Partition Keys Carefully
Consider query patterns, data distribution, and future growth.

### 2. Monitor Data Distribution
Regularly monitor partition sizes and query patterns.

### 3. Plan for Rebalancing
Design systems to handle rebalancing gracefully.

### 4. Minimize Cross-Partition Operations
Design schema and queries to minimize cross-partition operations.

### 5. Implement Fallback Strategies
Handle partition failures gracefully.

### 6. Test Partitioning Strategy
Test with realistic data and query patterns.

### 7. Document Partitioning Scheme
Maintain clear documentation of partitioning decisions.

### 8. Use Appropriate Tools
Leverage database and framework partitioning features.

## Conclusion

Data partitioning is essential for building scalable distributed systems. By understanding different partitioning strategies, choosing appropriate partition keys, and handling cross-partition queries effectively, you can design systems that scale horizontally while maintaining performance and availability. The key is to align your partitioning strategy with your specific access patterns and scalability requirements.