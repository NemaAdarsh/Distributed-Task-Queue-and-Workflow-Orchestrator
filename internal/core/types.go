package core

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusRetrying  TaskStatus = "retrying"
	TaskStatusCancelled TaskStatus = "cancelled"
)

type WorkflowStatus string

const (
	WorkflowStatusPending   WorkflowStatus = "pending"
	WorkflowStatusRunning   WorkflowStatus = "running"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusFailed    WorkflowStatus = "failed"
	WorkflowStatusCancelled WorkflowStatus = "cancelled"
)

type Task struct {
	ID          string                 `json:"id" db:"id"`
	WorkflowID  string                 `json:"workflow_id" db:"workflow_id"`
	Name        string                 `json:"name" db:"name"`
	Type        string                 `json:"type" db:"type"`
	Payload     map[string]interface{} `json:"payload" db:"payload"`
	Status      TaskStatus             `json:"status" db:"status"`
	Result      map[string]interface{} `json:"result,omitempty" db:"result"`
	Error       string                 `json:"error,omitempty" db:"error"`
	RetryCount  int                    `json:"retry_count" db:"retry_count"`
	MaxRetries  int                    `json:"max_retries" db:"max_retries"`
	Priority    int                    `json:"priority" db:"priority"`
	Dependencies []string              `json:"dependencies" db:"dependencies"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty" db:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
}

type Workflow struct {
	ID          string         `json:"id" db:"id"`
	Name        string         `json:"name" db:"name"`
	Description string         `json:"description" db:"description"`
	Status      WorkflowStatus `json:"status" db:"status"`
	Tasks       []Task         `json:"tasks"`
	Config      WorkflowConfig `json:"config" db:"config"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
	StartedAt   *time.Time     `json:"started_at,omitempty" db:"started_at"`
	CompletedAt *time.Time     `json:"completed_at,omitempty" db:"completed_at"`
}

type WorkflowConfig struct {
	MaxConcurrency int           `json:"max_concurrency" yaml:"max_concurrency"`
	Timeout        time.Duration `json:"timeout" yaml:"timeout"`
	RetryPolicy    RetryPolicy   `json:"retry_policy" yaml:"retry_policy"`
}

type RetryPolicy struct {
	MaxAttempts   int           `json:"max_attempts" yaml:"max_attempts"`
	InitialDelay  time.Duration `json:"initial_delay" yaml:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay" yaml:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor" yaml:"backoff_factor"`
}

type WorkerInfo struct {
	ID           string    `json:"id"`
	Address      string    `json:"address"`
	TaskTypes    []string  `json:"task_types"`
	Status       string    `json:"status"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	CurrentTasks []string  `json:"current_tasks"`
}

func NewTask(workflowID, name, taskType string, payload map[string]interface{}) *Task {
	return &Task{
		ID:           uuid.New().String(),
		WorkflowID:   workflowID,
		Name:         name,
		Type:         taskType,
		Payload:      payload,
		Status:       TaskStatusPending,
		RetryCount:   0,
		MaxRetries:   3,
		Priority:     1,
		Dependencies: []string{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func NewWorkflow(name, description string) *Workflow {
	return &Workflow{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Status:      WorkflowStatusPending,
		Tasks:       []Task{},
		Config: WorkflowConfig{
			MaxConcurrency: 10,
			Timeout:        time.Hour,
			RetryPolicy: RetryPolicy{
				MaxAttempts:   3,
				InitialDelay:  time.Second,
				MaxDelay:      time.Minute * 5,
				BackoffFactor: 2.0,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (t *Task) CanExecute(completedTasks map[string]bool) bool {
	for _, dep := range t.Dependencies {
		if !completedTasks[dep] {
			return false
		}
	}
	return t.Status == TaskStatusPending || t.Status == TaskStatusRetrying
}

func (t *Task) ToJSON() ([]byte, error) {
	return json.Marshal(t)
}

func TaskFromJSON(data []byte) (*Task, error) {
	var task Task
	err := json.Unmarshal(data, &task)
	return &task, err
}
