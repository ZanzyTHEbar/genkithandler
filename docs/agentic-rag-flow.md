# Agentic RAG Flow

OpenAI's Agentic RAG Flow is a framework that combines retrieval-augmented generation (RAG) with agentic capabilities. This allows for the generation of responses based on both pre-existing knowledge and real-time information retrieval.

## Key Components

1. **Load the Entire Document** into the context window.
   1. Logic to determine the selected models context window
   2. Appropriately and intelligently manage context window using advanced prompting techniques
2. **Chunk the Document** into 20 chunks that respect sentence boundaries.
   1. Late Chunking Strategy
   2. Genkit Embedding API
   3. Turso Driver for storing Vectors & Semantic information & Metadata
3. **Prompt the model** for which chunks might contain relevant information.
4. **Drill down** into the selected relevant chunks.
5. **Recursively call this function** until we reach paragraph-level content.
6. **Build Knowledge Graph** based on context
   1. If memory is enabled
   2. JSON structure
   3. Custom prompt instructions
7. **Build scratch pad**
   1. If memory is enabled
   2. markdown documents acting as "memory bank" between calls
   3. compressed after write, uncompressed before read
   4. custom prompt instructions
8. **Verify the answer** for factual accuracy.
9.  **Generate a response** based on the retrieved information & context.

- Advanced Prompt Engineering techniques
- Dynamic Prompt Templating and Chaining
- Generate Custom Prompts
- Send Prompt to LLM
- Concurrent Processing - Goroutines and Channels
- Custom Agent Module
- Inter-Agent Messaging - Message Bus / Actor Model
- Mixture of Agents - Ensemble and Iterative Refinement
- RAG Workflow Integration - Vector Retrieval and Context Augmentation
- LLM Final Response Generation
- Structured Output - JSON Schema Validation
- Observability and Metrics Logging
- Self-Optimization and Adaptive Tuning
- Robust Error Handling and Retry Logic
- Return Final Response to Application

Must use Hexagonal Arcitecture
Must use Functional Options Pattern
Must use Golang Best Practices
