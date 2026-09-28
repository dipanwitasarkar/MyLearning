# Caching

Learn about caching and when to use it in system design interviews.

In system design interviews, caching comes up almost every time you need to handle high read traffic. Your database becomes the bottleneck, latency starts creeping up, and the interviewer is waiting for you to say the word: cache.

Reading a user profile from Postgres may take 50 milliseconds, but reading from an in-memory cache like Redis takes just 1 millisecond. That's a 50x improvement in latency. Databases store data on disk, and every query pays the cost of disk access. Memory sits much closer to the CPU and avoids that entirely.

Caches are essential for scalable systems. They reduce load on the database and cut latency dramatically. But they also create new challenges around invalidation and failure handling.

This breakdown covers the basics of caching, when and where to use it, common pitfalls, and how to talk about caching clearly in interviews.

## Where to Cache

When most engineers hear caching, they immediately think of Redis or Memcached sitting between the application and the database. It is the most common type of cache and the one interviewers care about the most.

But caching shows up in multiple layers of a system. Browsers cache. CDNs cache. Applications cache. Even databases have built-in caching layers.

Let's look at the main places you can cache data, why each one exists, and when it makes sense to use it.

### External Caching

An external cache is a standalone cache service that your application talks to over the network. This is what most people think of when they hear caching. You store frequently accessed data in something like Redis or Memcached so you do not have to hit the database every time.

External caches scale well because every application server can share the same cache. They also support eviction policies like LRU and expiration via TTL so your memory footprint stays controlled.

In system design interviews, external caching with Redis is the default answer when discussing caching strategies. Interviewers expect you to mention it for any high-traffic system. Start here, then layer on other caching types such as CDN or client-side caching only if the problem calls for them.

### CDN (Content Delivery Network)

A CDN is a geographically distributed network of servers that caches content close to users. Instead of every request traveling to your origin server, a CDN stores copies of your content at edge servers around the world.

Modern CDNs like Cloudflare, Fastly, and Akamai can cache much more than static files. They can also cache public API responses, HTML pages, and even run edge logic to personalize content or enforce security rules before requests reach your servers. But the most common and most impactful use of a CDN is still media delivery.

How it works:
1. A user requests an image from your app.
2. The request goes to the nearest CDN edge server.
3. If the image is cached there, it is returned immediately.
4. If not, the CDN fetches it from your origin server, stores it, and returns it.
5. Future users in that region get the image instantly from the CDN.

Without a CDN, every image request travels to your origin. If your server is in Virginia and the user is in India, that adds 250–300 ms of latency per request. With a CDN, the same image is served from a nearby edge server in 20–40 ms. That is a massive performance difference.

Even though modern CDNs can cache API responses and dynamic content, in system design interviews the safest time to introduce a CDN is when your system serves static media at scale. Start with that reason first, then expand only if the problem calls for more.

### Client-Side Caching

Client-side caching stores data close to the requester to avoid unnecessary network calls. This usually means the user's device, like a browser (HTTP cache, localStorage) or mobile app using local memory or on-device storage.

But it can also mean caching within a client library. For example, Redis clients cache cluster metadata like which nodes are in the cluster and which slots are assigned to them. That way, the client can route requests directly to the right node without querying the cluster on every operation.

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

## Cache Architectures

Not all caching works the same way. How you read from and write to the cache changes performance, consistency, and complexity. These are the four core cache patterns you should know for system design interviews.

### Cache-Aside (Lazy Loading)

This is the most common caching pattern and the one you should default to in interviews.

How it works:
1. Application checks the cache.
2. If the data is there, return it.
3. If not, fetch from the database, store it in the cache, and return it.

Cache-aside only caches data when needed, which keeps the cache lean. The downside is that a cache miss causes extra latency.

If you only remember one caching pattern for interviews, make it cache-aside.

### Write-Through Caching

With write-through caching, the application writes only to the cache. The cache then synchronously writes to the database before returning to the application. The write operation does not complete until both the cache and database are updated.

In practice, this requires a cache implementation that supports write-through, like a caching library with a data store plugin. When you write to the cache, the library handles calling your database write logic before acknowledging the write. Redis itself does not natively support write-through, so you need application code or a framework to implement this pattern.

The tradeoff is slower writes because the application must wait for both the cache update and the database write to complete. Write-through can also pollute the cache with data that may never be read again.

Write-through still suffers from the dual-write problem. If the cache update succeeds but the database write fails, you have inconsistent state.

### Write-Back (Write-Behind) Caching

Write-back caching is the opposite of write-through. The application writes to the cache, and the cache asynchronously writes to the database later. The write operation completes immediately after updating the cache.

This gives you the fastest possible writes since the application doesn't wait for the database. But it comes with significant risks: if the cache crashes before writing to the database, data is lost permanently.

Write-back is rarely used in interviews unless you're specifically discussing systems where write performance is critical and some data loss is acceptable (like analytics or logging).

### Refresh-Ahead Caching

Refresh-ahead caching proactively refreshes cache entries before they expire. Instead of waiting for a cache miss, the system predicts which data will be needed and refreshes it in the background.

This is more complex to implement but can eliminate cache misses for predictable access patterns. In interviews, mention this as an advanced optimization if the interviewer pushes on cache hit rates.

## Cache Invalidation

The hardest part of caching is keeping the cache in sync with the source of truth. There are three main strategies:

### Time-to-Live (TTL)

Set an expiration time on cache entries. When the TTL expires, the entry is evicted. Simple and effective, but data can be stale until the next request.

This is the default strategy in most interviews. Just say "I'll set a reasonable TTL based on how often the data changes."

### Cache Invalidation on Write

When data is updated in the database, invalidate or update the corresponding cache entry. This keeps data fresh but adds complexity to write operations.

Common approaches:
- **Invalidation**: Delete the cache entry on write. Next read fetches fresh data.
- **Update**: Update the cache entry on write. Keeps cache fresh but risks race conditions.

### Event-Based Invalidation

Use a message queue or event system to propagate cache invalidations across services. When data changes, publish an event. All services listening for that event invalidate their cache.

This is more complex but necessary for distributed systems with multiple cache layers.

## Common Pitfalls

### Cache Stampede

When a cache entry expires and many requests simultaneously try to refresh it, you can overwhelm the database. Solutions:
- **Locking**: Only one request refreshes the cache, others wait
- **Probabilistic early expiration**: Refresh before expiration with some probability
- **Stale-while-revalidate**: Return stale data while refreshing in background

### Hot Keys

Some keys get accessed way more than others, creating hotspots. Solutions:
- **Shard the hot key**: Split the data across multiple cache keys
- **Local caching**: Cache hot keys in-process to reduce network calls
- **Rate limiting**: Limit requests to hot keys

### Memory Pressure

If your cache grows too large, it can cause eviction of important data or even crashes. Solutions:
- **Set appropriate TTLs**: Don't cache data forever
- **Monitor cache hit rates**: Adjust caching strategy based on metrics
- **Use eviction policies**: LRU, LFU, or custom policies

### Inconsistency

Cache and database can get out of sync. Solutions:
- **Accept eventual consistency**: For many use cases, slight staleness is fine
- **Use write-through**: Ensures cache and database are always in sync
- **Implement proper invalidation**: Invalidate cache on data changes

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