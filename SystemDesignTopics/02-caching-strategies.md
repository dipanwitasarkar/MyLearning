# Caching Strategies

## Overview
Caching is the process of storing frequently accessed data in a fast storage layer to reduce data retrieval time and system load. Caching is one of the most effective techniques for improving system performance and scalability.

## Why Caching?

### Benefits
- **Reduced Latency**: Data served from cache is much faster than from the original source
- **Reduced Load**: Fewer requests hit the database or backend services
- **Improved Scalability**: System can handle more traffic with the same resources
- **Cost Savings**: Reduced database usage can lower infrastructure costs
- **Better User Experience**: Faster response times improve user satisfaction

### When to Use Cache
- Read-heavy workloads
- Expensive computations
- Frequently accessed data
- Data that doesn't change frequently
- High-traffic endpoints

## Cache Types

### 1. Client-Side Caching
Cache stored on the client (browser, mobile app).

**Examples:**
- Browser cache (HTTP headers)
- Mobile app local storage
- Service Worker cache

**Implementation:**
```http
HTTP/1.1 200 OK
Cache-Control: max-age=3600
ETag: "33a64df551425fcc55e4d42a148795d9f25f89d4"
```

**Pros:**
- Reduces server load
- Improves user experience
- Works offline

**Cons:**
- Limited control
- Cache invalidation challenges
- Storage limitations

### 2. CDN Caching
Content Delivery Network caching at edge locations.

**Architecture:**
```
User → CDN Edge (Cache Hit) → Response
     → CDN Edge (Cache Miss) → Origin Server → Cache → Response
```

**Pros:**
- Global distribution
- Reduced latency
- Handles high traffic
- DDoS protection

**Cons:**
- Cost
- Cache invalidation complexity
- Limited dynamic content support

### 3. Application-Level Caching
Cache within the application layer (in-memory).

**Examples:**
- Redis, Memcached
- In-memory data structures
- Application cache

**Implementation:**
```python
import redis

cache = redis.Redis(host='localhost', port=6379, db=0)

def get_user(user_id):
    # Try cache first
    cached_user = cache.get(f'user:{user_id}')
    if cached_user:
        return json.loads(cached_user)
    
    # Cache miss - fetch from database
    user = db.query("SELECT * FROM users WHERE id = ?", user_id)
    
    # Store in cache
    cache.setex(f'user:{user_id}', 3600, json.dumps(user))
    
    return user
```

**Pros:**
- Fast access (in-memory)
- Shared across application instances
- Flexible data structures
- Rich feature set

**Cons:**
- Memory limitations
- Requires infrastructure
- Cache invalidation complexity

### 4. Database Caching
Cache at the database level (query cache, buffer pool).

**Examples:**
- MySQL query cache
- PostgreSQL buffer pool
- Database internal caching

**Pros:**
- Transparent to application
- Automatic management
- Optimized for database operations

**Cons:**
- Limited control
- Database-specific
- May not suit all use cases

## Caching Patterns

### 1. Cache-Aside (Lazy Loading)
Application checks cache first, loads from source on miss.

**Flow:**
```
1. Application requests data
2. Check cache
3. If cache hit: return cached data
4. If cache miss: fetch from source, update cache, return data
```

**Implementation:**
```python
def get_product(product_id):
    # Check cache
    product = cache.get(f'product:{product_id}')
    if product:
        return product
    
    # Cache miss - fetch from database
    product = db.query("SELECT * FROM products WHERE id = ?", product_id)
    
    # Update cache
    cache.set(f'product:{product_id}', product, ttl=3600)
    
    return product
```

**Pros:**
- Only requested data is cached
- Simple implementation
- Cache failure doesn't break application

**Cons:**
- Cache miss penalty
- Stale data possible
- Three round trips on miss

**Best For:**
- Read-heavy workloads
- On-demand caching
- General-purpose caching

### 2. Write-Through
Write to cache and database synchronously.

**Flow:**
```
1. Application writes data
2. Write to cache
3. Write to database
4. Return success
```

**Implementation:**
```python
def update_user(user_id, data):
    # Update cache
    cache.set(f'user:{user_id}', data, ttl=3600)
    
    # Update database
    db.execute("UPDATE users SET data = ? WHERE id = ?", data, user_id)
    
    return True
```

**Pros:**
- Strong consistency
- Data always in cache
- Good for write-heavy workloads

**Cons:**
- Higher write latency
- Unnecessary cache writes
- Cache becomes bottleneck

**Best For:**
- Write-heavy workloads
- Strong consistency requirements
- Frequently written data

### 3. Write-Behind (Write-Back)
Write to cache immediately, database asynchronously.

**Flow:**
```
1. Application writes data
2. Write to cache
3. Queue database write
4. Return success immediately
5. Background process writes to database
```

**Implementation:**
```python
def update_user(user_id, data):
    # Update cache
    cache.set(f'user:{user_id}', data, ttl=3600)
    
    # Queue database write
    write_queue.enqueue({
        'table': 'users',
        'id': user_id,
        'data': data
    })
    
    return True

# Background worker
def process_write_queue():
    while True:
        write_op = write_queue.dequeue()
        db.execute(f"UPDATE {write_op['table']} SET data = ? WHERE id = ?",
                   write_op['data'], write_op['id'])
```

**Pros:**
- Lowest write latency
- Can batch writes
- High write throughput

**Cons:**
- Data loss risk on cache failure
- Complex implementation
- Eventual consistency

**Best For:**
- High-write throughput
- Acceptable data loss
- Write-heavy workloads

### 4. Refresh-Ahead
Proactively refresh cache before expiration.

**Flow:**
```
1. Application requests data
2. Check cache
3. If cache hit and expiring soon: refresh in background
4. Return cached data
```

**Implementation:**
```python
def get_product(product_id):
    product = cache.get(f'product:{product_id}')
    
    if product:
        ttl = cache.ttl(f'product:{product_id}')
        # Refresh if expiring in less than 5 minutes
        if ttl < 300:
            background_refresh(product_id)
        return product
    
    # Cache miss
    product = db.query("SELECT * FROM products WHERE id = ?", product_id)
    cache.set(f'product:{product_id}', product, ttl=3600)
    return product

def background_refresh(product_id):
    product = db.query("SELECT * FROM products WHERE id = ?", product_id)
    cache.set(f'product:{product_id}', product, ttl=3600)
```

**Pros:**
- Reduces cache miss rate
- Better user experience
- Proactive cache management

**Cons:**
- Wasted resources on unused data
- Complex implementation
- May refresh unnecessarily

**Best For:**
- Hot keys
- Predictable access patterns
- Performance-critical data

## Cache Invalidation Strategies

### 1. Time-Based Expiration (TTL)
Cache entries expire after a fixed time.

**Implementation:**
```python
# Set TTL when writing to cache
cache.set('key', value, ttl=3600)  # Expires in 1 hour
```

**Pros:**
- Simple to implement
- Automatic cleanup
- No complex invalidation logic

**Cons:**
- Stale data possible
- May expire too early or too late
- Not event-driven

**Best For:**
- Data that changes infrequently
- When staleness is acceptable
- Simple use cases

### 2. Event-Based Invalidation
Invalidate cache when data changes.

**Implementation:**
```python
def update_user(user_id, data):
    # Update database
    db.execute("UPDATE users SET data = ? WHERE id = ?", data, user_id)
    
    # Invalidate cache
    cache.delete(f'user:{user_id}')
    
    return True
```

**Pros:**
- Immediate consistency
- No stale data
- Event-driven

**Cons:**
- Complex implementation
- Need to track all cache keys
- May have many invalidation events

**Best For:**
- Data that changes frequently
- When consistency is critical
- Event-driven architectures

### 3. Write-Through Invalidation
Update cache and database together.

**Implementation:**
```python
def update_user(user_id, data):
    # Update both cache and database
    cache.set(f'user:{user_id}', data, ttl=3600)
    db.execute("UPDATE users SET data = ? WHERE id = ?", data, user_id)
    
    return True
```

**Pros:**
- Strong consistency
- Simple invalidation
- Cache always up-to-date

**Cons:**
- Higher write latency
- Unnecessary cache writes
- Write amplification

**Best For:**
- Strong consistency requirements
- Frequently accessed data
- Write-through caching pattern

## Cache Eviction Policies

### 1. LRU (Least Recently Used)
Evict least recently used items when cache is full.

**Implementation:**
```python
from collections import OrderedDict

class LRUCache:
    def __init__(self, capacity):
        self.cache = OrderedDict()
        self.capacity = capacity
    
    def get(self, key):
        if key in self.cache:
            # Move to end (most recently used)
            self.cache.move_to_end(key)
            return self.cache[key]
        return None
    
    def set(self, key, value):
        if key in self.cache:
            self.cache.move_to_end(key)
        self.cache[key] = value
        if len(self.cache) > self.capacity:
            # Remove least recently used
            self.cache.popitem(last=False)
```

**Pros:**
- Good temporal locality
- Simple to understand
- Widely used

**Cons:**
- Doesn't consider access frequency
- Can be expensive to maintain
- May evict popular items that haven't been accessed recently

**Best For:**
- General-purpose caching
- Temporal locality workloads
- When simplicity is preferred

### 2. LFU (Least Frequently Used)
Evict least frequently used items.

**Implementation:**
```python
import heapq
from collections import defaultdict

class LFUCache:
    def __init__(self, capacity):
        self.capacity = capacity
        self.cache = {}
        self.freq = defaultdict(int)
        self.min_freq = 0
    
    def get(self, key):
        if key in self.cache:
            self.freq[key] += 1
            return self.cache[key]
        return None
    
    def set(self, key, value):
        if len(self.cache) >= self.capacity and key not in self.cache:
            # Evict least frequently used
            min_freq_keys = [k for k, v in self.freq.items() if v == self.min_freq]
            if min_freq_keys:
                del self.cache[min_freq_keys[0]]
                del self.freq[min_freq_keys[0]]
        
        self.cache[key] = value
        self.freq[key] = 1
        self.min_freq = 1
```

**Pros:**
- Keeps popular items
- Good for stable access patterns
- Frequency-based eviction

**Cons:**
- Complex implementation
- Slow to adapt to changing patterns
- More memory overhead

**Best For:**
- Stable access patterns
- When popularity is important
- Long-running caches

### 3. FIFO (First In First Out)
Evict oldest items first.

**Implementation:**
```python
from collections import deque

class FIFOCache:
    def __init__(self, capacity):
        self.cache = {}
        self.queue = deque()
        self.capacity = capacity
    
    def get(self, key):
        return self.cache.get(key)
    
    def set(self, key, value):
        if key in self.cache:
            return
        
        if len(self.cache) >= self.capacity:
            # Evict oldest
            oldest = self.queue.popleft()
            del self.cache[oldest]
        
        self.cache[key] = value
        self.queue.append(key)
```

**Pros:**
- Simple implementation
- Low overhead
- Predictable eviction

**Cons:**
- Doesn't consider access patterns
- May evict popular items
- Poor performance for many workloads

**Best For:**
- Simple use cases
- When access patterns are unknown
- Low-overhead requirements

## Distributed Caching

### Challenges
- Cache consistency across nodes
- Distributed cache invalidation
- Network partitions
- Scalability

### Solutions

#### 1. Cache Invalidation Broadcast
Broadcast invalidation events to all cache nodes.

**Implementation:**
```python
# Using pub/sub for invalidation
def invalidate_cache(key):
    # Local invalidation
    cache.delete(key)
    
    # Broadcast to other nodes
    pubsub.publish('cache_invalidation', key)

# Other nodes subscribe
def on_invalidation_message(key):
    cache.delete(key)
```

#### 2. Distributed Cache with Consistent Hashing
Use consistent hashing to distribute cache across nodes.

**Implementation:**
```python
import hashlib

class DistributedCache:
    def __init__(self, nodes):
        self.nodes = nodes
        self.ring = {}
        self._build_ring()
    
    def _build_ring(self):
        for node in self.nodes:
            for i in range(100):  # Virtual nodes
                key = f"{node}:{i}"
                hash_value = int(hashlib.md5(key.encode()).hexdigest(), 16)
                self.ring[hash_value] = node
    
    def get_node(self, key):
        hash_value = int(hashlib.md5(key.encode()).hexdigest(), 16)
        # Find next node clockwise
        for node_hash in sorted(self.ring.keys()):
            if node_hash >= hash_value:
                return self.ring[node_hash]
        return self.ring[min(self.ring.keys())]
```

#### 3. Cache Coherency Protocols
Implement protocols to maintain cache consistency.

**Examples:**
- Write-invalidate
- Write-update
- Write-broadcast

## Cache Performance Optimization

### 1. Batch Operations
Group multiple cache operations for efficiency.

```python
def get_multiple_products(product_ids):
    # Use MGET for multiple keys
    cached_products = cache.mget([f'product:{pid}' for pid in product_ids])
    
    # Fetch missing products
    missing_ids = [pid for pid, cached in zip(product_ids, cached_products) if cached is None]
    if missing_ids:
        products = db.query("SELECT * FROM products WHERE id IN (?)", missing_ids)
        # Cache the fetched products
        pipe = cache.pipeline()
        for product in products:
            pipe.set(f'product:{product.id}', product, ttl=3600)
        pipe.execute()
    
    return cached_products
```

### 2. Compression
Compress cached data to save memory.

```python
import zlib
import pickle

def set_compressed(key, value, ttl):
    serialized = pickle.dumps(value)
    compressed = zlib.compress(serialized)
    cache.set(key, compressed, ttl)

def get_compressed(key):
    compressed = cache.get(key)
    if compressed:
        decompressed = zlib.decompress(compressed)
        return pickle.loads(decompressed)
    return None
```

### 3. Sharding
Distribute cache across multiple instances.

```python
class ShardedCache:
    def __init__(self, shards):
        self.shards = shards
    
    def get_shard(self, key):
        index = hash(key) % len(self.shards)
        return self.shards[index]
    
    def get(self, key):
        shard = self.get_shard(key)
        return shard.get(key)
    
    def set(self, key, value, ttl):
        shard = self.get_shard(key)
        shard.set(key, value, ttl)
```

## Cache Monitoring

### Key Metrics
- **Cache Hit Ratio**: Percentage of requests served from cache
- **Cache Miss Ratio**: Percentage of requests that miss cache
- **Latency**: Cache response time
- **Memory Usage**: Cache memory consumption
- **Eviction Rate**: Rate of cache evictions

### Monitoring Implementation
```python
class MonitoredCache:
    def __init__(self, cache):
        self.cache = cache
        self.hits = 0
        self.misses = 0
    
    def get(self, key):
        value = self.cache.get(key)
        if value:
            self.hits += 1
        else:
            self.misses += 1
        return value
    
    def get_hit_ratio(self):
        total = self.hits + self.misses
        return self.hits / total if total > 0 else 0
```

## Common Pitfalls

### 1. Cache Stampede
Many requests miss cache simultaneously, overwhelming backend.

**Solution:**
```python
import threading

def get_with_lock(key):
    lock = locks.get(key)
    if lock is None:
        lock = threading.Lock()
        locks[key] = lock
    
    with lock:
        value = cache.get(key)
        if value is None:
            value = fetch_from_source(key)
            cache.set(key, value)
        return value
```

### 2. Thundering Herd
Cache expiration causes simultaneous backend requests.

**Solution:**
```python
def get_with_lease(key):
    value = cache.get(key)
    if value is None:
        # Try to acquire lease
        lease = cache.set(f'{key}:lease', '1', ttl=10, nx=True)
        if lease:
            # We got the lease, fetch from source
            value = fetch_from_source(key)
            cache.set(key, value, ttl=3600)
            cache.delete(f'{key}:lease')
        else:
            # Wait for lease holder
            time.sleep(0.1)
            return get_with_lease(key)
    return value
```

### 3. Stale Data
Serving outdated data from cache.

**Solution:**
- Use appropriate TTL
- Implement cache invalidation
- Use versioning

### 4. Cache Penetration
Repeated requests for non-existent data.

**Solution:**
```python
def get_with_null_cache(key):
    value = cache.get(key)
    if value is None:
        # Check if it's a null cache value
        if cache.get(f'{key}:null'):
            return None
        
        # Fetch from source
        value = fetch_from_source(key)
        if value is None:
            # Cache null result
            cache.set(f'{key}:null', '1', ttl=300)
        else:
            cache.set(key, value, ttl=3600)
    
    return value
```

## Best Practices

### 1. Layered Caching
Use multiple cache layers for optimal performance.

```
Browser Cache → CDN Cache → Application Cache → Database
```

### 2. Cache What's Expensive
Cache data that is expensive to compute or fetch.

### 3. Monitor Cache Performance
Track hit ratios, latency, and memory usage.

### 4. Plan for Cache Failure
Design your application to work even if cache fails.

### 5. Use Appropriate TTL
Set TTL based on data change frequency and staleness tolerance.

### 6. Implement Cache Warming
Pre-populate cache with frequently accessed data.

### 7. Handle Cache Inconsistency
Design for eventual consistency where needed.

### 8. Test Cache Behavior
Test cache hit/miss scenarios and invalidation.

## Conclusion

Caching is a powerful technique for improving system performance and scalability. By understanding different caching strategies, patterns, and best practices, you can design systems that are fast, efficient, and capable of handling high traffic loads. The key is to choose the right caching approach based on your specific requirements and workload characteristics.