# System Design Topics

This folder contains comprehensive guides on various system design topics, covering fundamental concepts, patterns, and modern technologies including LLM and RAG systems.

## Topics Covered

### 1. [Load Balancing](./01-load-balancing.md)
- Types of load balancers (Layer 4 vs Layer 7)
- Load balancing algorithms (Round Robin, Least Connections, IP Hash)
- Health checks and session persistence
- SSL termination
- HAProxy and NGINX configurations
- Best practices and common pitfalls

### 2. [Caching Strategies](./02-caching-strategies.md)
- Cache types (Client-side, CDN, Application-level, Database)
- Caching patterns (Cache-Aside, Write-Through, Write-Behind, Refresh-Ahead)
- Cache invalidation strategies
- Eviction policies (LRU, LFU, FIFO)
- Distributed caching challenges
- Cache stampede and thundering herd prevention

### 3. [Data Partitioning](./03-data-partitioning.md)
- Partitioning strategies (Horizontal, Vertical, Directory-based, Hash, Range)
- Consistent hashing implementation
- Partitioning key selection guidelines
- Cross-partition query handling
- Partition rebalancing strategies
- Geo-partitioning and time-based partitioning

### 4. [Database Replication](./04-database-replication.md)
- Replication topologies (Master-Slave, Master-Master, Ring, Star)
- Replication methods (Statement-based, Row-based, Mixed)
- Replication modes (Synchronous, Asynchronous, Semi-synchronous)
- Replication lag monitoring and mitigation
- Failover and recovery strategies
- Conflict resolution techniques

### 5. [Consistency and CAP Theorem](./05-consistency-cap-theorem.md)
- CAP theorem trade-offs (CA, CP, AP)
- Consistency models (Strong, Eventual, Causal, Read Your Writes, Monotonic Reads)
- Quorum-based consistency
- Leader-based consistency
- Gossip protocols
- Conflict resolution strategies (LWW, Vector Clocks, CRDTs)

### 6. [Message Queues](./06-message-queues.md)
- Queue architectures (Point-to-Point, Publish-Subscribe, Request-Reply)
- Message queue implementations (RabbitMQ, Kafka, SQS, Redis Streams)
- Message patterns (Work Queue, Pub-Sub, Routing, Dead Letter Queue)
- Message guarantees (At-least-once, At-most-once, Exactly-once)
- Performance optimization (Batching, Compression, Prefetch)
- Monitoring and observability

### 7. [LLM System Design](./07-llm-system-design.md)
- LLM architecture patterns (Direct API, Self-Hosted, Hybrid)
- LLM inference optimization (Batching, Caching, Quantization, Distillation)
- Prompt engineering strategies (Zero-shot, Few-shot, Chain-of-Thought, Role-based)
- LLM system components (API Gateway, Model Router, Rate Limiter, Monitoring)
- LLM fine-tuning approaches (Full, LoRA, RAG)
- LLM system scaling and security considerations

### 8. [RAG System Design](./08-rag-system-design.md)
- RAG architecture and pipeline components
- Document processing and chunking strategies
- Query processing (Expansion, Rewriting)
- Retrieval systems (Vector Search, Hybrid Search, Knowledge Graph)
- Re-ranking and context assembly techniques
- Advanced RAG techniques (Multi-Query, Recursive, Agentic)
- Vector database options (ChromaDB, Pinecone, Weaviate)
- RAG evaluation metrics and optimization strategies

## How to Use These Materials

### For Learning
1. Start with fundamental topics (Load Balancing, Caching, Data Partitioning)
2. Progress to advanced concepts (Consistency, Replication, Message Queues)
3. Explore modern AI/ML topics (LLM System Design, RAG)

### For Interview Preparation
1. Read each topic thoroughly
2. Understand the trade-offs and when to use each approach
3. Practice explaining concepts in your own words
4. Work through the implementation examples
5. Be prepared to discuss pros/cons of different approaches

### For System Design
1. Use these as reference materials when designing systems
2. Apply the patterns and best practices to your projects
3. Consider the trade-offs mentioned in each topic
4. Adapt the examples to your specific requirements

## Key Concepts Across Topics

### Scalability
- Horizontal vs vertical scaling
- Load balancing and distribution
- Caching for performance
- Data partitioning for scale

### Reliability
- Replication for fault tolerance
- Health checks and monitoring
- Failure handling and recovery
- Redundancy and failover

### Performance
- Caching strategies
- Load balancing algorithms
- Query optimization
- Batching and compression

### Consistency
- CAP theorem trade-offs
- Consistency models
- Replication consistency
- Conflict resolution

### Modern AI/ML
- LLM integration patterns
- RAG architecture
- Vector databases
- Prompt engineering

## Additional Resources

### System Design Fundamentals
- "Designing Data-Intensive Applications" by Martin Kleppmann
- "System Design Interview" by Alex Xu
- "The System Design Interview" by Insider.io

### Distributed Systems
- "Distributed Systems: Principles and Paradigms" by Andrew Tanenbaum
- "DDIA" (Designing Data-Intensive Applications)
- "Distributed Systems for Fun and Profit"

### AI/ML System Design
- "Building Machine Learning Powered Applications" by Ameet Gohil
- "Designing Machine Learning Systems" by Chip Huyen
- "AI Engineering" by Chip Huyen

## Contributing

Feel free to add more topics or improve existing ones. Each topic should include:
- Clear explanation of concepts
- Practical examples and implementations
- Trade-offs and when to use each approach
- Best practices and common pitfalls
- Real-world use cases

## License

These materials are for educational purposes. Feel free to use them for learning and interview preparation.