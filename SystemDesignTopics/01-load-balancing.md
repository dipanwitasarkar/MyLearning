# Load Balancing

## Overview
Load balancing is the process of distributing network traffic across multiple servers to ensure no single server bears too much demand. By spreading the work evenly, load balancers improve responsiveness and availability of applications.

## Why Load Balancing?

### Benefits
- **Scalability**: Handle increased traffic by adding more servers
- **Reliability**: If one server fails, traffic is redirected to healthy servers
- **Performance**: Distribute load to prevent any single server from becoming a bottleneck
- **Flexibility**: Easily add or remove servers without disrupting service

### When to Use
- High-traffic websites and applications
- Microservices architecture
- Any system requiring high availability
- Applications with variable traffic patterns

## Types of Load Balancers

### 1. Layer 4 (Transport Layer) Load Balancers
Operate at the transport layer (TCP/UDP) and make routing decisions based on IP addresses and ports.

**Pros:**
- Fast and efficient
- Low latency
- Less expensive
- Good for high-throughput scenarios

**Cons:**
- Limited visibility into application layer
- Can't make decisions based on content
- Limited routing options

**Use Cases:**
- General traffic distribution
- High-performance requirements
- Simple routing needs

**Example:**
```
Client → L4 Load Balancer → Server A (IP: 10.0.1.1)
                         → Server B (IP: 10.0.1.2)
                         → Server C (IP: 10.0.1.3)
```

### 2. Layer 7 (Application Layer) Load Balancers
Operate at the application layer and can make routing decisions based on content (HTTP headers, URLs, cookies).

**Pros:**
- Content-aware routing
- SSL termination
- Request/response manipulation
- Advanced routing rules

**Cons:**
- Higher latency
- More expensive
- More complex configuration
- Higher resource requirements

**Use Cases:**
- Content-based routing
- A/B testing
- Canary deployments
- Multi-tenant applications

**Example:**
```
Client → L7 Load Balancer → /api/users → User Service
                         → /api/orders → Order Service
                         → /static/* → CDN
```

## Load Balancing Algorithms

### 1. Round Robin
Distributes requests sequentially across servers.

**Algorithm:**
```
Request 1 → Server A
Request 2 → Server B
Request 3 → Server C
Request 4 → Server A (cycle repeats)
```

**Pros:**
- Simple to implement
- Fair distribution
- No server state needed

**Cons:**
- Doesn't consider server load
- Doesn't consider server capacity
- Can cause uneven distribution with varying request times

**Best For:**
- Servers with similar specifications
- Similar request processing times
- Simple scenarios

### 2. Weighted Round Robin
Similar to round robin but assigns weights to servers based on capacity.

**Algorithm:**
```
Server A (weight 3): Request 1, 2, 3
Server B (weight 2): Request 4, 5
Server C (weight 1): Request 6
```

**Pros:**
- Accounts for different server capacities
- Better resource utilization
- Still relatively simple

**Cons:**
- Requires manual weight configuration
- Weights may need adjustment over time
- Doesn't consider real-time load

**Best For:**
- Heterogeneous server environments
- Different server specifications
- Known capacity differences

### 3. Least Connections
Routes requests to the server with the fewest active connections.

**Algorithm:**
```
Server A: 5 connections
Server B: 2 connections
Server C: 8 connections

New request → Server B (fewest connections)
```

**Pros:**
- Considers current server load
- Better for varying request durations
- Dynamic load distribution

**Cons:**
- Requires connection tracking
- More complex implementation
- May not account for connection "weight"

**Best For:**
- Long-running connections
- Varying request processing times
- Real-time load balancing

### 4. IP Hash
Routes requests based on client IP address using a hash function.

**Algorithm:**
```
hash(client_ip) % num_servers = server_index

Client IP 192.168.1.1 → hash → Server A
Client IP 192.168.1.2 → hash → Server B
```

**Pros:**
- Session persistence (same client always hits same server)
- No additional storage needed
- Deterministic routing

**Cons:**
- Uneven distribution if IP distribution is skewed
- Doesn't adapt to server failures
- Can cause hot spots

**Best For:**
- Session-based applications
- Stateful services
- When session persistence is required

### 5. Least Response Time
Routes requests to the server with the lowest response time.

**Algorithm:**
```
Server A: avg response 50ms
Server B: avg response 120ms
Server C: avg response 80ms

New request → Server A (fastest response)
```

**Pros:**
- Considers actual server performance
- Adapts to changing conditions
- Optimizes for user experience

**Cons:**
- Requires response time tracking
- More complex implementation
- May oscillate between servers

**Best For:**
- Performance-critical applications
- Variable server performance
- User experience optimization

## Load Balancer Placement

### 1. Global Server Load Balancing (GSLB)
Distributes traffic across multiple data centers or geographic regions.

**Architecture:**
```
User → DNS/GSLB → US Data Center → L4/L7 LB → Servers
               → EU Data Center → L4/L7 LB → Servers
               → Asia Data Center → L4/L7 LB → Servers
```

**Benefits:**
- Geographic load distribution
- Disaster recovery
- Reduced latency for users
- Compliance with data residency

**Techniques:**
- DNS-based routing
- Anycast routing
- Proximity-based routing
- Health-based routing

### 2. Per-Data Center Load Balancing
Load balancing within a single data center.

**Architecture:**
```
User → CDN → Data Center → L4 LB → L7 LB → Application Servers
```

**Benefits:**
- Local load distribution
- High availability within data center
- Layered security
- Specialized load balancing at each layer

## Health Checks

### Purpose
Monitor server health and route traffic only to healthy servers.

### Types of Health Checks

#### 1. TCP Health Check
Attempt to establish TCP connection on specified port.

```python
def tcp_health_check(host, port, timeout=5):
    try:
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.settimeout(timeout)
        result = sock.connect_ex((host, port))
        sock.close()
        return result == 0
    except:
        return False
```

#### 2. HTTP Health Check
Make HTTP request to health endpoint.

```python
def http_health_check(url, timeout=5):
    try:
        response = requests.get(url, timeout=timeout)
        return response.status_code == 200
    except:
        return False
```

#### 3. Application-Level Health Check
Custom health check that verifies application dependencies.

```python
def application_health_check():
    try:
        # Check database connection
        db.ping()
        # Check cache connection
        cache.ping()
        # Check external services
        external_service.health()
        return True
    except:
        return False
```

### Health Check Configuration
```yaml
health_check:
  protocol: http
  path: /health
  interval: 30s
  timeout: 5s
  unhealthy_threshold: 3
  healthy_threshold: 2
```

## Session Persistence

### Why Session Persistence?
Some applications require that a user's requests always go to the same server (session affinity).

### Techniques

#### 1. Sticky Sessions (Cookie-based)
Load balancer adds a cookie to identify the server.

```python
# Load balancer logic
def route_request(request):
    session_cookie = request.cookies.get('SERVER_ID')
    if session_cookie:
        return get_server(session_cookie.value)
    else:
        server = select_server()
        response.set_cookie('SERVER_ID', server.id)
        return server
```

#### 2. IP Hash
Use client IP hash for routing (as described above).

#### 3. Session Replication
Replicate session data across all servers.

```python
# Session replication architecture
class SessionManager:
    def save_session(self, session_id, data):
        # Save to local cache
        local_cache.set(session_id, data)
        # Replicate to other servers
        for server in other_servers:
            async_replicate(server, session_id, data)
```

## SSL Termination

### What is SSL Termination?
Decrypting SSL/TLS traffic at the load balancer instead of application servers.

### Architecture
```
Client (HTTPS) → Load Balancer (SSL Termination) → Application Servers (HTTP)
```

### Benefits
- Reduces load on application servers
- Centralized certificate management
- Simplified application server configuration
- Better performance with hardware acceleration

### Implementation
```python
# SSL termination configuration
ssl_config = {
    'certificate': '/path/to/cert.pem',
    'private_key': '/path/to/key.pem',
    'protocols': ['TLSv1.2', 'TLSv1.3'],
    'ciphers': ['ECDHE-RSA-AES128-GCM-SHA256']
}
```

## Load Balancer Technologies

### 1. Hardware Load Balancers
Dedicated hardware devices (F5, Citrix ADC).

**Pros:**
- High performance
- Advanced features
- Reliable
- Vendor support

**Cons:**
- Expensive
- Less flexible
- Vendor lock-in
- Longer deployment cycles

### 2. Software Load Balancers
Software solutions (HAProxy, NGINX, Envoy).

**Pros:**
- Cost-effective
- Flexible
- Open-source options
- Easy to deploy

**Cons:**
- Requires server resources
- May have lower performance
- Requires maintenance

### 3. Cloud Load Balancers
Cloud provider solutions (AWS ALB/ELB, GCP Load Balancing).

**Pros:**
- Managed service
- Auto-scaling
- Integration with cloud services
- Pay-as-you-go

**Cons:**
- Vendor lock-in
- May have limitations
- Cost at scale

## Example Implementations

### HAProxy Configuration
```haproxy
frontend web_frontend
    bind *:80
    bind *:443 ssl crt /etc/ssl/certs/mycert.pem
    default_backend web_servers

backend web_servers
    balance roundrobin
    server web1 10.0.1.1:80 check
    server web2 10.0.1.2:80 check
    server web3 10.0.1.3:80 check backup
```

### NGINX Configuration
```nginx
upstream backend {
    least_conn;
    server 10.0.1.1:80;
    server 10.0.1.2:80;
    server 10.0.1.3:80;
}

server {
    listen 80;
    location / {
        proxy_pass http://backend;
    }
}
```

## Best Practices

### 1. Use Multiple Load Balancers
Deploy load balancers in pairs for high availability.

### 2. Implement Health Checks
Regular health checks to route traffic only to healthy servers.

### 3. Monitor Performance
Track metrics like response time, error rates, and connection counts.

### 4. Plan for Failures
Design for load balancer failures (active-passive or active-active).

### 5. Use Appropriate Algorithms
Choose the right algorithm based on your use case.

### 6. Consider SSL Termination
Terminate SSL at load balancer to reduce application server load.

### 7. Implement Rate Limiting
Protect against DDoS attacks and abuse.

### 8. Regular Testing
Test failover scenarios and configuration changes.

## Common Pitfalls

### 1. Single Point of Failure
Using a single load balancer creates a single point of failure.

### 2. Improper Health Checks
Health checks that don't accurately reflect application health.

### 3. Overloading Load Balancers
Not scaling load balancers with increased traffic.

### 4. Ignoring Session Persistence
Not handling session requirements properly.

### 5. Poor Algorithm Choice
Using an inappropriate load balancing algorithm.

## Conclusion

Load balancing is a critical component of modern distributed systems. By understanding the different types, algorithms, and best practices, you can design systems that are scalable, reliable, and performant. The right load balancing strategy depends on your specific requirements, traffic patterns, and infrastructure constraints.