# Database Replication

## Overview
Database replication is the process of copying and maintaining database objects, such as tables, in multiple database systems that make up a distributed database system. Replication provides redundancy, fault tolerance, and improved performance for database systems.

## Why Database Replication?

### Benefits
- **High Availability**: If primary database fails, replica can take over
- **Disaster Recovery**: Replicas in different locations provide backup
- **Performance**: Read operations can be distributed across replicas
- **Scalability**: Handle more read traffic by adding replicas
- **Data locality**: Place replicas closer to users

### When to Use
- High availability requirements
- Read-heavy workloads
- Geographic distribution
- Disaster recovery needs
- Analytics workloads

## Replication Topologies

### 1. Master-Slave (Primary-Replica)
One primary database handles writes, multiple replicas handle reads.

**Architecture:**
```
Application → Primary DB (Read/Write)
           → Replica 1 (Read Only)
           → Replica 2 (Read Only)
           → Replica 3 (Read Only)
```

**Pros:**
- Simple to implement
- Clear separation of read/write
- Easy to understand
- Good for read-heavy workloads

**Cons:**
- Single point of failure for writes
- Replication lag
- Potential data inconsistency
- Manual failover required

**Best For:**
- Read-heavy applications
- Simple availability requirements
- When write availability is less critical

### 2. Master-Master (Multi-Master)
Multiple databases can accept writes, replicating changes to each other.

**Architecture:**
```
Application → Master A (Read/Write) ↔ Master B (Read/Write)
           → Master C (Read/Write)
```

**Pros:**
- No single point of failure
- Write scalability
- Geographic distribution
- Automatic failover

**Cons:**
- Complex conflict resolution
- Potential data conflicts
- Higher complexity
- More expensive

**Best For:**
- High availability requirements
- Geographic distribution
- Write-heavy workloads
- When conflicts are rare

### 3. Ring Replication
Each database replicates to the next in a ring.

**Architecture:**
```
Master A → Master B → Master C → Master A
```

**Pros:**
- No single point of failure
- Good for distributed systems
- Automatic failover

**Cons:**
- Complex conflict resolution
- Replication delays can compound
- Hard to troubleshoot
- Ring failure affects all nodes

**Best For:**
- Distributed systems
- When automatic failover is critical
- Multi-datacenter deployments

### 4. Star Replication
Central hub replicates to multiple satellites.

**Architecture:**
```
        Satellite A
            ↑
            │
Satellite B → Hub ← Satellite C
            │
            ↓
        Satellite D
```

**Pros:**
- Centralized management
- Good for hub-and-spoke architectures
- Simplified monitoring

**Cons:**
- Hub is single point of failure
- Hub can become bottleneck
- Latency for distant satellites

**Best For:**
- Centralized management requirements
- Hub-and-spoke network topology
- When central coordination is needed

## Replication Methods

### 1. Statement-Based Replication
Replicate SQL statements from primary to replica.

**Example:**
```sql
-- Primary executes
UPDATE users SET name = 'John' WHERE id = 1;

-- Statement replicated to replica
UPDATE users SET name = 'John' WHERE id = 1;
```

**Pros:**
- Simple to implement
- Low bandwidth usage
- Works well for deterministic statements

**Cons:**
- Non-deterministic functions cause issues
- Requires identical schema and data
- Side effects can cause inconsistencies
- Higher CPU usage on replicas

**Best For:**
- Simple applications
- Deterministic operations
- When bandwidth is limited

### 2. Row-Based Replication
Replicate actual row changes from primary to replica.

**Example:**
```sql
-- Primary executes
UPDATE users SET name = 'John' WHERE id = 1;

-- Row change replicated to replica
BEFORE: (id=1, name='Jane')
AFTER: (id=1, name='John')
```

**Pros:**
- More reliable for non-deterministic operations
- Less CPU usage on replicas
- Better for heterogeneous environments
- More precise replication

**Cons:**
- Higher bandwidth usage
- More complex implementation
- Larger binary logs
- Harder to debug

**Best For:**
- Complex applications
- Non-deterministic operations
- Heterogeneous environments

### 3. Mixed (Statement + Row-Based)
Use statement-based when possible, fall back to row-based when needed.

**Example:**
```sql
-- Simple statement: use statement-based
INSERT INTO users (name) VALUES ('John');

-- Complex statement: use row-based
UPDATE users SET last_login = NOW() WHERE id = 1;
```

**Pros:**
- Best of both worlds
- Optimizes for different scenarios
- Flexible approach

**Cons:**
- More complex configuration
- Harder to predict behavior
- Requires monitoring

**Best For:**
- Mixed workloads
- When optimization is needed
- Complex applications

## Replication Modes

### 1. Synchronous Replication
Primary waits for replica acknowledgment before committing.

**Flow:**
```
1. Application writes to primary
2. Primary writes to transaction log
3. Primary replicates to replica
4. Replica acknowledges
5. Primary commits and acknowledges application
```

**Implementation:**
```python
def synchronous_write(data):
    # Write to primary
    primary.write(data)
    
    # Wait for replica acknowledgment
    replica.acknowledge()
    
    # Commit on primary
    primary.commit()
    
    return True
```

**Pros:**
- Strong consistency
- No data loss on failover
- Reliable

**Cons:**
- Higher latency
- Reduced throughput
- Replica availability affects primary
- More complex

**Best For:**
- Financial transactions
- Critical data
- When consistency is paramount

### 2. Asynchronous Replication
Primary commits immediately, replicates in background.

**Flow:**
```
1. Application writes to primary
2. Primary writes to transaction log
3. Primary commits and acknowledges application
4. Primary replicates to replica in background
```

**Implementation:**
```python
def asynchronous_write(data):
    # Write to primary
    primary.write(data)
    
    # Commit immediately
    primary.commit()
    
    # Replicate in background
    background_replicate(data)
    
    return True
```

**Pros:**
- Lower latency
- Higher throughput
- Primary not blocked by replicas
- Simpler implementation

**Cons:**
- Potential data loss on failover
- Eventual consistency
- Replication lag
- Harder to recover

**Best For:**
- High-performance requirements
- When eventual consistency is acceptable
- Read-heavy workloads

### 3. Semi-Synchronous Replication
Primary waits for at least one replica acknowledgment.

**Flow:**
```
1. Application writes to primary
2. Primary writes to transaction log
3. Primary replicates to replicas
4. At least one replica acknowledges
5. Primary commits and acknowledges application
```

**Implementation:**
```python
def semi_synchronous_write(data):
    # Write to primary
    primary.write(data)
    
    # Wait for at least one replica
    wait_for_any_replica_acknowledgment()
    
    # Commit on primary
    primary.commit()
    
    return True
```

**Pros:**
- Balance between consistency and performance
- Some data protection
- Better than pure asynchronous

**Cons:**
- Still potential for data loss
- More complex than asynchronous
- Latency higher than asynchronous
- Configuration complexity

**Best For:**
- When some consistency is needed
- Balance between performance and reliability
- Most production systems

## Replication Lag

### What is Replication Lag?
Delay between primary and replica data consistency.

### Causes
- Network latency
- Replica processing time
- Heavy write load on primary
- Replica resource constraints

### Monitoring Replication Lag
```sql
-- MySQL
SHOW SLAVE STATUS\G
-- Look at Seconds_Behind_Master

-- PostgreSQL
SELECT pg_is_in_recovery(),
       pg_last_xact_replay_timestamp(),
       now() - pg_last_xact_replay_timestamp() AS lag;
```

### Mitigation Strategies

#### 1. Optimize Network
- Use dedicated network links
- Reduce network distance
- Optimize network configuration

#### 2. Optimize Replica Performance
- Use similar hardware as primary
- Optimize database configuration
- Monitor replica performance

#### 3. Reduce Write Load
- Implement caching
- Batch writes
- Optimize queries

#### 4. Use Parallel Replication
```sql
-- MySQL parallel replication
CHANGE MASTER TO
  MASTER_HOST='primary',
  MASTER_PARALLEL_WORKERS=4;
```

## Failover and Recovery

### Automatic Failover
System automatically promotes replica to primary on failure.

**Implementation:**
```python
class FailoverManager:
    def __init__(self, primary, replicas):
        self.primary = primary
        self.replicas = replicas
        self.monitor = HealthMonitor()
    
    def check_primary_health(self):
        return self.monitor.is_healthy(self.primary)
    
    def promote_replica(self):
        # Select most up-to-date replica
        best_replica = self.select_best_replica()
        
        # Promote to primary
        best_replica.promote_to_primary()
        
        # Update application configuration
        self.update_configuration(best_replica)
        
        # Reconfigure other replicas
        for replica in self.replicas:
            if replica != best_replica:
                replica.change_master(best_replica)
```

### Manual Failover
Administrator manually promotes replica to primary.

**Steps:**
1. Stop application writes
2. Ensure replica is up-to-date
3. Promote replica to primary
4. Update application configuration
5. Reconfigure other replicas
6. Resume application writes

### Data Recovery
Recover data from replicas or backups.

**Strategies:**
- Point-in-time recovery
- Incremental backups
- Binary log recovery
- Snapshot recovery

## Conflict Resolution

### Write Conflicts
When two primaries try to update same data simultaneously.

### Resolution Strategies

#### 1. Last Write Wins
Most recent write wins based on timestamp.

**Implementation:**
```python
def resolve_conflict(conflict):
    if conflict.write_a.timestamp > conflict.write_b.timestamp:
        return conflict.write_a
    else:
        return conflict.write_b
```

**Pros:**
- Simple to implement
- Deterministic
- Low overhead

**Cons:**
- Can lose data
- Doesn't consider business logic
- May not be appropriate for all cases

#### 2. First Write Wins
First write to be applied wins.

**Implementation:**
```python
def resolve_conflict(conflict):
    if conflict.write_a.applied_before(conflict.write_b):
        return conflict.write_a
    else:
        return conflict.write_b
```

**Pros:**
- Simple to implement
- Deterministic
- Low overhead

**Cons:**
- Network latency affects outcome
- May not be fair
- Doesn't consider business logic

#### 3. Custom Conflict Resolution
Application-specific conflict resolution logic.

**Implementation:**
```python
def resolve_conflict(conflict):
    # Custom business logic
    if conflict.field == 'balance':
        # Sum both writes
        return conflict.write_a.value + conflict.write_b.value
    elif conflict.field == 'status':
        # Use higher priority status
        return max(conflict.write_a.value, conflict.write_b.value)
    else:
        # Default to last write wins
        return conflict.write_a if conflict.write_a.timestamp > conflict.write_b.timestamp else conflict.write_b
```

**Pros:**
- Business logic aware
- Flexible
- Can handle complex scenarios

**Cons:**
- Complex to implement
- Application-specific
- Hard to maintain

#### 4. Manual Resolution
Conflicts flagged for manual resolution.

**Implementation:**
```python
def resolve_conflict(conflict):
    # Flag for manual review
    conflict_queue.add(conflict)
    notify_administrator(conflict)
    return None  # No automatic resolution
```

**Pros:**
- No data loss
- Human oversight
- Can handle complex scenarios

**Cons:**
- Requires human intervention
- Delays resolution
- Not scalable

## Replication in Different Databases

### 1. MySQL Replication

#### Configuration
```sql
-- Primary configuration
[mysqld]
server-id = 1
log-bin = mysql-bin
binlog-format = ROW

-- Replica configuration
[mysqld]
server-id = 2
relay-log = mysql-relay-bin
read-only = 1
```

#### Setup
```sql
-- On primary
CREATE USER 'repl'@'%' IDENTIFIED BY 'password';
GRANT REPLICATION SLAVE ON *.* TO 'repl'@'%';
FLUSH PRIVILEGES;
SHOW MASTER STATUS;

-- On replica
CHANGE MASTER TO
  MASTER_HOST='primary',
  MASTER_USER='repl',
  MASTER_PASSWORD='password',
  MASTER_LOG_FILE='mysql-bin.000001',
  MASTER_LOG_POS=154;
START SLAVE;
```

### 2. PostgreSQL Replication

#### Configuration
```bash
# Primary configuration
wal_level = replica
max_wal_senders = 5
max_replication_slots = 5

# Replica configuration
hot_standby = on
max_standby_streaming_delay = 30s
wal_receiver_status_interval = 10s
```

#### Setup
```sql
-- On primary
CREATE USER replicator WITH REPLICATION ENCRYPTED PASSWORD 'password';
SELECT * FROM pg_create_physical_replication_slot('replica_slot');

-- On replica
standby_mode = 'on'
primary_conninfo = 'host=primary port=5432 user=replicator password=password'
primary_slot_name = 'replica_slot'
```

### 3. MongoDB Replication

#### Configuration
```yaml
# mongod.conf for replica set
replication:
  replSetName: "myReplicaSet"
```

#### Setup
```javascript
// Initialize replica set
rs.initiate({
  _id: "myReplicaSet",
  members: [
    { _id: 0, host: "mongo1:27017" },
    { _id: 1, host: "mongo2:27017" },
    { _id: 2, host: "mongo3:27017" }
  ]
})
```

## Replication Monitoring

### Key Metrics
- **Replication Lag**: Delay between primary and replica
- **Replication Throughput**: Bytes/second replicated
- **Connection Status**: Replication connection health
- **Error Rate**: Replication error frequency
- **Disk Usage**: Binary log and relay log size

### Monitoring Implementation
```python
class ReplicationMonitor:
    def __init__(self, primary, replicas):
        self.primary = primary
        self.replicas = replicas
        self.metrics = {}
    
    def collect_metrics(self):
        for replica in self.replicas:
            self.metrics[replica.id] = {
                'lag': self.get_replication_lag(replica),
                'status': self.get_replication_status(replica),
                'throughput': self.get_replication_throughput(replica)
            }
    
    def get_replication_lag(self, replica):
        # Implementation varies by database
        return replica.get_lag()
    
    def alert_on_issues(self):
        for replica_id, metrics in self.metrics.items():
            if metrics['lag'] > 30:  # 30 seconds
                send_alert(f"High replication lag on {replica_id}")
            if metrics['status'] != 'running':
                send_alert(f"Replication stopped on {replica_id}")
```

## Common Pitfalls

### 1. Ignoring Replication Lag
Not monitoring or addressing replication lag.

**Solution:**
- Monitor replication lag continuously
- Set up alerts for high lag
- Optimize replica performance

### 2. Single Point of Failure
Relying on single primary without failover.

**Solution:**
- Implement automatic failover
- Use multi-master topology
- Regular failover testing

### 3. Network Partition Issues
Network partitions cause split-brain scenarios.

**Solution:**
- Implement quorum-based decisions
- Use fencing mechanisms
- Network partition detection

### 4. Inconsistent Configuration
Different configurations across replicas.

**Solution:**
- Configuration management
- Regular configuration audits
- Automated configuration deployment

### 5. Insufficient Testing
Not testing failover and recovery procedures.

**Solution:**
- Regular failover drills
- Test recovery procedures
- Document procedures

## Best Practices

### 1. Monitor Replication Health
Continuously monitor replication status and lag.

### 2. Test Failover Regularly
Regularly test failover procedures.

### 3. Use Appropriate Replication Mode
Choose replication mode based on consistency requirements.

### 4. Plan for Network Partitions
Design for network partition scenarios.

### 5. Implement Monitoring and Alerting
Comprehensive monitoring and alerting for replication issues.

### 6. Document Procedures
Maintain clear documentation for failover and recovery.

### 7. Use Consistent Configuration
Ensure consistent configuration across replicas.

### 8. Plan for Growth
Design replication to handle future growth.

## Conclusion

Database replication is essential for building highly available and scalable database systems. By understanding different replication topologies, methods, and modes, you can design database architectures that meet your availability, performance, and consistency requirements. The key is to choose the right replication strategy based on your specific needs and to implement proper monitoring and failover mechanisms.