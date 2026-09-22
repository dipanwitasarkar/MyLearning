# Notification Service System Design

## 1. Clarify Requirements

### Functional Requirements
- Support multiple channels: Push (APNs/FCM), Email, SMS, In-app
- Template management with dynamic variable substitution
- User preference management (opt-out per channel/category)
- Priority queuing (transactional > operational > marketing)
- Deduplication to prevent duplicate sends
- Delivery tracking and status updates
- Scheduled and recurring notifications
- Batch/broadcast notifications
- Multi-language support

### Non-Functional Requirements
- **Availability**: 99.99% (must accept events even if providers are down)
- **Latency**: Transactional push < 1s P99, email/SMS within seconds
- **Reliability**: At-least-once delivery for every accepted event
- **Scalability**: 100M daily notifications, 10x growth in 3 years
- **Cost Control**: Enforce per-channel rate limits
- **Throughput**: Handle traffic spikes 10× daily peak

### Clarifying Questions
- What is the expected daily notification volume?
- What are the latency requirements for different notification types?
- Do we need international support (multiple languages, time zones)?
- What are the cost constraints per channel?
- Do we need to support notification scheduling?
- What are the compliance requirements (GDPR, CAN-SPAM)?
- Do we need to support rich media (attachments, images)?

## 2. Estimate Scale

### Traffic Assumptions
- 100M daily notifications
- Channel distribution: 40% push, 30% email, 20% SMS, 10% in-app
- Priority distribution: 10% transactional, 30% operational, 60% marketing
- Peak traffic: 10× average = 1B notifications/day during peak
- Average notification size: 1KB

### Throughput Calculation
- Daily average: 100M / 86400 = 1.2K notifications/sec
- Peak average: 1B / 86400 = 11.6K notifications/sec
- Peak burst: 10× peak average = 116K notifications/sec

### Storage Requirements
- Notification records: 100M × 1KB = 100GB/day
- Delivery status: 100M × 500 bytes = 50GB/day
- Templates: 1K templates × 10KB = 10MB
- User preferences: 100M users × 1KB = 100GB
- Total daily: 250GB
- 30-day retention: 7.5TB

### Database Capacity
- Write operations: 1.2K/sec average, 116K/sec peak
- Read operations: 5× writes = 6K/sec average, 580K/sec peak
- Storage growth: 250GB/day

### Queue Capacity
- Queue depth: 1-hour buffer = 4.2M messages
- Message size: 1KB average
- Queue storage: 4.2GB

## 3. APIs

### Send Notification
```
POST /api/v1/notifications
Request: {
  "event_id": "uuid",
  "event_type": "order.shipped",
  "user_id": "user123",
  "data": {
    "order_id": "42",
    "tracking_url": "https://..."
  },
  "priority": "operational",
  "channels": ["push", "email"],
  "scheduled_at": "2024-01-01T10:00:00Z"  // optional
}
Response: 202 Accepted
{
  "notification_id": "notif123",
  "status": "queued",
  "estimated_delivery": "2024-01-01T10:00:05Z"
}
```

### Get Notification Status
```
GET /api/v1/notifications/:notification_id
Response: {
  "notification_id": "notif123",
  "status": "delivered",
  "channels": {
    "push": {
      "status": "delivered",
      "delivered_at": "2024-01-01T10:00:05Z",
      "provider_message_id": "apns-msg-123"
    },
    "email": {
      "status": "delivered",
      "delivered_at": "2024-01-01T10:00:08Z",
      "provider_message_id": "sendgrid-msg-456"
    }
  }
}
```

### Create Template
```
POST /api/v1/templates
Request: {
  "template_id": "order_shipped",
  "event_type": "order.shipped",
  "channels": {
    "email": {
      "subject": "Your order {{order_id}} has shipped",
      "body": "Your order is on its way. Track here: {{tracking_url}}"
    },
    "push": {
      "title": "Order Shipped",
      "body": "Your order {{order_id}} is on its way"
    }
  }
}
Response: {
  "template_id": "order_shipped",
  "created_at": "2024-01-01T00:00:00Z"
}
```

### Update User Preferences
```
PUT /api/v1/users/:user_id/preferences
Request: {
  "email_enabled": true,
  "sms_enabled": false,
  "push_enabled": true,
  "marketing_email": false,
  "marketing_sms": false,
  "quiet_hours_start": "22:00",
  "quiet_hours_end": "08:00",
  "timezone": "America/New_York"
}
Response: {
  "user_id": "user123",
  "preferences_updated": true
}
```

### Batch Send
```
POST /api/v1/notifications/batch
Request: {
  "campaign_id": "campaign123",
  "event_type": "marketing.promotion",
  "user_ids": ["user1", "user2", ...],
  "data": {
    "promotion_code": "SAVE20"
  },
  "channels": ["email"],
  "schedule": "2024-01-01T10:00:00Z"
}
Response: {
  "batch_id": "batch456",
  "total_recipients": 10000,
  "status": "scheduled"
}
```

### Webhook Handler
```
POST /webhooks/:provider
Request: {
  "event": "delivered",
  "message_id": "sendgrid-msg-456",
  "timestamp": "2024-01-01T10:00:08Z",
  "provider_data": {...}
}
Response: 200 OK
```

## 4. Data Model

### Notifications Table
```sql
CREATE TABLE notifications (
    notification_id BIGSERIAL PRIMARY KEY,
    event_id VARCHAR(255) UNIQUE NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    priority VARCHAR(20) NOT NULL, -- transactional, operational, marketing
    status VARCHAR(20) DEFAULT 'queued',
    data JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    scheduled_at TIMESTAMP,
    sent_at TIMESTAMP,
    INDEX idx_event_id (event_id),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_scheduled_at (scheduled_at)
);
```

### Delivery Status Table
```sql
CREATE TABLE delivery_status (
    delivery_id BIGSERIAL PRIMARY KEY,
    notification_id BIGINT NOT NULL,
    channel VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL, -- queued, sent, delivered, failed
    provider_message_id VARCHAR(255),
    error_code VARCHAR(50),
    error_message TEXT,
    attempts INT DEFAULT 0,
    timestamp TIMESTAMP DEFAULT NOW(),
    INDEX idx_notification (notification_id),
    INDEX idx_status (status, timestamp),
    INDEX idx_channel (channel)
);
```

### Templates Table
```sql
CREATE TABLE templates (
    template_id VARCHAR(255) PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    channel VARCHAR(50) NOT NULL,
    subject_template TEXT,
    body_template TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    version INT DEFAULT 1,
    INDEX idx_event_type (event_type),
    INDEX idx_channel (channel)
);
```

### User Preferences Table
```sql
CREATE TABLE user_preferences (
    user_id VARCHAR(255) PRIMARY KEY,
    email_enabled BOOLEAN DEFAULT TRUE,
    sms_enabled BOOLEAN DEFAULT TRUE,
    push_enabled BOOLEAN DEFAULT TRUE,
    in_app_enabled BOOLEAN DEFAULT TRUE,
    marketing_email BOOLEAN DEFAULT FALSE,
    marketing_sms BOOLEAN DEFAULT FALSE,
    marketing_push BOOLEAN DEFAULT FALSE,
    quiet_hours_start TIME,
    quiet_hours_end TIME,
    timezone VARCHAR(50) DEFAULT 'UTC',
    language VARCHAR(10) DEFAULT 'en',
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### Category Preferences Table
```sql
CREATE TABLE category_preferences (
    user_id VARCHAR(255),
    category VARCHAR(50),
    enabled BOOLEAN DEFAULT TRUE,
    PRIMARY KEY (user_id, category)
);
```

### Rate Limits Table
```sql
CREATE TABLE rate_limits (
    limit_id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255),
    channel VARCHAR(50),
    category VARCHAR(50),
    limit_per_hour INT NOT NULL,
    current_count INT DEFAULT 0,
    window_start TIMESTAMP DEFAULT NOW(),
    INDEX idx_user_channel (user_id, channel)
);
```

### Deduplication Cache (Redis)
```
Key: dedup:{event_id}
Value: notification_id
TTL: 24 hours
```

### Rate Limiting Cache (Redis)
```
Key: ratelimit:{user_id}:{channel}
Value: count
TTL: 1 hour (sliding window)
```

## 5. High Level Design

```
┌─────────────────────────────────────────────────────────────┐
│                    Producer Services                          │
│  (Payment, Auth, E-commerce, Social, etc.)                   │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    API Gateway                                │
│              (Rate limiting, Authentication)                 │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                 Notification API (Intake)                     │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Dedup Check  │  │ Pref Check   │  │ Validation   │      │
│  │   (Redis)    │  │  (Postgres)  │  │              │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    Message Queue (Kafka)                      │
│                                                              │
│  ┌────────────────┐  ┌────────────────┐  ┌──────────────┐  │
│  │ notifications  │  │ notifications  │  │ notifications│  │
│  │   -intake      │  │   -email       │  │   -sms       │  │
│  └────────────────┘  └────────────────┘  └──────────────┘  │
│                                                              │
│  ┌────────────────┐  ┌────────────────┐                      │
│  │ notifications  │  │ notifications  │                      │
│  │   -push        │  │   -in_app      │                      │
│  └────────────────┘  └────────────────┘                      │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                  Router Service                               │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Preference   │  │ Template     │  │ Priority     │      │
│  │ Resolution   │  │ Rendering    │  │ Routing      │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└──────────────────────────┬──────────────────────────────────┘
                           │
           ┌───────────────┼───────────────┐
           ▼               ▼               ▼
┌──────────────────┐ ┌──────────────┐ ┌──────────────┐
│  Email Workers   │ │ SMS Workers  │ │ Push Workers │
│                  │ │              │ │              │
│  - SendGrid      │ │ - Twilio     │ │ - APNs       │
│  - Mailgun       │ │ - Plivo      │ │ - FCM        │
└──────────────────┘ └──────────────┘ └──────────────┘
           │               │               │
           └───────────────┴───────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                  Webhook Handler                             │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Status       │  │ Retry Logic  │  │ Dead Letter   │      │
│  │ Updates      │  │              │  │ Queue        │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                  Status Database                             │
│              (PostgreSQL with read replicas)                  │
└─────────────────────────────────────────────────────────────┘
```

### Key Design Decisions
- **Async processing**: Queue-based for reliability and scalability
- **Per-channel queues**: Independent scaling per channel
- **Priority queuing**: Transactional messages processed first
- **Deduplication**: Prevent duplicate sends from retries
- **Preference resolution**: Respect user opt-out choices

## 6. Deep Dive Components

### 6.1 Deduplication Strategy

#### Idempotency Keys
- Use event_id as idempotency key
- Check Redis for existing notification
- TTL: 24 hours (adjust based on retry window)

#### Implementation
```python
def send_notification(event):
    # Check for duplicate
    if redis.exists(f"dedup:{event.event_id}"):
        return {"status": "duplicate", "notification_id": redis.get(f"dedup:{event.event_id}")}

    # Process notification
    notification_id = create_notification(event)

    # Store dedup key
    redis.setex(f"dedup:{event.event_id}", 86400, notification_id)

    return {"status": "queued", "notification_id": notification_id}
```

#### Benefits
- Prevents duplicate sends from producer retries
- Handles network failures gracefully
- Provides idempotency for producers

### 6.2 Preference Resolution

#### Resolution Logic
1. Check user-level channel preferences
2. Check category-level preferences
3. Check quiet hours
4. Apply priority overrides (transactional bypasses preferences)

#### Implementation
```python
def resolve_channels(user_id, event_type, priority, requested_channels):
    preferences = get_user_preferences(user_id)
    category_prefs = get_category_preferences(user_id, event_type)

    resolved_channels = []
    for channel in requested_channels:
        # Check user-level preference
        if not getattr(preferences, f"{channel}_enabled"):
            continue

        # Check category-level preference
        if not category_prefs.get(event_type, {}).get(channel, True):
            continue

        # Check quiet hours (except transactional)
        if priority != "transactional" and in_quiet_hours(preferences):
            continue

        resolved_channels.append(channel)

    return resolved_channels
```

#### Quiet Hours Handling
- Convert user timezone to UTC
- Check if current time is in quiet window
- Bypass for transactional messages
- Queue for later if in quiet hours

### 6.3 Template Management

#### Template Storage
- Versioned templates for rollback
- Support for multiple languages
- Variable substitution with validation

#### Template Rendering
```python
def render_template(template_id, data, language='en'):
    template = get_template(template_id, language)
    try:
        subject = render(template.subject_template, data)
        body = render(template.body_template, data)
        return {"subject": subject, "body": body}
    except TemplateError as e:
        log_error(e)
        return render_default_template(template_id, data)
```

#### Variable Validation
- Validate required variables are present
- Type checking for variables
- Sanitization to prevent injection attacks

### 6.4 Worker Pool Design

#### Per-Channel Workers

##### Email Workers
- Connect to SMTP or email API (SendGrid/Mailgun)
- Handle templating and formatting
- Implement retry logic with exponential backoff
- Process in batches for efficiency

##### SMS Workers
- Connect to SMS gateway (Twilio/Plivo)
- Handle message formatting (character limits)
- Implement carrier-specific handling
- Rate limiting per carrier

##### Push Workers
- Maintain persistent HTTP/2 connections to APNs/FCM
- Handle device token management
- Implement priority and TTL
- Batch requests for efficiency

#### Scaling
- Auto-scale based on queue depth
- Separate scaling per channel
- Priority queues for transactional messages
- Backpressure handling

### 6.5 Rate Limiting

#### Per-User Rate Limiting
- Prevent notification fatigue
- Implement sliding window counter
- Different limits per channel/category

#### Implementation
```python
def check_rate_limit(user_id, channel, category):
    key = f"ratelimit:{user_id}:{channel}:{category}"
    count = redis.incr(key)
    if count == 1:
        redis.expire(key, 3600)  # 1 hour window

    limit = get_limit(channel, category)
    return count <= limit
```

#### Per-Provider Rate Limiting
- Protect provider relationships
- Implement token bucket or leaky bucket
- Queue throttling when limits approached

#### Rate Limit Tiers
- Transactional: No limit (or very high)
- Operational: 100/hour per user
- Marketing: 10/hour per user

### 6.6 Delivery Tracking

#### Status Updates
- Receive webhooks from providers
- Update delivery status in database
- Trigger retries for failed deliveries

#### Webhook Handling
```python
def handle_webhook(provider, payload):
    delivery_id = extract_delivery_id(payload)
    status = extract_status(payload)

    update_delivery_status(delivery_id, {
        "status": status,
        "provider_message_id": payload.get("message_id"),
        "timestamp": payload.get("timestamp"),
        "error_code": payload.get("error_code"),
        "error_message": payload.get("error_message")
    })

    if status == "failed":
        schedule_retry(delivery_id)
```

#### Retry Logic
- Exponential backoff: 1min, 5min, 15min, 1hour, 3hours
- Max retries: 5 attempts
- Dead letter queue after max retries
- Different retry policies per channel

## 7. Scaling Strategy

### 7.1 Horizontal Scaling

#### API Layer
- Stateless API instances
- Auto-scaling based on request rate
- Load balancer distributes traffic

#### Queue Layer
- Kafka partition scaling
- Increase partitions for higher throughput
- Consumer group scaling

#### Worker Layer
- Auto-scale workers based on queue depth
- Separate scaling per channel
- Priority-based scaling

#### Database Layer
- Read replicas for status queries
- Connection pooling
- Query optimization

### 7.2 Vertical Scaling
- Increase worker instance size
- Optimize database queries
- Increase queue throughput
- Optimize network configuration

### 7.3 Geographic Distribution
- **Multi-region deployment**: Deploy in multiple regions
- **Regional queues**: Process notifications regionally
- **Provider optimization**: Use regional provider endpoints
- **Data locality**: Store preferences regionally

## 8. Reliability Strategy

### 8.1 Queue Buffering
- Kafka retains messages for configured retention
- Acts as buffer during provider outages
- Prevents data loss during failures

### 8.2 Retry Logic
- Exponential backoff for retries
- Max retry limits
- Dead letter queue for failed messages
- Alert on high failure rates

### 8.3 Provider Redundancy
- Multiple providers per channel
- Automatic failover between providers
- Provider health monitoring
- Load balancing across providers

### 8.4 Monitoring and Alerting
- Queue depth monitoring
- Delivery rate monitoring
- Failure rate monitoring
- Provider health monitoring
- Cost monitoring

## 9. Tradeoffs

### 9.1 Push vs Pull for Workers
| Aspect | Push (Kafka) | Pull (Polling) |
|--------|--------------|----------------|
| Complexity | Simple | More complex |
| Load Balancing | Automatic | Manual |
| Latency | Low | Higher |
| Control | Less | More |

**Decision: Push (Kafka consumers) for simplicity**

### 9.2 Synchronous vs Asynchronous
| Aspect | Synchronous | Asynchronous |
|--------|-------------|--------------|
| Latency | Higher | Lower |
| Feedback | Immediate | Delayed |
| Complexity | Lower | Higher |
| Reliability | Lower | Higher |

**Decision: Asynchronous for reliability**

### 9.3 Single Provider vs Multiple
| Aspect | Single | Multiple |
|--------|--------|----------|
| Complexity | Simple | Complex |
| Reliability | Lower | Higher |
| Cost | Lower | Higher |
| Negotiation | Single | Multiple |

**Decision: Multiple providers for critical channels**

### 9.4 Immediate vs Batch Delivery
| Aspect | Immediate | Batch |
|--------|-----------|-------|
| Latency | Low | High |
| Cost | High | Low |
| Throughput | Lower | Higher |

**Decision: Immediate for transactional, batch for marketing**

## 10. Bottlenecks

### 10.1 Current Bottlenecks
1. **Provider rate limits**: External provider limitations
2. **Queue backlog**: During provider outages
3. **Database writes**: Status update volume
4. **Template rendering**: CPU-intensive for complex templates

### 10.2 Mitigation Strategies

#### Provider Rate Limits
- Multiple providers with load balancing
- Queue throttling when limits approached
- Provider-specific rate limiting
- Negotiate higher limits with providers

#### Queue Backlog
- Auto-scale worker pools
- Priority queuing (process transactional first)
- Alert on queue depth thresholds
- Shed low-priority load if necessary

#### Database Write Bottleneck
- Batch status updates
- Async write operations
- Database sharding
- Read replicas for queries

#### Template Rendering Bottleneck
- Pre-render common templates
- Cache rendered templates
- Optimize template engine
- Distribute rendering across workers

## 11. Follow-ups

### 11.1 Multi-language Support
**Question**: How do you handle multiple languages?
**Answer**:
- Template versions per language
- User language preference
- Fallback to default language
- Translation management system
- RTL language support

### 11.2 Scheduled Notifications
**Question**: How do you handle scheduled notifications?
**Answer**:
- Delayed queue processing
- Scheduler service
- Timezone handling
- Batch processing for scheduled times
- Priority for scheduled notifications

### 11.3 Rich Media
**Question**: How do you handle attachments and images?
**Answer**:
- Object storage for media
- CDN for media delivery
- Media optimization
- Provider-specific handling
- Size limits and validation

### 11.4 Compliance
**Question**: How do you handle compliance (GDPR, CAN-SPAM)?
**Answer**:
- Consent tracking
- Unsubscribe mechanisms
- Data retention policies
- User data export/deletion
- Compliance reporting

### 11.5 Cost Optimization
**Question**: How do you optimize notification costs?
**Answer**:
- Channel selection based on cost
- Batch processing for marketing
- Provider cost comparison
- Usage analytics
- Budget alerts

### 11.6 A/B Testing
**Question**: How do you support A/B testing for notifications?
**Answer**:
- Template variants
- User segmentation
- Delivery time testing
- Performance tracking
- Statistical analysis

### 11.7 User Engagement
**Question**: How do you track user engagement?
**Answer**:
- Click tracking
- Open tracking (for email)
- Delivery analytics
- User response analytics
- Engagement dashboards

### 11.8 Provider Failover
**Question**: How do you handle provider failover?
**Answer**:
- Health monitoring
- Automatic failover
- Queue replay
- Provider-specific retry logic
- Failback strategies

### 11.9 Notification Batching
**Question**: How do you batch notifications efficiently?
**Answer**:
- Time-based batching
- User-based batching
- Channel-specific batching
- Priority-aware batching
- Batching analytics

### 11.10 Real-time Updates
**Question**: How do you provide real-time notification status?
**Answer**:
- WebSocket connections
- Server-sent events
- Polling fallback
- Status caching
- Real-time dashboards

---

## Summary

The notification service is designed as a reliable, scalable multi-channel messaging platform:
- **Async processing** with Kafka for reliability and buffering
- **Per-channel queues** for independent scaling and failure isolation
- **Priority queuing** to ensure critical messages are delivered first
- **Deduplication** to prevent duplicate sends from retries
- **Preference resolution** to respect user opt-out choices
- **Provider redundancy** to handle provider outages
- **Rate limiting** to protect both users and provider relationships

The system separates the intake layer from delivery, allowing it to accept events even when downstream providers are experiencing issues, with queues acting as buffers to absorb traffic spikes.