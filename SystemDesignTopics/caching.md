# Caching Strategies

## Overview

Caching is the process of storing frequently accessed data in a fast storage layer to reduce data retrieval time and system load. Caching is one of the most effective techniques for improving system performance and scalability.

In system design interviews, caching comes up almost every time you need to handle high read traffic. Your database becomes the bottleneck, latency starts creeping up, and the interviewer is waiting for you to say the word: cache.

Reading a user profile from Postgres may take 50 milliseconds, but reading from an in-memory cache like Redis takes just 1 millisecond. That's a 50x improvement in latency. Databases store data on disk, and every query pays the cost of disk access. Memory sits much closer to the CPU and avoids that entirely.

Caches are essential for scalable systems. They reduce load on the database and cut latency dramatically. But they also create new challenges around invalidation and failure handling.

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

## Where to Cache

When most engineers hear caching, they immediately think of Redis or Memcached sitting between the application and the database. It is the most common type of cache and the one interviewers care about the most.

But caching shows up in multiple layers of a system. Browsers cache. CDNs cache. Applications cache. Even databases have built-in caching layers.

### External Caching

An external cache is a standalone cache service that your application talks to over the network. This is what most people think of when they hear caching. You store frequently accessed data in something like Redis or Memcached so you do not have to hit the database every time.

External caches scale well because every application server can share the same cache. They also support eviction policies like LRU and expiration via TTL so your memory footprint stays controlled.

In system design interviews, external caching with Redis is the default answer when discussing caching strategies. Interviewers expect you to mention it for any high-traffic system. Start here, then layer on other caching types such as CDN or client-side caching only if the problem calls for them.

### CDN (Content Delivery Network)

A CDN is a geographically distributed network of servers that caches content close to users. Instead of every request traveling to your origin server, a CDN stores copies of your content at edge servers around the world.

Modern CDNs like Cloudflare, Fastly, and Akamai can cache much more than static files. They can also cache public API responses, HTML pages, and even run edge logic to personalize content or enforce security rules before requests reach your servers. But the most common and most impactful use of a CDN is still media delivery.

**How it works:**
1. A user requests an image from your app.
2. The request goes to the nearest CDN edge server.
3. If the image is cached there, it is returned immediately.
4. If not, the CDN fetches it from your origin server, stores it, and returns it.
5. Future users in that region get the image instantly from the CDN.

Without a CDN, every image request travels to your origin. If your server is in Virginia and the user is in India, that adds 250–300 ms of latency per request. With a CDN, the same image is served from a nearby edge server in 20–40 ms. That is a massive performance difference.

Even though modern CDNs can cache API responses and dynamic content, in system design interviews the safest time to introduce a CDN is when your system serves static media at scale. Start with that reason first, then expand only if the problem calls for more.

**Pros:**
- Global distribution
- Reduced latency
- Handles high traffic
- DDoS protection

**Cons:**
- Cost
- Cache invalidation complexity
- Limited dynamic content support

### Client-Side Caching

Client-side caching stores data close to the requester to avoid unnecessary network calls. This usually means the user's device, like a browser (HTTP cache, localStorage) or mobile app using local memory or on-device storage.

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

It can also mean caching within a client library. For example, Redis clients cache cluster metadata like which nodes are in the cluster and which slots are assigned to them. That way, the client can route requests directly to the right node without querying the cluster on every operation.

For user-facing caching, you have limited control from the backend. Data can go stale and invalidation is harder.

### In-Process Caching

Most candidates, and engineers, overlook the fact that servers run on machines with a lot of memory. Each hardware generation ships more of it. You can use that memory to cache data directly inside the application process instead of always calling out to Redis or the database.

The idea is simple: if your service keeps requesting the same small pieces of data again and again, store them in a local cache inside the process. Reads from local memory are even faster than reads from Redis because they avoid any network call.

This light-weight form of caching makes sense for small pieces of data that are requested frequently like:
- Configuration values
- Feature flags
- Small reference datasets
- Hot keys
- Rate limiting counters
- Precomputed values

In-process caching is blazing fast, but it comes with obvious limitations. Each instance of your application has its own cache, so cached data is not shared across servers. If one instance updates or invalidates a cached value, the others will not know.

Use in-process caching for small, frequently accessed values that rarely change. It is great for speed but not a replacement for Redis. In system design interviews, mention this only as an **optimization layer** after you have already introduced an external cache.

### Database Caching

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

Not all caching works the same way. How you read from and write to the cache changes performance, consistency, and complexity. These are the four core cache patterns you should know for system design interviews.

### Cache-Aside (Lazy Loading)

This is the most common caching pattern and the one you should default to in interviews.

**How it works:**
1. Application checks the cache.
2. If the data is there, return it.
3. If not, fetch from the database, store it in the cache, and return it.

Cache-aside only caches data when needed, which keeps the cache lean. The downside is that a cache miss causes extra latency.

If you only remember one caching pattern for interviews, make it cache-aside.

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

### Write-Through Caching

With write-through caching, the application writes only to the cache. The cache then synchronously writes to the database before returning to the application. The write operation does not complete until both the cache and database are updated.

In practice, this requires a cache implementation that supports write-through, like a caching library with a data store plugin. When you write to the cache, the library handles calling your database write logic before acknowledging the write. Redis itself does not natively support write-through, so you need application code or a framework to implement this pattern.

**How it works:**
1. Application writes data
2. Write to cache
3. Write to database
4. Return success

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

Write-through still suffers from the dual-write problem. If the cache update succeeds but the database write fails, you have inconsistent state.

### Write-Back (Write-Behind) Caching

Write-back caching is the opposite of write-through. The application writes to the cache, and the cache asynchronously writes to the database later. The write operation completes immediately after updating the cache.

This gives you the fastest possible writes since the application doesn't wait for the database. But it comes with significant risks: if the cache crashes before writing to the database, data is lost permanently.

Write-back is rarely used in interviews unless you're specifically discussing systems where write performance is critical and some data loss is acceptable (like analytics or logging).

**How it works:**
1. Application writes data
2. Write to cache
3. Queue database write
4. Return success immediately
5. Background process writes to database

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

### Refresh-Ahead Caching

Refresh-ahead caching proactively refreshes cache entries before they expire. Instead of waiting for a cache miss, the system predicts which data will be needed and refreshes it in the background.

This is more complex to implement but can eliminate cache misses for predictable access patterns. In interviews, mention this as an advanced optimization if the interviewer pushes on cache hit rates.

**How it works:**
1. Application requests data
2. Check cache
3. If cache hit and expiring soon: refresh in background
4. Return cached data

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

The hardest part of caching is keeping the cache in sync with the source of truth. There are three main strategies:

### Time-to-Live (TTL)

Set an expiration time on cache entries. When the TTL expires, the entry is evicted. Simple and effective, but data can be stale until the next request.

This is the default strategy in most interviews. Just say "I'll set a reasonable TTL based on how often the data changes."

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

### Cache Invalidation on Write

When data is updated in the database, invalidate or update the corresponding cache entry. This keeps data fresh but adds complexity to write operations.

**Implementation:**
```python
def update_user(user_id, data):
    # Update database
    db.execute("UPDATE users SET data = ? WHERE id = ?", data, user_id)
    
    # Invalidate cache
    cache.delete(f'user:{user_id}')
    
    return True
```

**Common approaches:**
- **Invalidation**: Delete the cache entry on write. Next read fetches fresh data.
- **Update**: Update the cache entry on write. Keeps cache fresh but risks race conditions.

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

### Write-Through Invalidation

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

### Event-Based Invalidation

Use a message queue or event system to propagate cache invalidations across services. When data changes, publish an event. All services listening for that event invalidate their cache.

This is more complex but necessary for distributed systems with multiple cache layers.

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

## Cache Eviction Policies

### LRU (Least Recently Used)

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

### LFU (Least Frequently Used)

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

### FIFO (First In First Out)

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

When a cache entry expires and many requests simultaneously try to refresh it, you can overwhelm the database.

**Solutions:**
- **Locking**: Only one request refreshes the cache, others wait
- **Probabilistic early expiration**: Refresh before expiration with some probability
- **Stale-while-revalidate**: Return stale data while refreshing in background

**Implementation:**
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

### 3. Hot Keys

Some keys get accessed way more than others, creating hotspots.

**Solutions:**
- **Shard the hot key**: Split the data across multiple cache keys
- **Local caching**: Cache hot keys in-process to reduce network calls
- **Rate limiting**: Limit requests to hot keys

### 4. Memory Pressure

If your cache grows too large, it can cause eviction of important data or even crashes.

**Solutions:**
- **Set appropriate TTLs**: Don't cache data forever
- **Monitor cache hit rates**: Adjust caching strategy based on metrics
- **Use eviction policies**: LRU, LFU, or custom policies

### 5. Cache Penetration

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

### 6. Inconsistency

Cache and database can get out of sync.

**Solutions:**
- **Accept eventual consistency**: For many use cases, slight staleness is fine
- **Use write-through**: Ensures cache and database are always in sync
- **Implement proper invalidation**: Invalidate cache on data changes

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

## When to Use Caching

In interviews, introduce caching when:
- You have high read traffic
- Database is becoming a bottleneck
- Latency requirements are strict
- Data is read much more often than written
- Some data staleness is acceptable

Don't use caching when:
- Data changes frequently
- Strong consistency is required
- Cache hit rate would be low
- The complexity outweighs the benefits

## Summary

For system design interviews:
1. Default to external caching with Redis
2. Use cache-aside (lazy loading) as your primary pattern
3. Set appropriate TTLs based on data change frequency
4. Consider CDN for static media
5. Mention client-side and in-process caching as optimizations
6. Discuss cache invalidation strategies
7. Be aware of common pitfalls (stampede, hot keys, inconsistency)
8. Only introduce caching when it makes sense for the problem

Caching is a powerful technique for improving system performance and scalability. By understanding different caching strategies, patterns, and best practices, you can design systems that are fast, efficient, and capable of handling high traffic loads. The key is to choose the right caching approach based on your specific requirements and workload characteristics.