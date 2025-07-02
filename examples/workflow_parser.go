package main

import (
	"context"
	"log"

	"flowctl/internal/core"
)

func main() {
	workflow, err := core.ParseWorkflowFromYAML("examples/etl_pipeline.yaml")
	if err != nil {
		log.Fatalf("Failed to parse workflow: %v", err)
	}

	log.Printf("Loaded workflow: %s", workflow.Name)
	log.Printf("Description: %s", workflow.Description)
	log.Printf("Tasks: %d", len(workflow.Tasks))

	for _, task := range workflow.Tasks {
		log.Printf("  - %s (%s) [priority: %d, deps: %v]", 
			task.Name, task.Type, task.Priority, task.Dependencies)
	}
}
