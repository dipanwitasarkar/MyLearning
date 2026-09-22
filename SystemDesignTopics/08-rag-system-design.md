# RAG (Retrieval-Augmented Generation) System Design

## Overview
Retrieval-Augmented Generation (RAG) is a technique that enhances Large Language Models (LLMs) by retrieving relevant information from external knowledge bases and including it in the prompt context. This approach addresses LLM limitations like knowledge cutoffs, hallucinations, and lack of domain-specific knowledge.

## Why RAG?

### Benefits
- **Current Information**: Access to up-to-date information beyond training data
- **Reduced Hallucinations**: Grounded responses in retrieved documents
- **Domain Knowledge**: Incorporate specialized knowledge bases
- **Transparency**: Source citations and traceability
- **Customization**: Tailor responses to specific domains
- **Cost Efficiency**: Smaller models with RAG vs larger fine-tuned models

### When to Use
- Question answering over documents
- Knowledge base integration
- Customer support systems
- Research and analysis
- Document summarization
- Fact-checking and verification

## RAG Architecture

### Basic RAG Pipeline
```
User Query → Query Processing → Vector Search → Context Retrieval → LLM Generation → Response
```

### Advanced RAG Pipeline
```
User Query → Query Processing → Hybrid Search → Re-ranking → Context Assembly → LLM Generation → Response
                ↓                    ↓
            Query Expansion    Knowledge Graph
```

## Core Components

### 1. Document Processing Pipeline
Convert raw documents into searchable embeddings.

**Architecture:**
```
Raw Documents → Text Extraction → Chunking → Embedding Generation → Vector Store
```

**Implementation:**
```python
from langchain.text_splitter import RecursiveCharacterTextSplitter
from sentence_transformers import SentenceTransformer
import chromadb

class DocumentProcessor:
    def __init__(self, embedding_model="all-MiniLM-L6-v2"):
        self.embedder = SentenceTransformer(embedding_model)
        self.text_splitter = RecursiveCharacterTextSplitter(
            chunk_size=1000,
            chunk_overlap=200,
            length_function=len
        )
        self.vector_store = chromadb.Client()
    
    def process_document(self, document):
        # Extract text
        text = self.extract_text(document)
        
        # Split into chunks
        chunks = self.text_splitter.split_text(text)
        
        # Generate embeddings
        embeddings = self.embedder.encode(chunks)
        
        # Store in vector database
        collection = self.vector_store.get_or_create_collection("documents")
        collection.add(
            documents=chunks,
            embeddings=embeddings.tolist(),
            metadatas=[{"source": document.name} for _ in chunks]
        )
        
        return len(chunks)
    
    def extract_text(self, document):
        # Handle different document types
        if document.type == "pdf":
            return self.extract_from_pdf(document)
        elif document.type == "docx":
            return self.extract_from_docx(document)
        else:
            return document.text
```

### 2. Query Processing
Enhance and optimize user queries for better retrieval.

**Techniques:**

#### Query Expansion
```python
class QueryExpander:
    def __init__(self, llm):
        self.llm = llm
    
    def expand_query(self, original_query):
        prompt = f"""
        Generate 3 alternative queries for: "{original_query}"
        Focus on different aspects and phrasings.
        """
        
        expanded_queries = self.llm.generate(prompt)
        return [original_query] + self.parse_expanded_queries(expanded_queries)
```

#### Query Rewriting
```python
class QueryRewriter:
    def rewrite_query(self, query, conversation_history):
        if conversation_history:
            prompt = f"""
            Conversation history: {conversation_history}
            Current query: {query}
            
            Rewrite the query to be self-contained and clear.
            """
            return self.llm.generate(prompt)
        return query
```

### 3. Retrieval System
Find relevant documents using various search strategies.

#### Vector Search
```python
class VectorRetriever:
    def __init__(self, vector_store, embedder):
        self.vector_store = vector_store
        self.embedder = embedder
    
    def retrieve(self, query, top_k=5):
        # Generate query embedding
        query_embedding = self.embedder.encode(query)
        
        # Vector search
        results = self.vector_store.query(
            query_embeddings=[query_embedding.tolist()],
            n_results=top_k
        )
        
        return results
```

#### Hybrid Search (Vector + Keyword)
```python
class HybridRetriever:
    def __init__(self, vector_store, keyword_search):
        self.vector_store = vector_store
        self.keyword_search = keyword_search
    
    def retrieve(self, query, top_k=5, alpha=0.5):
        # Vector search
        vector_results = self.vector_store.search(query, top_k=top_k*2)
        
        # Keyword search
        keyword_results = self.keyword_search.search(query, top_k=top_k*2)
        
        # Combine and re-rank
        combined_results = self.reciprocal_rank_fusion(
            vector_results, 
            keyword_results, 
            alpha=alpha
        )
        
        return combined_results[:top_k]
    
    def reciprocal_rank_fusion(self, results1, results2, alpha=0.5, k=60):
        scores = {}
        
        # Score from first result set
        for rank, doc in enumerate(results1):
            doc_id = doc['id']
            scores[doc_id] = scores.get(doc_id, 0) + alpha / (k + rank + 1)
        
        # Score from second result set
        for rank, doc in enumerate(results2):
            doc_id = doc['id']
            scores[doc_id] = scores.get(doc_id, 0) + (1 - alpha) / (k + rank + 1)
        
        # Sort by score
        ranked_results = sorted(scores.items(), key=lambda x: x[1], reverse=True)
        return [doc for doc, score in ranked_results]
```

#### Knowledge Graph Retrieval
```python
class KnowledgeGraphRetriever:
    def __init__(self, graph_db):
        self.graph_db = graph_db
    
    def retrieve(self, query, top_k=5):
        # Extract entities from query
        entities = self.extract_entities(query)
        
        # Query knowledge graph
        graph_results = []
        for entity in entities:
            # Find related nodes
            related = self.graph_db.find_related(entity, depth=2)
            graph_results.extend(related)
        
        return graph_results[:top_k]
```

### 4. Re-ranking
Improve retrieval quality by re-ranking initial results.

**Implementation:**
```python
class ReRanker:
    def __init__(self, reranker_model):
        self.reranker = reranker_model
    
    def rerank(self, query, documents, top_k=5):
        # Score each document
        scores = []
        for doc in documents:
            score = self.reranker.score(query, doc['text'])
            scores.append((doc, score))
        
        # Sort by score
        reranked = sorted(scores, key=lambda x: x[1], reverse=True)
        
        return [doc for doc, score in reranked[:top_k]]
```

### 5. Context Assembly
Combine retrieved documents into coherent context.

**Strategies:**

#### Simple Concatenation
```python
def simple_context_assembly(documents, max_tokens=2000):
    context = ""
    for doc in documents:
        if len(context) + len(doc['text']) < max_tokens:
            context += doc['text'] + "\n\n"
        else:
            break
    return context
```

#### Relevance-Based Assembly
```python
def relevance_context_assembly(query, documents, max_tokens=2000):
    # Score documents by relevance
    scored_docs = [(doc, relevance_score(query, doc)) for doc in documents]
    scored_docs.sort(key=lambda x: x[1], reverse=True)
    
    # Assemble context
    context = ""
    for doc, score in scored_docs:
        if len(context) + len(doc['text']) < max_tokens:
            context += f"[Relevance: {score:.2f}] {doc['text']}\n\n"
        else:
            break
    return context
```

#### Diversity-Based Assembly
```python
def diversity_context_assembly(documents, max_tokens=2000):
    # Select diverse documents
    selected = []
    remaining = documents.copy()
    
    while remaining and len(selected) < len(documents):
        # Select most different from already selected
        best_doc = None
        best_score = -1
        
        for doc in remaining:
            diversity_score = calculate_diversity(doc, selected)
            if diversity_score > best_score:
                best_score = diversity_score
                best_doc = doc
        
        if best_doc and len(context) + len(best_doc['text']) < max_tokens:
            selected.append(best_doc)
            remaining.remove(best_doc)
        else:
            break
    
    return "\n\n".join([doc['text'] for doc in selected])
```

### 6. LLM Generation
Generate responses using retrieved context.

**Implementation:**
```python
class RAGGenerator:
    def __init__(self, llm, prompt_template):
        self.llm = llm
        self.prompt_template = prompt_template
    
    def generate(self, query, context):
        # Construct prompt
        prompt = self.prompt_template.format(
            context=context,
            question=query
        )
        
        # Generate response
        response = self.llm.generate(prompt)
        
        return response
    
    def generate_with_citations(self, query, documents):
        # Add source information to context
        context_with_sources = []
        for doc in documents:
            context_with_sources.append(f"""
            Source: {doc['metadata']['source']}
            {doc['text']}
            """)
        
        context = "\n\n".join(context_with_sources)
        
        # Generate response
        response = self.generate(query, context)
        
        # Extract citations
        citations = self.extract_citations(response, documents)
        
        return {
            "response": response,
            "citations": citations,
            "sources": [doc['metadata']['source'] for doc in documents]
        }
```

## Advanced RAG Techniques

### 1. Multi-Query RAG
Generate multiple queries and retrieve documents for each.

**Implementation:**
```python
class MultiQueryRAG:
    def __init__(self, retriever, llm):
        self.retriever = retriever
        self.llm = llm
    
    def retrieve(self, query, num_queries=3):
        # Generate multiple queries
        queries = self.generate_queries(query, num_queries)
        
        # Retrieve for each query
        all_results = []
        for q in queries:
            results = self.retriever.retrieve(q)
            all_results.extend(results)
        
        # Deduplicate and re-rank
        unique_results = self.deduplicate(all_results)
        reranked = self.rerank(query, unique_results)
        
        return reranked
```

### 2. Recursive RAG
Iteratively retrieve and generate to improve results.

**Implementation:**
```python
class RecursiveRAG:
    def __init__(self, retriever, generator, max_iterations=3):
        self.retriever = retriever
        self.generator = generator
        self.max_iterations = max_iterations
    
    def generate(self, query, context=""):
        for iteration in range(self.max_iterations):
            # Generate response
            response = self.generator.generate(query, context)
            
            # Check if more information needed
            if self.needs_more_info(response):
                # Retrieve additional context
                new_context = self.retriever.retrieve(query)
                context += "\n\n" + new_context
            else:
                return response
        
        return response
```

### 3. Agentic RAG
Use agents to decide retrieval strategy.

**Implementation:**
```python
class AgenticRAG:
    def __init__(self, retrievers, generator):
        self.retrievers = retrievers
        self.generator = generator
    
    def process_query(self, query):
        # Analyze query type
        query_type = self.analyze_query(query)
        
        # Select appropriate retriever
        retriever = self.select_retriever(query_type)
        
        # Retrieve context
        context = retriever.retrieve(query)
        
        # Generate response
        response = self.generator.generate(query, context)
        
        return response
```

## Vector Database Options

### 1. ChromaDB
Open-source vector database built for AI applications.

**Setup:**
```python
import chromadb

# Create client
client = chromadb.Client()

# Create collection
collection = client.create_collection("documents")

# Add documents
collection.add(
    documents=["doc1", "doc2"],
    embeddings=[[1.0, 2.0], [3.0, 4.0]],
    metadatas=[{"source": "file1"}, {"source": "file2"}]
)

# Query
results = collection.query(
    query_embeddings=[[1.1, 2.1]],
    n_results=2
)
```

**Pros:**
- Open source
- Easy to use
- Good for small to medium datasets
- Local deployment

**Cons:**
- Limited scalability
- Fewer advanced features
- Performance limitations at scale

### 2. Pinecone
Managed vector database service.

**Setup:**
```python
import pinecone

# Initialize
pinecone.init(api_key="your-api-key")
index = pinecone.Index("your-index")

# Upsert vectors
index.upsert([
    ("id1", [1.0, 2.0], {"source": "file1"}),
    ("id2", [3.0, 4.0], {"source": "file2"})
])

# Query
results = index.query(
    vector=[1.1, 2.1],
    top_k=2,
    include_metadata=True
)
```

**Pros:**
- Fully managed
- Scalable
- Good performance
- Advanced features

**Cons:**
- Cost at scale
- Vendor lock-in
- Limited control

### 3. Weaviate
Open-source vector database with GraphQL API.

**Setup:**
```python
import weaviate

# Create client
client = weaviate.Client("http://localhost:8080")

# Create schema
client.schema.create_class({
    "class": "Document",
    "properties": [
        {"name": "text", "dataType": ["text"]},
        {"name": "source", "dataType": ["string"]}
    ]
})

# Add data
client.data_object.create({
    "text": "document text",
    "source": "file1"
}, class_name="Document")

# Query
results = client.query.get(
    "Document",
    near_vector={"vector": [1.1, 2.1]},
    limit=2
).with_limit(2).do()
```

**Pros:**
- GraphQL API
- Rich features
- Good performance
- Modular architecture

**Cons:**
- Steeper learning curve
- Resource intensive
- Complex setup

## Embedding Models

### 1. Sentence Transformers
Open-source sentence embedding models.

**Models:**
- `all-MiniLM-L6-v2`: Fast, good quality
- `all-mpnet-base-v2`: Higher quality, slower
- `e5-large-v2`: Good for retrieval

**Implementation:**
```python
from sentence_transformers import SentenceTransformer

# Load model
model = SentenceTransformer('all-MiniLM-L6-v2')

# Generate embeddings
embeddings = model.encode([
    "This is a sentence.",
    "This is another sentence."
])
```

### 2. OpenAI Embeddings
Commercial embedding API.

**Implementation:**
```python
import openai

# Generate embeddings
response = openai.Embedding.create(
    input="Your text here",
    model="text-embedding-ada-002"
)

embedding = response['data'][0]['embedding']
```

**Pros:**
- High quality
- No infrastructure
- Easy to use

**Cons:**
- Cost at scale
- Vendor lock-in
- Network latency

### 3. Cohere Embeddings
Alternative commercial embedding service.

**Implementation:**
```python
import cohere

co = cohere.Client('your-api-key')

# Generate embeddings
response = co.embed(
    texts=["Your text here"],
    model='embed-english-v2.0'
)

embedding = response.embeddings[0]
```

## Evaluation Metrics

### 1. Retrieval Quality
Measure how well the system retrieves relevant documents.

**Metrics:**
- **Precision@K**: Percentage of retrieved documents that are relevant
- **Recall@K**: Percentage of relevant documents that are retrieved
- **MRR (Mean Reciprocal Rank)**: Average of reciprocal ranks of first relevant document
- **NDCG (Normalized Discounted Cumulative Gain)**: Accounts for ranking position

**Implementation:**
```python
def evaluate_retrieval(retrieved_docs, relevant_docs, k=5):
    retrieved_ids = [doc['id'] for doc in retrieved_docs[:k]]
    relevant_ids = set(relevant_docs)
    
    # Precision@K
    precision = len(set(retrieved_ids) & relevant_ids) / k
    
    # Recall@K
    recall = len(set(retrieved_ids) & relevant_ids) / len(relevant_ids)
    
    # MRR
    mrr = 0
    for i, doc_id in enumerate(retrieved_ids):
        if doc_id in relevant_ids:
            mrr = 1 / (i + 1)
            break
    
    return {
        'precision': precision,
        'recall': recall,
        'mrr': mrr
    }
```

### 2. Generation Quality
Measure the quality of generated responses.

**Metrics:**
- **Faithfulness**: Does the response stick to retrieved context?
- **Answer Relevance**: Is the response relevant to the query?
- **Context Precision**: Are the retrieved contexts actually relevant?
- **Context Recall**: Did we retrieve all necessary contexts?

**Implementation:**
```python
def evaluate_generation(query, response, context):
    # Faithfulness - check if response is grounded in context
    faithfulness = check_faithfulness(response, context)
    
    # Answer relevance
    relevance = check_relevance(query, response)
    
    # Context precision
    context_precision = evaluate_context_precision(query, context)
    
    return {
        'faithfulness': faithfulness,
        'relevance': relevance,
        'context_precision': context_precision
    }
```

### 3. End-to-End Performance
Measure overall system performance.

**Metrics:**
- **Latency**: End-to-end response time
- **Throughput**: Requests per second
- **Cost**: Cost per query
- **User Satisfaction**: Human evaluation

## Optimization Strategies

### 1. Caching
Cache query results and embeddings.

**Implementation:**
```python
from functools import lru_cache
import hashlib

class RAGCache:
    def __init__(self):
        self.query_cache = {}
        self.embedding_cache = {}
    
    def get_cached_response(self, query):
        cache_key = hashlib.md5(query.encode()).hexdigest()
        return self.query_cache.get(cache_key)
    
    def cache_response(self, query, response):
        cache_key = hashlib.md5(query.encode()).hexdigest()
        self.query_cache[cache_key] = response
    
    @lru_cache(maxsize=1000)
    def get_embedding(self, text):
        return self.embedder.encode(text)
```

### 2. Async Processing
Process retrieval and generation asynchronously.

**Implementation:**
```python
import asyncio

class AsyncRAG:
    async def process_query(self, query):
        # Async retrieval
        retrieval_task = asyncio.create_task(self.async_retrieve(query))
        
        # Async other operations
        other_task = asyncio.create_task(self.async_preprocess(query))
        
        # Wait for both
        context = await retrieval_task
        preprocessed = await other_task
        
        # Generate response
        response = await self.async_generate(query, context)
        
        return response
```

### 3. Batch Processing
Process multiple queries together for efficiency.

**Implementation:**
```python
class BatchRAG:
    def process_batch(self, queries):
        # Batch embedding generation
        embeddings = self.embedder.encode_batch(queries)
        
        # Batch retrieval
        all_contexts = []
        for query, embedding in zip(queries, embeddings):
            contexts = self.vector_store.search(embedding)
            all_contexts.append(contexts)
        
        # Batch generation
        responses = self.llm.generate_batch(queries, all_contexts)
        
        return responses
```

## Common Pitfalls

### 1. Poor Chunking Strategy
Ineffective document chunking leads to poor retrieval.

**Solution:**
- Experiment with different chunk sizes
- Use semantic chunking
- Consider document structure
- Maintain context overlap

### 2. Low-Quality Retrieval
Retrieved documents are not relevant to the query.

**Solution:**
- Improve query processing
- Use hybrid search
- Implement re-ranking
- Fine-tune embedding models

### 3. Context Overload
Too much context confuses the LLM.

**Solution:**
- Limit context length
- Use re-ranking to select best documents
- Implement context compression
- Use longer context models

### 4. Hallucination
LLM generates information not in retrieved context.

**Solution:**
- Constrain responses to context
- Use faithfulness metrics
- Implement citation requirements
- Fine-tune for grounded generation

### 5. Scalability Issues
System doesn't scale with increasing data or queries.

**Solution:**
- Use scalable vector databases
- Implement caching
- Optimize embedding generation
- Consider distributed processing

## Best Practices

### 1. Start Simple
Begin with basic RAG before adding complexity.

### 2. Evaluate Continuously
Regularly evaluate retrieval and generation quality.

### 3. Monitor Performance
Track latency, throughput, and quality metrics.

### 4. Iterate on Chunking
Experiment with different chunking strategies.

### 5. Use Hybrid Search
Combine vector and keyword search for better results.

### 6. Implement Caching
Cache frequently accessed queries and embeddings.

### 7. Design for Scale
Choose components that can scale with your needs.

### 8. Consider Cost
Balance quality with cost considerations.

## Conclusion

RAG systems are powerful for building knowledge-intensive AI applications. By understanding the core components, advanced techniques, and best practices, you can design systems that effectively combine retrieval and generation to provide accurate, context-aware responses. The key is to start simple, evaluate continuously, and iterate based on your specific requirements and constraints. As RAG technology continues to evolve, staying current with new techniques and tools will be essential for building effective systems.