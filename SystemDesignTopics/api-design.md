# API Design

API design principles and patterns for system design interviews: REST, GraphQL, RPC, pagination, idempotency, versioning, and security.

In system design interviews, you'll want to define how clients interact with your system as part of the API step in the Delivery framework.

API design follows predictable patterns. You'll pick a protocol, define your resources, and specify how clients pass data and get responses back. This article covers all the basics needed to impress in this 5-minute period of your system design interview.

Before we go deep here, I want to make one thing super clear: most interviewers don't care about your API design being perfect. They want to see that you can design a reasonable API and move on to the more complex parts of your system.

That said, if you're interviewing for frontend or product roles, API design matters more since you'll be working closely with APIs daily. Also, for junior roles, there's less expectation on your ability to design distributed systems, so there may be more time spent in the interview on APIs.

## API Types

In an interview, you'll typically choose between three main API protocols:

1. **REST (Representational State Transfer)** - REST uses standard HTTP methods (GET, POST, PUT, DELETE) to manipulate resources identified by URLs. For standard CRUD operations in web and mobile applications, REST maps naturally to your database operations and HTTP semantics, making it the go-to protocol for most web services. This should be your default choice.

2. **GraphQL** - Unlike REST's fixed endpoints, GraphQL uses a single endpoint with a query language that lets clients specify exactly what data they need. Think about a mobile app that needs only basic user information versus a web dashboard that displays comprehensive analytics - with REST, you'd either create multiple endpoints or force clients to fetch more data than they need, but GraphQL lets each client request exactly what it needs in a single query. If your interviewer mentions "flexible data fetching" or talks about avoiding over-fetching and under-fetching, they're signaling you to consider GraphQL.

3. **RPC (Remote Procedure Call)** - RPC protocols like gRPC use binary serialization and HTTP/2 for efficient communication between services. While REST treats everything as resources, RPC lets you think in terms of actions and procedures - when your user service needs to quickly validate permissions with your auth service, an RPC call like checkPermission(userId, resource) is more natural than trying to model this as a REST resource. If the interviewer specifically mentions microservices or internal APIs, consider RPC for those high-performance connections. Use RPC when performance is critical.

Default to REST unless you have a specific reason not to. It's well-understood, has great tooling, and works for 90% of use cases. If you're unsure, just say "I'll use REST APIs" and move on.

For real-time features like notifications, chat, or live updates, you'll need different protocols like WebSockets or Server-Sent Events. These aren't traditional APIs - they're persistent connections.

### REST

Since REST is your default choice, let's spend most of our time here understanding how to design REST APIs that work well in system design interviews.

#### Resource Modeling

The foundation of good REST API design is identifying your resources correctly. If you've followed the delivery framework, resources are just your core entities.

Take Ticketmaster as an example. Your core entities might be events, venues, tickets, and bookings. These naturally map to REST resources:

```
GET /events                    # Get all events
GET /events/{id}               # Get a specific event
GET /venues/{id}               # Get a specific venue
GET /events/{id}/tickets       # Get available tickets for an event
POST /events/{id}/bookings     # Create a new booking for an event
GET /bookings/{id}             # Get a specific booking
```

Importantly, REST resources should represent *things* in your system, not *actions*. Instead of thinking about what users can do (like "book" or "purchase"), think about what exists in your system (events, venues, tickets, bookings).

Resources should always be plural nouns, i.e., bookings, events, tickets, etc. Most interviewers don't care about this, but some do, and it's easy enough to get right, so you might as well.

When handling relationships between resources, you have two main approaches. You can nest resources when there's a clear parent-child relationship, like /events/{id}/tickets for all tickets belonging to a specific event. Alternatively, you can keep resources flat and use query parameters for filtering, like /tickets?event_id=123.

The key difference is whether the relationship is required or optional. Use path parameters (or nested resources) when the value is required - like /events/{id}/tickets where you always need to specify which event's tickets you want. Use query parameters when the filter is optional - like /tickets?event_id=123&section=VIP where you might want all tickets, or tickets filtered by event, or tickets filtered by both event and section.

In interviews, this distinction helps you make the right choice quickly. If the relationship is always required for the query to make sense, use a path parameter. If it's an optional filter among many possible filters, use a query parameter.

#### HTTP Methods

Once you've identified your resources, you need to decide how clients interact with them. HTTP provides a set of methods (verbs) that map naturally to common operations, and understanding when to use each one is crucial for interviews.

**GET** is for retrieving data without changing anything. Use GET for /events/{id} to fetch event details or /events to list all events.

**POST** creates new resources. When a user books tickets, you'd POST to /events/{id}/bookings with the booking details in the request body. The server assigns an ID and returns the newly created booking. POST is neither safe nor idempotent. In other words, calling it multiple times creates multiple bookings.

**PUT** replaces an entire resource with what you send, or creates it if it doesn't exist. If you're updating a user's profile completely, PUT to /users/{id} with the full user object. Unlike POST, PUT is idempotent, so sending the same data multiple times results in the same final state.

**PATCH** updates part of a resource. When a user changes just their email address, PATCH to /users/{id} with only the email field. Unlike PUT, PATCH is not guaranteed to be idempotent - the result depends on how you implement it. A PATCH that says "set email to X" is idempotent, but a PATCH that says "append to list" is not.

**DELETE** removes a resource. DELETE /bookings/{id} cancels a booking. It's idempotent because repeated calls leave the server in the same state (the resource stays deleted), even if the response codes differ (first call might return 204, subsequent calls might return 404).

#### Status Codes

HTTP status codes tell the client what happened with their request. You don't need to memorize all of them, but knowing the common ones helps you communicate clearly:

- **200 OK** - Request succeeded
- **201 Created** - Resource created successfully (usually after POST)
- **204 No Content** - Request succeeded but no content to return (usually after DELETE)
- **400 Bad Request** - Client sent invalid data
- **401 Unauthorized** - Client needs to authenticate
- **403 Forbidden** - Client authenticated but not allowed
- **404 Not Found** - Resource doesn't exist
- **409 Conflict** - Request conflicts with current state (e.g., duplicate)
- **500 Internal Server Error** - Something went wrong on the server

#### Pagination

When you have large datasets, you can't return everything at once. Pagination lets clients request data in chunks.

Common pagination strategies:

**Offset-based pagination**: Use `?offset=0&limit=50` to get the first 50 items, then `?offset=50&limit=50` for the next 50. Simple but inefficient for large offsets.

**Cursor-based pagination**: Return a cursor (often an encoded timestamp or ID) with each page. Clients pass the cursor to get the next page. More efficient but slightly more complex.

**Keyset pagination**: Use a column like `?last_id=123&limit=50` to get items after ID 123. Efficient and simple.

In interviews, offset-based is fine to start with. If the interviewer pushes on performance, mention cursor-based as an optimization.

#### Idempotency

Idempotency means that making the same request multiple times has the same effect as making it once. This is crucial for network reliability - if a client's request times out, they can safely retry it.

GET, PUT, and DELETE are idempotent by design. POST is not. For POST operations that need to be idempotent (like creating a booking), you can use idempotency keys:

```
POST /bookings
Headers: Idempotency-Key: uuid-123
```

The server checks if it has already processed a request with this key. If yes, it returns the cached response instead of creating a duplicate.

#### Versioning

APIs evolve over time. Versioning lets you make changes without breaking existing clients.

Common versioning strategies:

**URL versioning**: `/v1/events`, `/v2/events` - Simple and clear
**Header versioning**: `Accept: application/vnd.api.v1+json` - Cleaner URLs
**Query parameter versioning**: `/events?version=1` - Less common

In interviews, URL versioning is the easiest to explain and understand.

#### Security

Basic security considerations for APIs:

**Authentication**: Verify who the client is (API keys, OAuth tokens, JWTs)
**Authorization**: Verify what the client is allowed to do (role-based access, permissions)
**Rate limiting**: Prevent abuse by limiting request rates
**HTTPS**: Encrypt all traffic in transit
**Input validation**: Never trust client input
**Output sanitization**: Never return sensitive data

In interviews, mention authentication and HTTPS as baseline security. Add rate limiting if the system is public-facing.

## GraphQL

GraphQL is worth mentioning when the interviewer talks about flexible data fetching or avoiding over-fetching/under-fetching.

Key advantages:
- Clients request exactly what they need
- Single endpoint for all operations
- Strongly typed schema
- Real-time subscriptions

Key considerations:
- More complex to set up
- Caching is harder (no HTTP caching)
- Over-fetching can still happen on the server side
- N+1 query problem if not careful

## RPC/gRPC

RPC is worth mentioning for internal microservices communication or when performance is critical.

Key advantages:
- Binary serialization (smaller payloads, faster)
- HTTP/2 support (multiplexing, streaming)
- Strongly typed contracts (protobuf)
- Built-in code generation

Key considerations:
- Not human-readable
- Requires code generation
- Less tooling than REST
- Overkill for simple CRUD

## Summary

For system design interviews:
1. Default to REST APIs
2. Identify resources (nouns, not actions)
3. Use appropriate HTTP methods
4. Handle pagination for large datasets
5. Consider idempotency for POST operations
6. Plan for versioning
7. Mention basic security (auth, HTTPS, rate limiting)
8. Only bring up GraphQL or RPC if the problem specifically calls for them