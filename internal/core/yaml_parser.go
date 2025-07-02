package core

import (
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v3"
)

type WorkflowSpec struct {
	Name        string              `yaml:"name"`
	Description string              `yaml:"description"`
	Config      WorkflowConfigSpec  `yaml:"config,omitempty"`
	Tasks       []TaskSpec          `yaml:"tasks"`
}

type WorkflowConfigSpec struct {
	MaxConcurrency int    `yaml:"max_concurrency,omitempty"`
	Timeout        string `yaml:"timeout,omitempty"`
	RetryPolicy    RetryPolicySpec `yaml:"retry_policy,omitempty"`
}

type RetryPolicySpec struct {
	MaxAttempts   int    `yaml:"max_attempts,omitempty"`
	InitialDelay  string `yaml:"initial_delay,omitempty"`
	MaxDelay      string `yaml:"max_delay,omitempty"`
	BackoffFactor float64 `yaml:"backoff_factor,omitempty"`
}

type TaskSpec struct {
	Name         string                 `yaml:"name"`
	Type         string                 `yaml:"type"`
	Payload      map[string]interface{} `yaml:"payload,omitempty"`
	MaxRetries   int                    `yaml:"max_retries,omitempty"`
	Priority     int                    `yaml:"priority,omitempty"`
	Dependencies []string               `yaml:"depends_on,omitempty"`
}

func ParseWorkflowFromYAML(filename string) (*Workflow, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	return ParseWorkflowFromYAMLBytes(data)
}

func ParseWorkflowFromYAMLBytes(data []byte) (*Workflow, error) {
	var spec WorkflowSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return convertSpecToWorkflow(&spec)
}

func convertSpecToWorkflow(spec *WorkflowSpec) (*Workflow, error) {
	workflow := NewWorkflow(spec.Name, spec.Description)

	if spec.Config.MaxConcurrency > 0 {
		workflow.Config.MaxConcurrency = spec.Config.MaxConcurrency
	}

	if spec.Config.Timeout != "" {
		timeout, err := time.ParseDuration(spec.Config.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout duration: %w", err)
		}
		workflow.Config.Timeout = timeout
	}

	if spec.Config.RetryPolicy.MaxAttempts > 0 {
		workflow.Config.RetryPolicy.MaxAttempts = spec.Config.RetryPolicy.MaxAttempts
	}

	if spec.Config.RetryPolicy.InitialDelay != "" {
		delay, err := time.ParseDuration(spec.Config.RetryPolicy.InitialDelay)
		if err != nil {
			return nil, fmt.Errorf("invalid initial delay: %w", err)
		}
		workflow.Config.RetryPolicy.InitialDelay = delay
	}

	if spec.Config.RetryPolicy.MaxDelay != "" {
		delay, err := time.ParseDuration(spec.Config.RetryPolicy.MaxDelay)
		if err != nil {
			return nil, fmt.Errorf("invalid max delay: %w", err)
		}
		workflow.Config.RetryPolicy.MaxDelay = delay
	}

	if spec.Config.RetryPolicy.BackoffFactor > 0 {
		workflow.Config.RetryPolicy.BackoffFactor = spec.Config.RetryPolicy.BackoffFactor
	}

	taskMap := make(map[string]*Task)
	
	for _, taskSpec := range spec.Tasks {
		task := NewTask(workflow.ID, taskSpec.Name, taskSpec.Type, taskSpec.Payload)
		
		if taskSpec.MaxRetries > 0 {
			task.MaxRetries = taskSpec.MaxRetries
		}
		
		if taskSpec.Priority > 0 {
			task.Priority = taskSpec.Priority
		}

		task.Dependencies = taskSpec.Dependencies
		
		taskMap[taskSpec.Name] = task
		workflow.Tasks = append(workflow.Tasks, *task)
	}

	if err := validateWorkflowDependencies(workflow.Tasks); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	return workflow, nil
}

func validateWorkflowDependencies(tasks []Task) error {
	taskNames := make(map[string]bool)
	for _, task := range tasks {
		taskNames[task.Name] = true
	}

	for _, task := range tasks {
		for _, dep := range task.Dependencies {
			if !taskNames[dep] {
				return fmt.Errorf("task %s depends on non-existent task %s", task.Name, dep)
			}
		}
	}

	if hasCycle(tasks) {
		return fmt.Errorf("workflow contains circular dependencies")
	}

	return nil
}

func hasCycle(tasks []Task) bool {
	taskDeps := make(map[string][]string)
	for _, task := range tasks {
		taskDeps[task.Name] = task.Dependencies
	}

	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycleUtil func(string) bool
	hasCycleUtil = func(taskName string) bool {
		visited[taskName] = true
		recStack[taskName] = true

		for _, dep := range taskDeps[taskName] {
			if !visited[dep] && hasCycleUtil(dep) {
				return true
			} else if recStack[dep] {
				return true
			}
		}

		recStack[taskName] = false
		return false
	}

	for _, task := range tasks {
		if !visited[task.Name] && hasCycleUtil(task.Name) {
			return true
		}
	}

	return false
}
