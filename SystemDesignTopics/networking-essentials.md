# Networking Essentials

Learn the important parts of networking that you'll need to know for your system design interviews.

Networking is a fundamental part of system design: you're nearly always going to be designing systems comprised of independent devices that communicate over a network. But the field of networking is vast and complex, and it's easy to get lost.

In this guide we're going to cover the **most important parts** of networking that you'll need to know for your system design interviews. In later deep dives, patterns, and problem breakdowns, we'll build on these basics to solve for the problems you'll face as you design your systems.

To do this, we'll start with the fundamentals of how networks operate, then examine key protocols at different layers of the networking stack. For each concept, we'll cover its purpose, how it works, and when to apply it in your system designs.

Networking tends to be a stronger focus in infrastructure and distributed systems interviews. For full-stack and product-focused roles, you'll likely only need a surface understanding of networking concepts. Understanding these fundamentals will help you make better decisions, even if the minute details aren't going to be tested in your interviews.

### Networking 101

At its core, networking is about connecting devices and enabling them to communicate. Networks are built on a layered architecture (the so-called "OSI model") which greatly simplifies the world for us application developers who sit on top of it.

Effectively, network layers are just abstractions that allow us to reason about the communication between devices in simpler terms relevant to our application. This way, when you're requesting a webpage, you don't need to know which voltages represent a 1 or a 0 on the network wire — you just need to know how to use the next layer down the stack.

#### Networking Layers

While the full networking stack is fascinating, there are three key layers that come up most often in system design interviews.

**Network Layer (Layer 3)**

At this layer is IP, the protocol that handles routing and addressing. It's responsible for breaking the data into packets, handling packet forwarding between networks, and providing best-effort delivery to any destination IP address on the network. While there are other protocols at this layer (like InfiniBand, which is used extensively for massive ML training workloads), IP by far the most common for system design interviews.

**Transport Layer (Layer 4)**

At this layer, we have TCP, QUIC, and UDP, which provide end-to-end communication services. Think of them like a layer that provides features like reliability, ordering, and flow control on top of the network layer.

**Application Layer (Layer 7)**

At the final layer are the application protocols like DNS, HTTP, Websockets, WebRTC. These are common protocols that build on top of TCP (or UDP, in the case of WebRTC) to provide a layer of abstraction for different types of data typically associated with web applications.

These layers work together to enable all our network communications. To see how they interact in practice, let's walk through a concrete example of how a simple web request works.

#### Example: A Simple Web Request

When you type a URL into your browser, several layers of networking protocols spring into action. Let's break down how these layers work together to retrieve a simple web page over HTTP on TCP.

First, we use DNS to convert a human-readable domain name like example.com into an IP address like 32.42.52.62. Then, a series of carefully orchestrated steps begins. We set up a TCP connection over IP, send our HTTP request, get a response, and tear down the connection.

In detail:

1. **DNS Resolution**: The client starts by resolving the domain name of the website to an IP address using DNS (Domain Name System).

2. **TCP Handshake**: The client initiates a TCP connection with the server using a three-way handshake:
   - **SYN**: The client sends a SYN (synchronize) packet to the server to request a connection.
   - **SYN-ACK**: The server responds with a SYN-ACK (synchronize-acknowledge) packet to acknowledge the request.
   - **ACK**: The client sends an ACK (acknowledge) packet to establish the connection.

3. **HTTP Request**: Once the TCP connection is established, the client sends an HTTP GET request to the server to request the web page.

4. **Server Processing**: The server processes the request, retrieves the requested web page, and prepares an HTTP response. (This is usually the only latency most SWE's think about and control!)

5. **HTTP Response**: The server sends the HTTP response back to the client, which includes the requested web page content.

6. **TCP Teardown**: After the data transfer is complete, the client and server close the TCP connection using a four-way handshake:
   - **FIN**: The client sends a FIN (finish) packet to the server to terminate the connection.
   - **ACK**: The server acknowledges the FIN packet with an ACK.
   - **FIN**: The server sends a FIN packet to the client to terminate its side of the connection.
   - **ACK**: The client acknowledges the server's FIN packet with an ACK.

It's less common recently in BigTech, but it used to be a popular interview question to ask candidates to dive into the details of "what happens when you type (e.g.) hellointerview.com into your browser and press enter?".

Details like these aren't typically a part of a system design interview but it's helpful to understand the basics of networking. It may save you some headaches on the job!

While the specific details of TCP handshakes and teardowns might seem too esoteric to apply to interviews, there's a few things to observe which we'll build upon:

First, as an application developer we are able to simplify our mental models dramatically. The application can take for granted that the data is transmitted with a degree of reliability and ordering: the TCP layer ensures that the data is delivered correctly and in order, and will provide a response to the application if it doesn't arrive. We also never have to concern ourselves with finding a specific server in the world and driving a pulse train of electrons to get there. With DNS, we can look up the IP address, and with IP the various networking hardware between us, our ISP, backbone providers, etc. can route the data to the destination.

Second, while we have one conceptual "request" and "response" here, there were many more packets and requests exchanged between servers to make it happen. All of these introduce latency that we can ignore ... until we can't. The higher in the stack we go, the more latency and processing required. This is relevant for our load balancer discussion shortly!

Finally note that the connection between client and server is stateful while it's open: both sides need to track the connection state. This is important for understanding how connection pooling works and why it's valuable.

### Key Networking Concepts for System Design

#### DNS (Domain Name System)

DNS is the phonebook of the internet. It translates human-readable domain names (like hellointerview.com) into IP addresses (like 32.42.52.62) that computers can understand.

**How it works:**
- When you type a domain name into your browser, your computer first checks its local DNS cache
- If not found locally, it queries a DNS resolver (usually provided by your ISP)
- The resolver then queries the DNS hierarchy to find the authoritative name server for that domain
- The authoritative name server returns the IP address

**System Design Considerations:**
- DNS has caching built-in, which reduces load on DNS servers
- DNS records have TTL (Time To Live) values that control how long they're cached
- DNS can be used for load balancing through techniques like round-robin DNS
- DNS is a single point of failure if not properly designed with redundancy

#### TCP (Transmission Control Protocol)

TCP is a reliable, connection-oriented protocol that provides ordered, error-checked delivery of data between applications.

**Key Features:**
- **Reliability**: TCP ensures that data arrives intact and in order
- **Flow Control**: TCP manages the rate of data transmission to prevent overwhelming the receiver
- **Congestion Control**: TCP adjusts transmission rates based on network conditions
- **Connection-oriented**: TCP establishes a connection before data transfer (three-way handshake)

**System Design Considerations:**
- TCP's reliability comes at the cost of overhead and latency
- For real-time applications where some packet loss is acceptable, UDP might be preferable
- TCP connection establishment adds latency (three-way handshake)
- Connection pooling can help reduce the overhead of establishing new connections

#### UDP (User Datagram Protocol)

UDP is a simple, connectionless protocol that provides unreliable, unordered delivery of data.

**Key Features:**
- **Unreliable**: UDP does not guarantee delivery or ordering
- **Low overhead**: UDP has minimal protocol overhead compared to TCP
- **Connectionless**: UDP does not establish connections before sending data
- **Fast**: UDP is faster than TCP due to its simplicity

**System Design Considerations:**
- UDP is ideal for real-time applications like video streaming, gaming, and VoIP
- UDP is used when speed is more important than reliability
- Applications using UDP must implement their own reliability mechanisms if needed
- UDP is commonly used for DNS queries

#### HTTP/HTTPS

HTTP (Hypertext Transfer Protocol) is the foundation of data communication on the World Wide Web. HTTPS is the secure version of HTTP, which uses TLS/SSL for encryption.

**Key Features:**
- **Request-Response Model**: HTTP follows a client-server model where clients send requests and servers send responses
- **Stateless**: HTTP is stateless, meaning each request is independent
- **Methods**: HTTP defines methods like GET, POST, PUT, DELETE, etc.
- **Status Codes**: HTTP uses status codes to indicate the result of requests (200 OK, 404 Not Found, etc.)

**System Design Considerations:**
- HTTP/1.1 introduced persistent connections to reduce connection overhead
- HTTP/2 introduced multiplexing to allow multiple requests over a single connection
- HTTP/3 uses QUIC (based on UDP) instead of TCP for improved performance
- HTTPS adds encryption overhead but is essential for security
- HTTP caching can significantly improve performance

#### Load Balancing

Load balancing distributes incoming network traffic across multiple servers to ensure no single server bears too much demand.

**Types of Load Balancing:**
- **Layer 4 Load Balancing**: Based on IP address and port (transport layer)
- **Layer 7 Load Balancing**: Based on application layer data like HTTP headers, URLs, etc.
- **Software Load Balancing**: Using software like Nginx, HAProxy
- **Hardware Load Balancing**: Using dedicated hardware appliances

**Load Balancing Algorithms:**
- **Round Robin**: Distributes requests evenly across servers
- **Least Connections**: Routes requests to the server with the fewest active connections
- **IP Hash**: Routes requests based on client IP address
- **Weighted Round Robin**: Distributes requests based on server capacity

**System Design Considerations:**
- Load balancers can be a single point of failure if not properly designed
- Health checks are essential to ensure traffic is only routed to healthy servers
- Session persistence (sticky sessions) may be required for stateful applications
- Load balancers can be used for SSL termination to reduce load on application servers

#### CDNs (Content Delivery Networks)

CDNs are distributed networks of servers that deliver content to users based on their geographic location.

**Key Benefits:**
- **Reduced Latency**: Content is served from servers closer to users
- **Reduced Load**: Origin servers experience less load
- **Improved Availability**: CDNs can handle traffic spikes and DDoS attacks
- **Caching**: CDNs cache content to reduce the need to fetch from origin servers

**System Design Considerations:**
- CDNs are essential for serving static assets (images, CSS, JavaScript)
- CDNs can also cache dynamic content with proper cache headers
- CDN invalidation strategies are important for ensuring fresh content
- CDNs add cost but significantly improve user experience

---

*This is Part 2. Part 3 covers Low-Level Design introduction.*