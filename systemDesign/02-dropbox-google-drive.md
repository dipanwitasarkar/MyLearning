# Dropbox/Google Drive System Design

## 1. Clarify Requirements

### Functional Requirements
- Upload and download files of any size (KB to GB)
- Automatically sync files across all user devices
- Share files/folders with other users
- Maintain version history and restore previous versions
- Support offline access with cached files
- Conflict resolution for concurrent edits
- Real-time collaboration (optional for Google Drive)
- File and folder management (create, delete, rename, move)

### Non-Functional Requirements
- **Durability**: 99.999999999% (11 nines) - zero data loss
- **Availability**: 99.99% - files accessible globally
- **Sync Latency**: < 5 seconds for small files on good connection
- **Upload Performance**: Resumable uploads for large files
- **Storage Efficiency**: Block-level deduplication across all users
- **Scalability**: Petabytes of data, billions of files
- **Bandwidth Efficiency**: Delta sync for large file edits

### Clarifying Questions
- What is the expected number of users and files per user?
- What is the average file size and maximum file size?
- Do we need real-time collaboration or just file sync?
- What is the expected concurrent user count?
- Do we need to support file versioning and how many versions?
- What are the sharing requirements (public links, specific users)?
- Do we need to support encryption at rest?

## 2. Estimate Scale

### Traffic Assumptions
- 100M users
- 10K files per user average
- 1MB average file size
- 10% of files change daily
- 100M daily file uploads
- 3 devices per user average

### Storage Calculation
- Total files: 100M × 10K = 1 trillion files
- Raw storage: 1T × 1MB = 1EB (exabyte)
- With 70% deduplication: 300PB
- Metadata overhead: ~10% = 330PB total

### Bandwidth Calculation
- Daily uploads: 100M × 1MB = 100TB/day
- Daily sync: 300M devices × 10MB avg sync = 3PB/day
- Peak bandwidth: 10× average = 10TB/s during peak hours
- With delta sync (90% reduction): 1TB/s peak

### QPS Estimation
- Upload requests: 100M/day / 86400 = 1.2K req/s peak
- Download requests: 10× uploads = 12K req/s peak
- Metadata operations: 50× uploads = 60K req/s peak
- Sync requests: 300M devices polling every 30s = 10K req/s

### Block Storage Calculation
- 4MB block size
- 1EB / 4MB = 250B blocks
- With 70% deduplication: 75B unique blocks
- Block metadata: 100 bytes per block = 7.5TB metadata

## 3. APIs

### Upload File
```
POST /api/v1/files/upload
Request: {
  "file_name": "document.pdf",
  "file_size": 5242880,
  "parent_folder_id": "folder123",
  "block_hashes": ["hash1", "hash2", ...]
}
Response: {
  "file_id": "file456",
  "upload_urls": {
    "hash1": "https://s3.../presigned-url",
    "hash2": "https://s3.../presigned-url"
  },
  "missing_blocks": ["hash1", "hash2"]
}
```

### Complete Upload
```
POST /api/v1/files/upload/complete
Request: {
  "file_id": "file456",
  "block_hashes": ["hash1", "hash2", ...]
}
Response: {
  "file_id": "file456",
  "version": 1,
  "created_at": "2024-01-01T00:00:00Z"
}
```

### Download File
```
GET /api/v1/files/:file_id
Response: {
  "file_id": "file456",
  "file_name": "document.pdf",
  "file_size": 5242880,
  "block_hashes": ["hash1", "hash2", ...],
  "download_urls": {
    "hash1": "https://s3.../presigned-url",
    "hash2": "https://s3.../presigned-url"
  }
}
```

### Sync Changes
```
GET /api/v1/sync/changes?since_version=123
Response: {
  "changes": [
    {
      "type": "file_added",
      "file_id": "file456",
      "version": 124
    },
    {
      "type": "file_deleted",
      "file_id": "file789",
      "version": 125
    }
  ],
  "latest_version": 125
}
```

### Share File
```
POST /api/v1/files/:file_id/share
Request: {
  "share_type": "user",  // or "link" for public link
  "recipient_email": "user@example.com",
  "permission": "editor"  // or "viewer", "owner"
}
Response: {
  "share_id": "share123",
  "shared_url": "https://drive.google.com/..."
}
```

### Get File Versions
```
GET /api/v1/files/:file_id/versions
Response: {
  "versions": [
    {
      "version": 1,
      "created_at": "2024-01-01T00:00:00Z",
      "block_hashes": ["hash1", "hash2"]
    },
    {
      "version": 2,
      "created_at": "2024-01-02T00:00:00Z",
      "block_hashes": ["hash1", "hash3"]
    }
  ]
}
```

## 4. Data Model

### Users Table
```sql
CREATE TABLE users (
    user_id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    storage_quota BIGINT DEFAULT 107374182400,  -- 100GB
    storage_used BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    INDEX idx_email (email)
);
```

### Namespaces Table (Folders/Shared Spaces)
```sql
CREATE TABLE namespaces (
    namespace_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    parent_namespace_id BIGINT,
    shared_with_user_id BIGINT,
    permissions VARCHAR(50) DEFAULT 'owner',
    created_at TIMESTAMP DEFAULT NOW(),
    INDEX idx_user_id (user_id),
    INDEX idx_parent (parent_namespace_id),
    INDEX idx_shared (shared_with_user_id)
);
```

### Files Table
```sql
CREATE TABLE files (
    file_id BIGSERIAL PRIMARY KEY,
    namespace_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    size BIGINT NOT NULL,
    version INT NOT NULL,
    block_hashes TEXT[] NOT NULL,
    content_hash VARCHAR(64),
    created_at TIMESTAMP DEFAULT NOW(),
    modified_at TIMESTAMP DEFAULT NOW(),
    is_deleted BOOLEAN DEFAULT FALSE,
    INDEX idx_namespace (namespace_id),
    INDEX idx_name (namespace_id, name),
    INDEX idx_content_hash (content_hash)
);
```

### File Versions Table
```sql
CREATE TABLE file_versions (
    version_id BIGSERIAL PRIMARY KEY,
    file_id BIGINT NOT NULL,
    version INT NOT NULL,
    block_hashes TEXT[] NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    INDEX idx_file (file_id, version)
);
```

### Block Metadata Table
```sql
CREATE TABLE block_metadata (
    block_hash VARCHAR(64) PRIMARY KEY,
    size BIGINT NOT NULL,
    reference_count INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT NOW(),
    storage_location VARCHAR(255)
);
```

### Sharing Table
```sql
CREATE TABLE shares (
    share_id BIGSERIAL PRIMARY KEY,
    file_id BIGINT NOT NULL,
    owner_user_id BIGINT NOT NULL,
    shared_with_user_id BIGINT,
    share_url VARCHAR(255),
    permissions VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    INDEX idx_file (file_id),
    INDEX idx_owner (owner_user_id),
    INDEX idx_shared_with (shared_with_user_id)
);
```

### Device Registry Table
```sql
CREATE TABLE devices (
    device_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    device_name VARCHAR(255),
    device_type VARCHAR(50),
    last_synced_at TIMESTAMP,
    sync_token INT,
    INDEX idx_user (user_id)
);
```

## 5. High Level Design

```
┌─────────────────────────────────────────────────────────────┐
│                      Client Applications                     │
│  (Desktop, Mobile, Web - with local cache and sync engine)  │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                        API Gateway                           │
│              (Authentication, Rate Limiting, Routing)         │
└──────────────────────────┬──────────────────────────────────┘
                           │
           ┌───────────────┴───────────────┐
           ▼                               ▼
┌──────────────────────┐      ┌──────────────────────┐
│  Metadata Service    │      │  Block Storage       │
│                      │      │  Service             │
│  - User accounts     │      │                      │
│  - File hierarchy    │      │  - Block dedup       │
│  - Permissions       │      │  - Upload coord      │
│  - Version tracking  │      │  - Download coord    │
│  - Change detection  │      │  - Storage mgmt      │
└──────────┬───────────┘      └──────────┬───────────┘
           │                              │
           ▼                              ▼
┌──────────────────────┐      ┌──────────────────────┐
│  Sharded PostgreSQL  │      │  Object Storage      │
│  (Metadata DB)       │      │  (S3/GCS)            │
│                      │      │                      │
│  - Users             │      │  - Block storage     │
│  - Files             │      │  - CDN integration   │
│  - Namespaces        │      │  - Multi-region      │
│  - Shares            │      │  - Lifecycle mgmt    │
└──────────────────────┘      └──────────────────────┘
           │
           ▼
┌──────────────────────┐
│  Notification Service │
│                      │
│  - Real-time sync    │
│  - Push notifications│
│  - WebSocket mgmt    │
└──────────────────────┘
```

### Key Design Decisions
- **Metadata-Content Separation**: Critical for scalability and deduplication
- **Block-level storage**: Enables efficient deduplication and delta sync
- **Client-side chunking**: Reduces server load and enables resumable uploads
- **Async notification**: Decouples sync from file operations
- **Multi-tier storage**: Hot data on fast storage, cold data on archive

## 6. Deep Dive Components

### 6.1 File Chunking Strategy

#### Chunking Process
1. Client splits file into fixed-size blocks (4-8MB)
2. Compute SHA-256 hash for each block
3. Upload only blocks that don't already exist (deduplication)
4. Store block hashes in file metadata

#### Block Size Trade-offs
| Block Size | Deduplication | Metadata Overhead | Upload Granularity |
|------------|---------------|-------------------|-------------------|
| 1MB | Excellent | High | Fine-grained |
| 4MB | Good | Medium | Balanced |
| 8MB | Fair | Low | Coarse |
| 16MB | Poor | Very Low | Very Coarse |

**Decision: 4MB blocks based on empirical data**

#### Implementation
```python
def chunk_file(file_path, block_size=4*1024*1024):
    blocks = []
    with open(file_path, 'rb') as f:
        while True:
            block = f.read(block_size)
            if not block:
                break
            block_hash = hashlib.sha256(block).hexdigest()
            blocks.append({
                'hash': block_hash,
                'size': len(block),
                'data': block
            })
    return blocks
```

### 6.2 Deduplication Strategy

#### Content-Addressable Storage
- Block hash = storage key
- Same content = same storage location
- Immutable blocks (never modified)

#### Deduplication Flow
1. Client computes block hashes locally
2. Client sends hash list to metadata service
3. Metadata service checks which blocks already exist
4. Service returns list of missing blocks
5. Client uploads only missing blocks
6. Reference count updated for existing blocks

#### Benefits
- **Storage savings**: 70% deduplication in real-world Dropbox data
- **Bandwidth savings**: Only upload changed blocks
- **Faster uploads**: Skip existing blocks

### 6.3 Metadata Service Design

#### Database Sharding Strategy
- **Shard by user_id**: User-specific queries optimized
- **Shard by namespace_id**: Folder operations optimized
- **Read replicas**: For analytics and reporting

#### Query Optimization
- Index on frequently queried columns
- Materialized views for common aggregations
- Query result caching for hot paths

#### Transaction Management
- ACID transactions for file operations
- Optimistic locking for concurrent edits
- Two-phase commit for cross-shard operations

### 6.4 Sync Protocol

#### Change Detection
- Client maintains local file state
- Periodically polls for changes (or uses push notifications)
- Compare local metadata with server metadata
- Identify changed, added, and deleted files

#### Conflict Resolution Strategies

##### Last-Write-Wins (Binary Files)
```python
def resolve_conflict(file_id, local_version, server_version):
    if local_version.timestamp > server_version.timestamp:
        return local_version
    else:
        return server_version
```

##### Operational Transformation (Collaborative Docs)
- Transform operations to merge edits
- Preserve user intent
- Complex implementation

##### Three-Way Merge (Text Files)
- Use common ancestor version
- Merge changes line by line
- Insert conflict markers for unresolvable conflicts

##### User Prompt
- Detect conflicts
- Present options to user
- Let user choose resolution

#### Sync Optimization
- **Delta sync**: Only upload changed blocks
- **Batch operations**: Group multiple file changes
- **Compression**: Compress blocks before upload
- **Bandwidth throttling**: Respect network conditions

### 6.5 Notification Service

#### Real-time Sync Options
- **WebSocket**: Persistent connection for active clients
- **Push notifications**: For mobile devices
- **Long-polling**: Fallback for web clients
- **Email**: For inactive users

#### Change Notification Flow
1. Metadata service detects file change
2. Publish change event to message queue (Kafka)
3. Notification service consumes event
4. Push notification to all user's connected devices
5. Devices pull updated metadata and blocks

#### Fan-out Strategy
- For shared files, notify all users with access
- Use topic-based pub/sub for efficient fan-out
- Batch notifications for multiple changes

### 6.6 Storage Backend

#### Object Storage (S3/GCS)
- **Primary storage**: Hot data
- **Lifecycle policies**: Move old data to cheaper storage
- **Multi-region replication**: For durability and low latency
- **CDN integration**: For fast downloads

#### Storage Tiers
- **Hot storage**: Frequently accessed data (SSD)
- **Warm storage**: Infrequently accessed (HDD)
- **Cold storage**: Archives (Glacier/Deep Archive)

#### Pre-signed URLs
- Generate time-limited URLs for direct upload/download
- Bypass application servers for data transfer
- Reduce server load and improve performance

## 7. Scaling Strategy

### 7.1 Horizontal Scaling

#### Metadata Service
- Stateless service instances
- Auto-scaling based on CPU/memory
- Load balancer distributes requests

#### Database Scaling
- **Sharding**: Distribute by user_id or namespace_id
- **Read replicas**: Multiple read replicas for analytics
- **Connection pooling**: Efficient database connections

#### Block Storage Service
- **Service instances**: Multiple instances for upload/download coordination
- **Object storage**: S3/GCS handles horizontal scaling
- **CDN**: Global edge caching

#### Notification Service
- **Fan-out instances**: Multiple instances for push notifications
- **Message queue**: Kafka handles high throughput
- **WebSocket connections**: Distributed connection management

### 7.2 Vertical Scaling
- Increase database server resources
- Use SSD storage for faster I/O
- Optimize database queries and indexes
- Increase cache sizes

### 7.3 Geographic Distribution
- **Multi-region deployment**: Deploy in multiple regions
- **Data locality**: Store data close to users
- **Cross-region replication**: For disaster recovery
- **Regional read replicas**: For low-latency reads

## 8. Reliability Strategy

### 8.1 Data Durability
- **Multi-region replication**: Replicate data across regions
- **Erasure coding**: Protect against data loss
- **Regular backups**: Point-in-time recovery
- **Version history**: Restore previous versions

### 8.2 High Availability
- **Redundant components**: No single points of failure
- **Automatic failover**: Detect and replace failed components
- **Health checks**: Monitor all components
- **Circuit breakers**: Prevent cascade failures

### 8.3 Data Consistency
- **ACID transactions**: For metadata operations
- **Optimistic locking**: For concurrent edits
- **Conflict resolution**: Handle concurrent modifications
- **Eventual consistency**: For some operations

### 8.4 Disaster Recovery
- **RTO**: 1 hour
- **RPO**: 5 minutes
- **Failover testing**: Regular drills
- **Documentation**: Clear runbooks

## 9. Tradeoffs

### 9.1 Block Size
| Block Size | Deduplication | Metadata | Upload Granularity |
|------------|---------------|----------|-------------------|
| Small (1MB) | Excellent | High overhead | Fine-grained |
| Medium (4MB) | Good | Medium | Balanced |
| Large (16MB) | Poor | Low overhead | Coarse |

**Decision: 4MB for optimal balance**

### 9.2 Sync Frequency
| Frequency | Real-time | Server Load | Battery Usage |
|-----------|-----------|-------------|---------------|
| Real-time | Excellent | High | High |
| Periodic (30s) | Good | Medium | Medium |
| Periodic (5m) | Fair | Low | Low |

**Decision: Adaptive based on network and battery**

### 9.3 Conflict Resolution
| Strategy | Data Integrity | User Experience | Complexity |
|----------|----------------|-----------------|------------|
| Automatic | Variable | Good | Low |
| Manual | High | Poor | Low |
| Hybrid | High | Good | High |

**Decision: Hybrid with user prompts for conflicts**

### 9.4 Storage Cost vs Performance
| Storage Type | Cost | Performance | Use Case |
|--------------|------|-------------|----------|
| SSD | High | Excellent | Hot data |
| HDD | Medium | Good | Warm data |
| Archive | Low | Poor | Cold data |

**Decision: Tiered storage with lifecycle policies**

## 10. Bottlenecks

### 10.1 Current Bottlenecks
1. **Metadata database**: Can become bottleneck with many users
2. **Block storage coordination**: Upload/download coordination overhead
3. **Notification fan-out**: Shared files with many users
4. **Network bandwidth**: Large file transfers

### 10.2 Mitigation Strategies

#### Metadata Database Bottleneck
- Database sharding by user_id
- Read replicas for analytics queries
- Query optimization and indexing
- Caching frequently accessed metadata

#### Block Storage Coordination Bottleneck
- Pre-signed URLs for direct upload/download
- Async coordination processing
- Batch processing for multiple blocks
- CDN integration for downloads

#### Notification Fan-out Bottleneck
- Message queue for async fan-out
- Batch notifications for multiple changes
- Push notification service optimization
- Rate limiting per user

#### Network Bandwidth Bottleneck
- Delta sync for large file edits
- Compression for data transfer
- Adaptive quality based on network
- Background sync for large transfers

## 11. Follow-ups

### 11.1 Real-time Collaboration
**Question**: How do you support real-time collaboration like Google Docs?
**Answer**:
- Operational Transformation (OT) or CRDTs
- WebSocket connections for real-time updates
- Conflict-free replicated data types
- Presence awareness (cursors, user presence)
- Version control with merge capabilities

### 11.2 Encryption
**Question**: How do you handle encryption at rest and in transit?
**Answer**:
- TLS for all network communications
- Client-side encryption for sensitive data
- Server-side encryption for storage
- Key management service (KMS)
- Zero-knowledge architecture option

### 11.3 File Preview
**Question**: How do you generate file previews?
**Answer**:
- Async preview generation service
- Support multiple formats (images, PDF, documents)
- Cache previews at edge locations
- Progressive loading for large previews
- Fallback to generic icons

### 11.4 Search
**Question**: How do you implement file search?
**Answer**:
- Full-text search index (Elasticsearch)
- Metadata-based search
- Content-based search (OCR for images)
- Faceted search (filters, sorting)
- Search analytics and suggestions

### 11.5 Bandwidth Optimization
**Question**: How do you optimize bandwidth usage?
**Answer**:
- Delta sync for large file edits
- Compression for data transfer
- Adaptive quality based on network
- Background sync for large transfers
- Differential sync for binary files

### 11.6 Offline Support
**Question**: How do you handle offline access?
**Answer**:
- Local cache with metadata
- Conflict resolution on reconnect
- Operation queue for offline changes
- Sync priority based on file importance
- Storage quota management

### 11.7 Large File Handling
**Question**: How do you handle very large files (GB+)?
**Answer**:
- Larger block sizes for large files
- Parallel block upload/download
- Resumable transfers
- Streaming for preview/playback
- Progressive upload for video files

### 11.8 Sharing Permissions
**Question**: How do you implement complex sharing permissions?
**Answer**:
- ACL-based permission model
- Inheritance for nested folders
- Permission propagation
- Expiration for shared links
- Audit logging for shared files

### 11.9 Storage Quotas
**Question**: How do you enforce storage quotas?
**Answer**:
- Per-user quota tracking
- Pre-upload quota check
- Quota exceeded handling
- Storage cleanup suggestions
- Upgrade prompts

### 11.10 Mobile Optimization
**Question**: How do you optimize for mobile devices?
**Answer**:
- Adaptive sync based on network
- Battery-aware sync scheduling
- Thumbnail generation for images
- Selective sync for large folders
- Background sync restrictions

---

## Summary

The Dropbox/Google Drive system is designed with metadata-content separation as the key architectural decision:
- **Block-level storage** enables efficient deduplication (70% savings)
- **Delta sync** only uploads changed blocks for large file edits
- **Multi-tier storage** balances cost and performance
- **Conflict resolution** handles concurrent edits
- **Real-time sync** keeps devices updated within seconds
- **Horizontal scaling** at all layers for petabyte-scale storage

The system separates metadata (file hierarchy, permissions, versions) from content (actual file bytes), allowing independent scaling and optimization of each component.