# Kafka-like Queue System Design

## 1. Clarify Requirements

### Functional Requirements
- Publish-subscribe messaging model
- Persistent message storage with configurable retention
- Partitioning for parallelism and scalability
- Replication for fault tolerance
- Consumer groups for parallel consumption
- Offset tracking per consumer
- At-least-once delivery semantics
- Optional exactly-once semantics
- Message ordering guarantees (within partition)
- Replay capability for consumers

### Non-Functional Requirements
- **Throughput**: Millions of messages per second
- **Latency**: Millisecond-level publish latency
- **Durability**: No data loss with proper configuration
- **Scalability**: Horizontal scaling of producers, brokers, consumers
- **Availability**: Survive broker failures without data loss
- **Storage Efficiency**: High compression ratio for large datasets

### Clarifying Questions
- What is the expected message throughput and message size?
- What are the ordering requirements (total vs per-key)?
- What is the required retention period?
- How many consumer groups will consume the data?
- What are the durability requirements (sync vs async replication)?
- Do we need exactly-once semantics or is at-least-once acceptable?
- What is the expected failure rate and recovery time objectives?

## 2. Estimate Scale

### Traffic Assumptions
- 10M messages per second peak
- Average message size: 1KB
- 100 topics
- 10 partitions per topic average
- 3 replicas per partition
- 7-day retention period
- 10 consumer groups per topic

### Storage Requirements
- Daily message volume: 10M × 1KB × 86400 = 864GB/day
- With replication (factor 3): 2.6TB/day
- 7-day retention: 18.2TB
- With compression (2x): 9.1TB
- Index and metadata overhead: 20% = 11TB total

### Network Requirements
- Inbound traffic: 10M × 1KB = 10GB/s
- Replication traffic: 10GB/s × 2 (for 2 replicas) = 20GB/s
- Consumer traffic: 10GB/s × 10 consumer groups = 100GB/s
- Total network: 130GB/s

### Broker Count Estimation
- Single broker capacity: 100K messages/sec, 2TB storage
- Required throughput: 10M messages/sec
- Brokers for throughput: 10M / 100K = 100 brokers
- Brokers for storage: 11TB / 2TB = 6 brokers
- **Total: 100 brokers (throughput limited)**

### Partition Count
- Total partitions: 100 topics × 10 partitions = 1,000 partitions
- Partitions per broker: 1,000 / 100 = 10 partitions/broker
- Replicas: 1,000 × 3 = 3,000 partition replicas

## 3. APIs

### Producer API
```
POST /topics/:topic_name
Request: {
  "key": "user123",  // optional, used for partitioning
  "value": {"order_id": "42", "amount": 100.50},
  "headers": {
    "content-type": "application/json"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
Response: {
  "topic": "orders",
  "partition": 5,
  "offset": 12345,
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### Consumer API
```
POST /consumers/:group_id/subscribe
Request: {
  "topics": ["orders", "payments"],
  "auto_offset_reset": "latest"  // or "earliest"
}
Response: {
  "consumer_id": "consumer123",
  "assignments": [
    {"topic": "orders", "partition": 0},
    {"topic": "orders", "partition": 1}
  ]
}
```

### Fetch Messages
```
GET /consumers/:group_id/messages
Response: {
  "messages": [
    {
      "topic": "orders",
      "partition": 0,
      "offset": 12345,
      "key": "user123",
      "value": {"order_id": "42", "amount": 100.50},
      "timestamp": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### Commit Offset
```
POST /consumers/:group_id/offsets
Request: {
  "offsets": [
    {"topic": "orders", "partition": 0, "offset": 12346}
  ]
}
Response: 200 OK
```

### Create Topic
```
POST /topics
Request: {
  "topic_name": "orders",
  "partitions": 10,
  "replication_factor": 3,
  "retention_ms": 604800000,  // 7 days
  "cleanup_policy": "delete"  // or "compact"
}
Response: {
  "topic_name": "orders",
  "partitions": 10,
  "replication_factor": 3
}
```

### Get Topic Metadata
```
GET /topics/:topic_name
Response: {
  "name": "orders",
  "partitions": [
    {
      "partition_id": 0,
      "leader": "broker1",
      "replicas": ["broker1", "broker2", "broker3"],
      "isr": ["broker1", "broker2", "broker3"]
    }
  ],
  "retention_ms": 604800000,
  "cleanup_policy": "delete"
}
```

## 4. Data Model

### Topic Metadata
```sql
CREATE TABLE topics (
    topic_name VARCHAR(255) PRIMARY KEY,
    partitions INT NOT NULL,
    replication_factor INT NOT NULL,
    retention_ms BIGINT,
    cleanup_policy VARCHAR(20), -- delete, compact
    created_at TIMESTAMP DEFAULT NOW()
);
```

### Partition Metadata
```sql
CREATE TABLE partitions (
    topic_name VARCHAR(255),
    partition_id INT,
    leader_broker VARCHAR(50) NOT NULL,
    replica_brokers VARCHAR(50)[],
    isr_brokers VARCHAR(50)[],  -- in-sync replicas
    high_watermark BIGINT,
    PRIMARY KEY (topic_name, partition_id),
    FOREIGN KEY (topic_name) REFERENCES topics(topic_name)
);
```

### Consumer Group Metadata
```sql
CREATE TABLE consumer_groups (
    group_id VARCHAR(255) PRIMARY KEY,
    state VARCHAR(20), -- stable, preparing_rebalance, etc.
    generation INT NOT NULL,
    protocol VARCHAR(50),
    leader VARCHAR(255),
    assigned_partitions JSONB
);
```

### Consumer Offsets
```sql
CREATE TABLE consumer_offsets (
    group_id VARCHAR(255),
    topic_name VARCHAR(255),
    partition_id INT,
    offset BIGINT NOT NULL,
    metadata TEXT,
    timestamp TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (group_id, topic_name, partition_id)
);
```

### Broker Metadata
```sql
CREATE TABLE brokers (
    broker_id VARCHAR(50) PRIMARY KEY,
    host VARCHAR(255) NOT NULL,
    port INT NOT NULL,
    rack VARCHAR(50),
    status VARCHAR(20) DEFAULT 'active',
    last_heartbeat TIMESTAMP DEFAULT NOW()
);
```

### Log Segment Metadata (stored on disk)
```
Segment: 00000000000000000000.log
Index: 00000000000000000000.index
Timeindex: 00000000000000000000.timeindex
Leader-epoch-checkpoint: leader-epoch-checkpoint
```

## 5. High Level Design

```
┌─────────────────────────────────────────────────────────────┐
│                      Producers                               │
│  (Multiple producer applications)                            │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                   Broker Cluster                             │
│                                                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │  Broker 1   │  │  Broker 2   │  │  Broker 3   │          │
│  │             │  │             │  │             │          │
│  │ Topic A     │  │ Topic A     │  │ Topic A     │          │
│  │ Partitions: │  │ Partitions: │  │ Partitions: │          │
│  │ 0,3,6       │  │ 1,4,7       │  │ 2,5,8       │          │
│  │             │  │             │  │             │          │
│  │ Topic B     │  │ Topic B     │  │ Topic B     │          │
│  │ Partitions: │  │ Partitions: │  │ Partitions: │          │
│  │ 0,2         │  │ 1,3         │  │ 4,5         │          │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘          │
│         │                 │                 │                 │
│         └─────────────────┴─────────────────┘                 │
│                    Replication                                 │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                   Consumer Groups                            │
│                                                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │  Group A    │  │  Group B    │  │  Group C    │          │
│  │             │  │             │  │             │          │
│  │ Consumer 1  │  │ Consumer 1  │  │ Consumer 1  │          │
│  │ Consumer 2  │  │ Consumer 2  │  │ Consumer 2  │          │
│  │ Consumer 3  │  │             │  │             │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
```

### Key Design Decisions
- **Log-based storage**: Append-only, immutable segments
- **Partitioning**: Unit of parallelism and distribution
- **Replication**: Leader-follower model with ISR
- **Consumer groups**: Independent consumption with offset tracking
- **Pull-based consumption**: Consumers control their pace

## 6. Deep Dive Components

### 6.1 Topic and Partition Design

#### Topic
- Named stream of records
- Example: "orders", "clicks", "user-events"
- Configurable retention and cleanup policies

#### Partition
- Ordered, immutable sequence of records
- Each record has unique offset
- Unit of parallelism and distribution

#### Partitioning Strategy
```python
def partition(key, num_partitions):
    if key is None:
        return round_robin(num_partitions)
    return hash(key) % num_partitions
```

#### Benefits of Partitioning
- **Parallelism**: Multiple consumers per group
- **Scalability**: Distribute across brokers
- **Ordering**: Guaranteed within partition

#### Trade-off
- Total ordering vs parallelism
- Choose: Ordering within partition, parallelism across partitions

### 6.2 Replication Design

#### Replication Factor
- Number of copies of each partition
- Typical: 3 for production

#### Leader-Follower Model
- Leader handles all reads and writes
- Followers replicate from leader
- Automatic leader election on failure

#### In-Sync Replicas (ISR)
- Set of replicas fully caught up with leader
- Only ISR members can become new leader
- Configured lag threshold

#### Ack Configuration
- `acks=0`: Fire and forget (no durability)
- `acks=1`: Wait for leader ack (default)
- `acks=all`: Wait for all ISR acks (highest durability)

#### Replication Flow
1. Producer sends message to leader
2. Leader appends to local log
3. Leader replicates to followers
4. Followers acknowledge
5. Leader acknowledges producer (based on acks config)

### 6.3 Producer Design

#### Message Flow
1. Producer serializes message
2. Selects partition (based on key or round-robin)
3. Sends to partition leader
4. Waits for ack (based on configuration)
5. Retries on failure

#### Key Features
- **Batching**: Group messages for efficiency
- **Compression**: Reduce network overhead
- **Idempotence**: Prevent duplicates on retry
- **Transactions**: Atomic writes across partitions

#### Producer Configuration Trade-offs
| Config | Low Latency | High Throughput | High Durability |
|--------|-------------|-----------------|------------------|
| Batch Size | Small | Large | Medium |
| Compression | None | Snappy/LZ4 | GZIP |
| Acks | 0 | 1 | all |
| Retries | 0 | 3 | Unlimited |

### 6.4 Consumer Design

#### Consumer Group
- Set of consumers coordinating consumption
- Each partition consumed by one consumer in group
- Load balancing across consumers

#### Offset Tracking
- Each consumer tracks offset per partition
- Stored in Kafka (__consumer_offsets topic)
- Can commit automatically or manually

#### Delivery Semantics

##### At-least-once (Default)
- Offset committed after processing
- Possible duplicates on failure
- No data loss

##### At-most-once
- Offset committed before processing
- Possible data loss on failure
- No duplicates

##### Exactly-once
- Transactions + idempotent consumers
- Complex implementation
- No duplicates, no data loss

#### Rebalancing
- Group coordinator manages partition assignment
- Triggered on consumer join/leave/failure
- Stop-the-world pause during rebalance

#### Implementation
```python
def consume_messages(consumer_group, topic):
    consumer = create_consumer(consumer_group)
    consumer.subscribe([topic])

    while True:
        messages = consumer.poll(timeout_ms=1000)
        for message in messages:
            try:
                process_message(message)
                consumer.commit(offsets=message.offset)
            except Exception as e:
                log_error(e)
                # Don't commit offset, will retry
```

### 6.5 Storage Design

#### Log Segments
- Partition divided into segments
- Each segment: data file + index file
- Segments rolled over based on size or time

#### Index Files
- **Offset index**: Map offset to file position
- **Time index**: Map timestamp to offset
- Enable efficient lookups

#### Segment Structure
```
00000000000000000000.log  (actual messages)
00000000000000000000.index (offset -> position)
00000000000000000000.timeindex (timestamp -> offset)
```

#### Retention Policies

##### Time-based
- Delete messages older than X days
- Configurable per topic
- Default: 7 days

##### Size-based
- Delete when partition exceeds X GB
- Configurable per topic
- Default: No limit

##### Compaction
- Retain latest value per key
- Useful for changelog streams
- Enables log as table semantics

#### Compaction Process
1. Identify latest value per key
2. Delete older values
3. Clean up tombstone records
4. Compact segments

### 6.6 Controller and Coordination

#### Controller
- One broker acts as controller
- Manages partition leadership
- Handles broker failures
- Coordinates topic creation/deletion

#### Leader Election
- Controller selects new leader from ISR
- Updates metadata
- Notifies producers and consumers

#### Metadata Management
- Topic and partition metadata
- Broker membership
- Consumer group coordination
- Stored in internal topics

## 7. Scaling Strategy

### 7.1 Horizontal Scaling

#### Adding Brokers
1. Add new broker to cluster
2. Controller detects new broker
3. Rebalance partitions across brokers
4. Migrate partition data
5. Update metadata

#### Adding Partitions
1. Increase partition count for topic
2. Controller assigns new partitions
3. Producers start using new partitions
4. Consumers rebalance to assign new partitions

#### Adding Consumers
1. New consumer joins group
2. Group coordinator triggers rebalance
3. Partitions reassigned
4. New consumer starts consuming

#### Scaling Producers
- Add more producer instances
- No coordination needed
- Automatic load distribution

### 7.2 Vertical Scaling
- Increase broker resources (CPU, RAM, disk)
- Optimize disk I/O with SSD
- Increase network bandwidth
- Tune JVM parameters

### 7.3 Geographic Distribution
- **Multi-region clusters**: Separate clusters per region
- **Replication across regions**: MirrorMaker for cross-region
- **Geo-replication**: Async replication for disaster recovery
- **Regional consumers**: Consumers read from local cluster

## 8. Reliability Strategy

### 8.1 Replication
- **Replication factor**: 3 for production
- **ISR mechanism**: Only in-sync replicas can become leader
- **Leader election**: Automatic on failure
- **Ack configuration**: Choose durability level

### 8.2 Failure Detection
- **Controller health checks**: Monitor broker availability
- **ISR monitoring**: Track replica lag
- **Heartbeat mechanism**: Detect broker failures
- **Session timeout**: Configurable failure detection

### 8.3 Data Durability
- **Sync replication**: For critical data
- **Async replication**: For high throughput
- **Commit to disk**: Configurable flush interval
- **Replication factor**: Choose based on durability requirements

### 8.4 Disaster Recovery
- **Multi-region deployment**: Separate clusters per region
- **MirrorMaker**: Cross-region replication
- **Backup snapshots**: Periodic cluster backups
- **Failover testing**: Regular disaster recovery drills

## 9. Tradeoffs

### 9.1 Throughput vs Latency
| Configuration | High Throughput | Low Latency |
|---------------|-----------------|-------------|
| Batch Size | Large (100KB) | Small (1KB) |
| Compression | Enabled | Disabled |
| Acks | 1 | 0 |
| linger_ms | Higher (10ms) | Lower (0ms) |

**Decision: Configure based on use case**

### 9.2 Durability vs Latency
| Acks Config | Durability | Latency | Use Case |
|-------------|------------|---------|----------|
| 0 | Low | Lowest | Non-critical data |
| 1 | Medium | Low | Default |
| all | High | High | Critical data |

**Decision: acks=1 for most use cases, acks=all for critical data**

### 9.3 Retention vs Cost
| Retention | Cost | Replay Capability | Use Case |
|-----------|------|-------------------|----------|
| Short (1 day) | Low | Limited | Real-time processing |
| Medium (7 days) | Medium | Good | Standard processing |
| Long (30 days) | High | Excellent | Analytics |
| Infinite | Very High | Unlimited | Event sourcing |

**Decision: Based on replay requirements**

### 9.4 Partition Count
| Partitions | Parallelism | Overhead | Use Case |
|------------|-------------|----------|----------|
| Few (1-10) | Low | Low | Low throughput |
| Medium (10-100) | Medium | Medium | Standard |
| Many (100+) | High | High | High throughput |

**Decision: Start with 10-100, scale based on throughput**

## 10. Bottlenecks

### 10.1 Current Bottlenecks
1. **Network bandwidth**: High replication and consumer traffic
2. **Disk I/O**: Log segment writes and compaction
3. **Broker CPU**: Message serialization and compression
4. **Controller coordination**: Rebalancing overhead

### 10.2 Mitigation Strategies

#### Network Bandwidth Bottleneck
- Compression to reduce data size
- Optimize replication factor
- Use faster network (10GbE)
- Geographic distribution for local consumers

#### Disk I/O Bottleneck
- Use SSD storage
- Optimize segment size
- Tune fsync interval
- Multiple disk volumes per broker

#### Broker CPU Bottleneck
- Optimize compression algorithm
- Increase broker count
- Use more powerful hardware
- Tune JVM garbage collection

#### Controller Coordination Bottleneck
- Optimize rebalancing algorithm
- Reduce consumer group churn
- Increase controller resources
- Use incremental cooperative rebalancing

## 11. Follow-ups

### 11.1 Exactly-Once Semantics
**Question**: How do you achieve exactly-once semantics?
**Answer**:
- Idempotent producers
- Transactional writes
- Idempotent consumers
- External system coordination
- Performance trade-offs

### 11.2 Schema Evolution
**Question**: How do you handle schema evolution?
**Answer**:
- Schema registry integration
- Avro/Protobuf serialization
- Forward and backward compatibility
- Schema validation
- Migration strategies

### 11.3 Dead Letter Queues
**Question**: How do you handle failed messages?
**Answer**:
- Dead letter topic per consumer
- Retry with exponential backoff
- Error handling strategies
- Monitoring and alerting
- Manual intervention

### 11.4 Security
**Question**: How do you secure the cluster?
**Answer**:
- TLS for network encryption
- SASL for authentication
- ACLs for authorization
- Network isolation (VPC)
- Audit logging

### 11.5 Monitoring
**Question**: How do you monitor cluster health?
**Answer**:
- Broker metrics (under-replicated partitions, offline partitions)
- Producer metrics (retry rate, error rate)
- Consumer metrics (lag, commit rate)
- Controller metrics
- Resource utilization

### 11.6 Consumer Lag
**Question**: How do you handle consumer lag?
**Answer**:
- Monitor lag metrics
- Scale consumer group
- Optimize consumer processing
- Increase partition count
- Alert on high lag

### 11.7 Hot Partitions
**Question**: How do you handle hot partitions?
**Answer**:
- Monitor partition-level metrics
- Increase partition count
- Better partitioning strategy
- Producer-side load balancing
- Consumer scaling

### 11.8 Compact Topics
**Question**: When should you use compacted topics?
**Answer**:
- Changelog data
- Latest value per key
- State table representation
- KTable in Kafka Streams
- Database change logs

### 11.9 Cross-Region Replication
**Question**: How do you replicate across regions?
**Answer**:
- MirrorMaker 2.0
- Confluent Replicator
- Async replication
- Conflict resolution
- Latency considerations

### 11.10 Stream Processing
**Question**: How do you integrate with stream processing?
**Answer**:
- Kafka Streams API
- ksqlDB
- Flink/Spark Streaming
- Exactly-once processing
- State management

---

## Summary

The Kafka-like queue system is designed as a distributed, partitioned, replicated commit log:
- **Log-based storage** enables message replay and multiple consumer groups
- **Partitioning** provides parallelism and horizontal scalability
- **Replication** ensures fault tolerance and durability
- **Consumer groups** allow independent consumption with offset tracking
- **Pull-based consumption** gives consumers control over their pace
- **Configurable durability** through acks and replication settings

The key insight is that Kafka is a log, not a traditional queue. Messages are retained and can be replayed, enabling multiple independent consumer groups to read the same stream at their own pace. This design prioritizes throughput, durability, and replay capability over simple queue semantics.