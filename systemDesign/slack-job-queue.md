# How Slack Put Kafka in Front of Its Redis Job Queue

Managing Long Running Tasks, Scaling Writes

## The TLDR

On its busiest days, Slack's async job queue handles more than 1.4 billion jobs, peaking at 33,000 per second. It powers nearly everything too slow for a web request, including posting messages, sending push notifications, generating URL unfurls, triggering calendar reminders, and running billing calculations.

That queue was built entirely in Redis, which held both the backlog and the dispatch bookkeeping. That is until a database slowdown made jobs pile up until Redis hit its memory limit. At that point, the queue could neither accept new jobs nor, since dequeuing also needed a little free memory, hand out the old ones.

To prevent this from happening again, they added Kafka in front of Redis as a durable buffer plus two new Go services to move jobs in and out, keeping the application's queue logic untouched. As a result, a backlog now piles up on disk while operators throttle the flow into Redis.

This is a canonical example of putting a durable log in front of a fast dispatch layer in order to handle bursts.

## The Problem

### The whole queue lives in RAM

Slack's job queue runs on a fleet of Redis clusters, with pools of worker machines polling those clusters for new work. When something in Slack needs to happen later, the web app builds a job identifier from the job type and its arguments, then hashes it with the logical queue name to pick which Redis host the job lands on. That host runs limited deduplication, discarding the request if an identical identifier is already waiting. Workers pull jobs off the pending queue, moving each onto an in-flight list before spawning an async task to run it. Finished jobs come off the list, and failures retry until they land in a permanently-failed list that humans repair by hand.

Jobs run from a few milliseconds to several minutes, and the queue had scaled from Slack's earliest days without the core architecture changing. The same finite pool of RAM held both the buffer that absorbs bursts of new work and the working space where dispatch, dedup, and retries get tracked.

### Why the queue stayed down after the database recovered

Resource contention in the database layer slowed job execution, so jobs accumulated in Redis until it hit its configured maximum memory. Enqueues began failing, and with them every Slack feature that depended on the queue. The dequeue path then prevented the queue from recovering. Pulling a job off the queue moves it onto a processing list, and that move needs free memory a full Redis doesn't have, so even after the database recovered the queue stayed stuck, and it took extensive manual intervention to revive.

Memory exhaustion was only one of the problems the post-mortem turned up. Every job-queue client connected to every Redis instance, a complete bipartite graph. Workers couldn't be scaled independently either, because each added worker adds polling load on Redis, so adding execution capacity could push an already overloaded Redis further under. The linear-cost dequeue made a second feedback loop, since a longer queue slows every dequeue at the moment draining matters most.

## The Solution

### Kafka in front

The post-mortem set the requirements. The backlog had to live somewhere durable that couldn't run out of memory whether jobs flooded in or drained too slowly, scheduling needed real rate limits and priorities, and execution capacity had to scale without adding polling load on Redis. Slack put Kafka in front of Redis instead of replacing Redis with it.

Slack could have rebuilt the queue directly on Kafka. It's a durable append-only log, split into partitions, where each consumer tracks its own read position as an offset, and the log lives on disk, bounded by a retention window rather than by RAM, so a burst of enqueues just makes the log longer. An offset is a position in a partition, not per-job state, so per-job acknowledgment, retries, dedup, and in-flight tracking have to be built on top.

Slack had all of those semantics already built in Redis, with years of application code depending on their exact behavior, so rebuilding on Kafka alone would have turned a narrowly scoped availability fix into a rewrite.

RabbitMQ, SQS, and beanstalkd all persisted to disk, and all had the same problem, since each brings its own delivery semantics and adopting one meant the same rewrite against a different API. Turning on Redis persistence wouldn't have helped either, because a Redis dataset has to fit in RAM whether or not it's also written to disk, and the outage was a capacity problem, not a durability one.

So Slack made what they called the "minimum viable change". The web app enqueues into Kafka, which durably buffers everything, a relay drains Kafka into Redis at a controlled rate, and Redis keeps providing the short queue and the semantics the application understands. Neither the enqueue interface nor the worker side changed.

### Kafkagate and JQRelay

Slack wrote two new stateless Go services to run the new path. Kafkagate is the enqueue path from the web app into Kafka. The interface is a plain HTTP POST:

```json
POST /enqueue
{
  "topic": "jobs-cluster-07",
  "partition": 12,
  "content": { "type": "url_unfurl", "args": { ... } }
}
```

Kafkagate holds persistent broker connections, follows partition leadership as it moves, and writes synchronously, so the caller always gets a positive ack or an error. It waits only for the partition leader to acknowledge a write, not for replication, taking the lowest latency in exchange for a small chance of losing a job if a broker dies before replicating, a tradeoff Slack judged right for most jobs.

JQRelay is the drain, one instance relaying one Kafka topic into its corresponding Redis cluster. It pulls messages from Kafka and enqueues them into Redis using the existing queue logic. If Redis is full or slow, JQRelay backs off, letting the backlog grow in Kafka instead of overwhelming Redis.

## Results

The new architecture separates concerns: Kafka handles durable storage and buffering, Redis handles fast dispatch and complex queue semantics, and the two Go services manage the flow between them. When the database slows down again, jobs pile up in Kafka (on disk) instead of Redis (in RAM), and operators can throttle the flow into Redis to prevent memory exhaustion.

## Key Takeaways

1. **Separate buffering from dispatch**: Use a durable log (Kafka) for buffering and a fast in-memory system (Redis) for dispatch logic.

2. **Memory limits are real**: In-memory systems can hit hard limits that prevent recovery, even after the root cause is fixed.

3. **Minimum viable changes**: Sometimes the best solution is to add a layer rather than rewrite everything.

4. **Durable storage as a buffer**: Disk-based storage can absorb bursts that would overwhelm memory-based systems.