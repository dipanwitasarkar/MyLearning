# How Discord Moved Trillions of Messages to ScyllaDB

Scaling Reads, Dealing with Contention

## The TLDR

At the start of 2022, Discord was storing trillions of messages across 177 Cassandra nodes. Cassandra was a strong fit for Discord's write-heavy workload. But the problem was reads, as a popular channel could send thousands of requests to the same partition, overwhelming the nodes and slowing unrelated queries on those same machines. On top of this, compaction backlogs and JVM garbage-collection pauses made the cluster harder to operate.

To address the problems inside the database, Discord replaced Cassandra with ScyllaDB, a compatible reimplementation in C++ that removed the JVM garbage-collection pauses and made the storage layer faster and easier to operate.

But a faster database could not stop thousands of people from requesting the same messages at once. For that, Discord put Rust data services in front of the database. Requests for the same channel route to the same service instance, where overlapping reads for the same message are merged into one database query.

The new cluster runs 72 nodes instead of 177, and p99 reads fell from a range of 40 to 125 milliseconds to a steady 15.

This cleanly separates two problems that are easy to conflate. ScyllaDB made the storage layer faster and easier to operate, while request coalescing reduced how many reads reached it in the first place.

## The Problem

### How Discord stored messages

Discord had stored messages in Cassandra since 2017. By early 2022, the cluster had grown to 177 nodes holding trillions of messages. All Discord messages were stored in a single table that looked like this:

```sql
CREATE TABLE messages (
   channel_id bigint,
   bucket int,
   message_id bigint,
   author_id bigint,
   content text,
   PRIMARY KEY ((channel_id, bucket), message_id)
) WITH CLUSTERING ORDER BY (message_id DESC);
```

In Cassandra, (channel_id, bucket) is the partition key, which determines where the messages are stored. A channel is one chat room inside a Discord server, while the bucket represents a window of time. Together, they place one channel's messages from one period into the same partition, replicated across three nodes.

The message_id then orders messages within that partition, with the newest messages first. The time bucket is what keeps any one partition from growing without bound. Without it, every message ever sent in a popular channel would accumulate in the same partition.

This schema did what it had to in order to distribute the write load across the cluster, but it couldn't make the read traffic evenly distributed. Most live reads for a channel hit its newest bucket, and Discord's channels vary enormously in popularity. A busy public channel could therefore concentrate thousands of reads on one partition and the three nodes that stored it.

The partition was hot because the data itself was popular, not because Discord had chosen too few buckets. Further sharding could spread the load, but only by making the primary query, reading a channel in order, more complicated.

### How one popular channel slowed unrelated requests

Cassandra is designed to make writes cheap. A new message is appended to a commit log and written to an in-memory structure called a memtable. When that memtable fills, Cassandra flushes it to disk as an immutable file called an SSTable.

Because Cassandra never updates those files in place, the latest version of a row may be spread across the memtable and several SSTables, forcing reads to check multiple places and merge what they find. Periodically, a background process called compaction combines SSTables so future reads have fewer files to inspect.

Now put thousands of readers on the partition holding a popular channel's newest messages. Every request lands on the same three replicas, and each replica has to perform the same relatively expensive read work.

Even worse still, those three nodes also stored partitions belonging to thousands of quieter channels. And since Discord ran reads and writes at quorum, meaning each query waited for two of a partition's three replicas to respond, when a node was overwhelmed by traffic for one popular channel, unrelated queries that also needed that node slowed down behind it.

### Compaction debt and garbage-collection pauses

As if things weren't bad enough, the Cassandra cluster was also becoming harder to operate over time.

As mentioned, Cassandra depends on that compaction process to keep the number of SSTables under control, but Discord's cluster regularly fell behind. As more uncompacted files accumulated, reads became slower. Those slower reads consumed more of each node's resources, leaving even less capacity for compaction and allowing the backlog to grow further.

Simply running compaction more aggressively doesn't solve the problem because compaction competes with live traffic for the same CPU and disk, so increasing it just ends up making latency worse.

To catch up, engineers performed what Discord called the "gossip dance." They removed one node from service so it could compact without handling live traffic, brought it back, waited for it to replay the writes it had missed, and then repeated the process on the next node.

Cassandra also runs on the JVM, whose garbage collector periodically pauses application work while reclaiming memory. Those pauses appeared as latency spikes, and the worst lasted long enough that operators had to reboot nodes and nurse them back into the cluster. Together, the compaction backlog and garbage-collection pauses made the cluster fragile. Engineers were spending too much time manually keeping it healthy, even before accounting for the hot partitions caused by Discord's traffic.

## The Solution

### Replacing Cassandra with ScyllaDB

To fix the problems inside the database, Discord replaced Cassandra with ScyllaDB. ScyllaDB uses the same query language and data model as Cassandra, but it is implemented in C++ rather than running on the JVM. That meant Discord could preserve its existing schema while eliminating the garbage collector behind so many of its latency spikes and stability incidents.

ScyllaDB also uses a shard-per-core architecture, where each CPU core handles its own slice of the node's data and requests. This gave Discord better performance from each machine and made repairs faster, reducing the operational work required to keep the cluster healthy.

But changing databases doesn't eliminate hot partitions. ScyllaDB could make each read cheaper, but it can't stop thousands of people from requesting messages stored in the same partition. To solve that problem, Discord had to reduce how many reads reached the database in the first place.

### Request coalescing

Suppose someone posts an @everyone announcement in a huge server and thousands of members open the channel at nearly the same time. Without coalescing, every request independently asks the database for the same message, sending thousands of reads to the same partition.

With coalescing, the first request starts the database query. Any identical request that arrives while that query is still running joins the existing task instead of starting another one. When the database returns the message, the data service sends the result to all waiting callers.

Discord implemented this in Rust data services that sit in front of the database. Requests for the same channel route to the same service instance, where overlapping reads for the same message are merged into one database query.

This dramatically reduced the read load on the database, especially for popular channels where many users might be reading the same messages simultaneously.

## Results

The new cluster runs 72 nodes instead of 177, and p99 reads fell from a range of 40 to 125 milliseconds to a steady 15. The operational burden also decreased significantly, with no more compaction backlogs or garbage-collection pauses to manage.

## Key Takeaways

1. **Separate storage from access patterns**: ScyllaDB made the storage layer faster, but request coalescing was needed to reduce how many reads reached it.

2. **Hot partitions are a real problem**: Even with good partitioning, popular data can overwhelm specific nodes, affecting unrelated queries.

3. **Consider the whole stack**: Sometimes the solution isn't just a faster database, but also reducing the load that reaches it.

4. **Request coalescing is powerful**: Merging identical requests can dramatically reduce database load, especially for read-heavy workloads with popular data.