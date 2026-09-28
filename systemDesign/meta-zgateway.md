# 5 Techniques Meta Uses to Scale a Database to Millions of Clients

Scaling Reads, Scaling Writes

## The TLDR

Meta's ZippyDB gives hundreds of internal teams a shared database for fast key-value lookups. With more than a million client machines using it, managing connections and repeated requests becomes a massive scaling problem.

Meta added a reverse proxy they named ZGateway to coordinate that traffic before it reaches the database. By bringing requests through one shared service, the ZippyDB team can reuse connections and avoid repeating work for different callers. The team can also turn away excess traffic before it overwhelms storage.

This is a playbook in production for scaling reads and writes. Five familiar techniques address different bottlenecks.

## The Problem

When many applications share a database, each opens its own connections and sends its own requests. The database has to maintain all those connections and answer each request, even when thousands of applications are asking for the same data.

And connections aren't free! Each one needs memory to track its state and buffer data, even while idle. Opening a connection also takes CPU time for authentication and setting up encryption. Meta describes database hosts accepting tens of thousands of incoming connections. If many applications restart together, the database has to re-establish their connections while still handling queries. And all that extra work can overwhelm it.

Then there's the work those applications actually ask it to do. Every small request has overhead to receive, decode, and process it. If a thousand applications read the same profile, the database may perform the same lookup a thousand times. If they ask again a second later, it does that work again, even if the profile hasn't changed.

All of this consumes capacity that other applications need. Worse yet, a traffic spike from one team can fill the queues and slow down everyone else, even if their traffic hasn't changed.

## The Solution

Meta added a reverse proxy they named ZGateway between their internal applications and ZippyDB. Applications send their requests to the gateway, and the gateway handles the database connections on their behalf.

Because requests from different applications now pass through the same service, the gateway can combine work those applications would otherwise do separately. It also gives the ZippyDB team a place to control how much work gets through when traffic exceeds capacity.

There are five techniques they use to pull this off, and they're relevant to almost any system that needs to scale reads and writes.

### 1. Connection pooling

Connection pooling means keeping a set of database connections open and reusing them across requests. When an application needs to query the database, it uses a connection that's already established. Once the request finishes, that connection stays available for more work. This way we avoid paying the cost of opening, authenticating, and closing a connection for every query.

An application will often manage its own pool. That works well for reusing connections within that application, but each new application instance brings another pool. The database still has to maintain connections from all of them, even when many are idle.

In Meta's case, that's more than a million internal client hosts, each managing its own connections to the database hosts it needs. Meta describes typical clients holding tens of thousands of outgoing connections, with individual ZippyDB hosts accepting tens of thousands of incoming ones. Pooling within each client still leaves the database maintaining connections from all those separate clients.

By adding a gateway between the clients and the database shards, Meta can share those database connections across clients. Each client connects to a small set of ZGateway hosts, and the gateways forward its requests over their shared pools of connections to ZippyDB. A new client can use those existing backend connections instead of opening its own connections to every database host it needs.

### 2. Batching

Sharing connections reduces the cost of keeping applications connected, but the database still has to receive and process their individual requests. Batching lets several operations share that overhead.

Suppose three application instances send requests to fetch profiles 73, 81, and 92 for the same use case and database shard. Without batching, ZGateway could forward three separate requests to ZippyDB. Instead, it sends all three keys in a single backend request, then routes the individual results back to the correct callers.

The database still performs three lookups, but they share the overhead of serialization, network transmission, and handling a backend request. Those fixed costs would otherwise be paid once per operation.

The same idea works for writes. Several independent updates can travel together while still being applied as separate changes once they reach the database.

Batching also reduces the backend requests counted against a use case's rate-limit budget, letting it perform more operations before being throttled.

The tradeoff is that those requests won't necessarily arrive at the same time. When the request for profile 73 arrives, the gateway could send it immediately. Holding it briefly gives the requests for 81 and 92 a chance to join, so all three can share the overhead. But the caller asking for 73 now waits longer for its answer. The longer you wait, the more efficient the batches, but it also adds more delay before the database even starts the work.

ZGateway bounds this by flushing a batch when either a timer expires or the batch reaches a size or request-count limit. It also limits how many batches can be in flight at once to protect gateway memory when the database slows.

### 3. Request coalescing

Now suppose the three requests arriving at a gateway all want user:73.

For a simplified example, caller A arrives first and the gateway begins a lookup against ZippyDB. Before that request finishes, callers B and C arrive asking for the same key. Rather than issue two additional database reads, the gateway can attach those callers to the lookup that is already in flight. When the result comes back, all three callers receive it.

This is request coalescing. The difference from batching is how much work reaches the database. A batch containing three different keys still needs three lookups, even though they travel together. With coalescing, the callers are asking for the same thing, so one lookup can answer all of them. This becomes especially useful when a profile suddenly gets popular. Thousands of people might request it at once, and many of those reads can share a lookup instead of each creating more work for the database.

For that to happen, the requests need to reach the same gateway while the lookup is still running. They also need to refer to the same record with compatible read requirements.

### 4. Caching

ZGateway also caches responses. When a request completes, the gateway stores the result. Subsequent requests for the same key can be served from cache without hitting the database at all.

This is particularly effective for read-heavy workloads where the same data is requested repeatedly. The cache reduces database load and improves latency for cached requests.

### 5. Rate limiting

With all these techniques in place, ZGateway still needs to protect the database from being overwhelmed. Rate limiting at the gateway level allows the ZippyDB team to control how much traffic reaches the database.

When traffic exceeds capacity, the gateway can reject excess requests or queue them, preventing the database from being overwhelmed. This protects the database and ensures that critical requests still get through.

## Key Takeaways

1. **Use a gateway for shared resources**: When many clients access the same resource, a gateway can pool connections, batch requests, and implement shared policies.

2. **Connection pooling at the gateway level**: Pooling connections at the gateway rather than per-client dramatically reduces connection overhead on the database.

3. **Batching reduces per-request overhead**: Grouping multiple operations together shares the fixed costs of network transmission and request handling.

4. **Request coalescing for duplicate requests**: When multiple clients request the same data, serve them from a single database lookup.

5. **Cache to reduce database load**: Caching frequently accessed data at the gateway level reduces database load and improves latency.

6. **Rate limiting protects the database**: Implement rate limiting at the gateway to prevent overwhelming the database during traffic spikes.
