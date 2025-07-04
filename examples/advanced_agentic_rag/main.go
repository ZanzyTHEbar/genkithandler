package main

import (
	"context"
	"fmt"
	"log"

	genkit_agentic_rag "github.com/ZanzyTHEbar/genkit-agentic-rag"
	"github.com/ZanzyTHEbar/genkit-agentic-rag/plugin"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
)

func main() {
	ctx := context.Background()

	// Create advanced agentic RAG configuration with dotprompt support
	config := &plugin.AgenticRAGConfig{
		ModelName: "googleai/gemini-2.5-flash",
		Processing: plugin.ProcessingConfig{
			DefaultChunkSize:      800, // Smaller chunks for better precision
			DefaultMaxChunks:      25,  // More chunks for comprehensive analysis
			DefaultRecursiveDepth: 4,   // Deeper recursive analysis
			RespectSentences:      true,
		},
		KnowledgeGraph: plugin.KnowledgeGraphConfig{
			Enabled:                true,
			EntityTypes:            []string{"PERSON", "ORGANIZATION", "TECHNOLOGY", "CONCEPT", "EVENT", "LOCATION"},
			RelationTypes:          []string{"DEVELOPS", "USES", "FOUNDED", "LOCATED_IN", "WORKS_FOR", "INVENTED"},
			MinConfidenceThreshold: 0.8, // Higher confidence threshold for quality
		},
		FactVerification: plugin.FactVerificationConfig{
			Enabled:            true,
			RequireEvidence:    true,
			MinConfidenceScore: 0.75,
		},
		Prompts: plugin.PromptsConfig{
			Directory:                 "../../prompts", // Path to prompts from example directory
			RelevanceScoringPrompt:    "relevance_scoring",
			ResponseGenerationPrompt:  "response_generation",
			KnowledgeExtractionPrompt: "knowledge_extraction",
			FactVerificationPrompt:    "fact_verification",
			Variants: map[string]string{
				"relevance_scoring":   "strict",   // Use strict variant for higher precision
				"response_generation": "creative", // Use creative variant for engaging responses
			},
			CustomHelpers: true,
		},
	}

	// Initialize GenKit with Google AI plugin and prompts support
	g, err := genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.GoogleAI{}),
		genkit.WithPromptDir(config.Prompts.Directory),
	)
	if err != nil {
		log.Fatalf("Failed to initialize GenKit: %v", err)
	}

	// Store GenKit instance in config
	config.Genkit = g

	// Optional: Set a specific model instance if available
	model := genkit.LookupModel(g, "googleai", "gemini-2.5-flash")
	if model != nil {
		config.Model = model
	}

	// Initialize the enhanced agentic RAG plugin
	if err := genkit_agentic_rag.InitializeAgenticRAG(g, config); err != nil {
		log.Fatalf("Failed to initialize Agentic RAG: %v", err)
	}

	// Example 1: Basic Agentic RAG with knowledge graph
	fmt.Println("=== Example 1: Advanced Agentic RAG with Knowledge Graph ===")
	processor := genkit_agentic_rag.NewAgenticRAGProcessor(config)

	// Sample technical document
	request := plugin.AgenticRAGRequest{
		Query: "What are the key features and benefits of microservices architecture?",
		Documents: []string{
			`Microservices architecture is a design approach where applications are built as a collection of small, 
			independent services that communicate through well-defined APIs. Each service is responsible for a specific 
			business function and can be developed, deployed, and scaled independently. Key benefits include improved 
			scalability, technology diversity, faster deployment cycles, and better fault isolation. Companies like 
			Netflix, Amazon, and Google have successfully adopted microservices to handle massive scale and complexity. 
			However, microservices also introduce challenges such as increased complexity in service communication, 
			data consistency issues, and the need for sophisticated monitoring and orchestration tools like Kubernetes.`,
		},
		Options: plugin.AgenticRAGOptions{
			MaxChunks:              20,
			RecursiveDepth:         3,
			EnableKnowledgeGraph:   true,
			EnableFactVerification: true,
			Temperature:            0.3,
		},
	}

	response, err := processor.Process(ctx, request)
	if err != nil {
		log.Fatalf("Failed to process request: %v", err)
	}

	fmt.Printf("Answer: %s\n\n", response.Answer)
	fmt.Printf("Knowledge Graph Entities: %d\n", len(response.KnowledgeGraph.Entities))
	fmt.Printf("Knowledge Graph Relations: %d\n", len(response.KnowledgeGraph.Relations))

	if response.FactVerification != nil {
		fmt.Printf("Fact Verification Status: %s\n", response.FactVerification.Overall)
		fmt.Printf("Claims Verified: %d\n", len(response.FactVerification.Claims))
	}

	fmt.Printf("Processing Time: %v\n", response.ProcessingMetadata.ProcessingTime)
	fmt.Printf("Tokens Used: %d\n\n", response.ProcessingMetadata.TokensUsed)

	// Example 2: Complex technical analysis
	fmt.Println("=== Example 2: Complex Technical Analysis ===")
	complexRequest := plugin.AgenticRAGRequest{
		Query: "How do distributed databases handle consistency and partition tolerance in CAP theorem?",
		Documents: []string{
			`The CAP theorem, formulated by Eric Brewer, states that distributed data stores can only guarantee two 
			out of three properties: Consistency (all nodes see the same data simultaneously), Availability (the system 
			remains operational), and Partition tolerance (the system continues to function despite network failures). 
			This has profound implications for database design. Traditional ACID databases like PostgreSQL prioritize 
			consistency and availability but struggle with partition tolerance. NoSQL databases like Cassandra and DynamoDB 
			choose availability and partition tolerance, implementing eventual consistency. MongoDB offers configurable 
			consistency levels. Modern solutions like CockroachDB and Spanner attempt to provide strong consistency across 
			distributed systems using sophisticated consensus algorithms like Raft and Paxos.`,

			`Distributed consensus algorithms are crucial for maintaining consistency in distributed systems. The Raft 
			algorithm, developed by Diego Ongaro and John Ousterhout, provides a more understandable alternative to Paxos. 
			Raft divides consensus into leader election, log replication, and safety. In Raft, one server acts as leader 
			and handles all client requests, replicating entries to follower servers. The Byzantine Fault Tolerance (BFT) 
			algorithms handle more complex failure scenarios where nodes might behave maliciously. Practical Byzantine Fault 
			Tolerance (pBFT) and its variants are used in blockchain systems and critical distributed applications.`,
		},
		Options: plugin.AgenticRAGOptions{
			MaxChunks:              25,
			RecursiveDepth:         4,
			EnableKnowledgeGraph:   true,
			EnableFactVerification: true,
			Temperature:            0.2, // Lower temperature for technical accuracy
		},
	}

	complexResponse, err := processor.Process(ctx, complexRequest)
	if err != nil {
		log.Fatalf("Failed to process complex request: %v", err)
	}

	fmt.Printf("Answer: %s\n\n", complexResponse.Answer)

	// Display knowledge graph insights
	if complexResponse.KnowledgeGraph != nil {
		fmt.Println("Knowledge Graph Insights:")
		for _, entity := range complexResponse.KnowledgeGraph.Entities {
			if entity.Confidence > 0.9 {
				fmt.Printf("  - %s (%s): %.2f confidence\n", entity.Name, entity.Type, entity.Confidence)
			}
		}

		fmt.Println("Key Relations:")
		for _, relation := range complexResponse.KnowledgeGraph.Relations {
			if relation.Confidence > 0.8 {
				fmt.Printf("  - %s %s %s (%.2f confidence)\n",
					relation.Subject, relation.Predicate, relation.Object, relation.Confidence)
			}
		}
	}

	fmt.Printf("\nProcessing Metadata:\n")
	fmt.Printf("  Chunks Processed: %d\n", complexResponse.ProcessingMetadata.ChunksProcessed)
	fmt.Printf("  Recursive Levels: %d\n", complexResponse.ProcessingMetadata.RecursiveLevels)
	fmt.Printf("  Model Calls: %d\n", complexResponse.ProcessingMetadata.ModelCalls)
	fmt.Printf("  Processing Time: %v\n", complexResponse.ProcessingMetadata.ProcessingTime)

	fmt.Println("\n=== Advanced Agentic RAG Demo Complete ===")
}
