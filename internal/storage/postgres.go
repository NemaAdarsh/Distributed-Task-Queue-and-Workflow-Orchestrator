package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"flowctl/internal/core"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type PostgresStore struct {
	db     *sql.DB
	logger *logrus.Logger
}

func NewPostgresStore(connStr string, logger *logrus.Logger) (*PostgresStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &PostgresStore{
		db:     db,
		logger: logger,
	}

	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return store, nil
}

func (s *PostgresStore) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS workflows (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			status VARCHAR(20) NOT NULL,
			config JSONB NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
			started_at TIMESTAMP WITH TIME ZONE,
			completed_at TIMESTAMP WITH TIME ZONE
		)`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id VARCHAR(36) PRIMARY KEY,
			workflow_id VARCHAR(36) NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(100) NOT NULL,
			payload JSONB NOT NULL,
			status VARCHAR(20) NOT NULL,
			result JSONB,
			error TEXT,
			retry_count INTEGER NOT NULL DEFAULT 0,
			max_retries INTEGER NOT NULL DEFAULT 3,
			priority INTEGER NOT NULL DEFAULT 1,
			dependencies JSONB NOT NULL DEFAULT '[]',
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
			started_at TIMESTAMP WITH TIME ZONE,
			completed_at TIMESTAMP WITH TIME ZONE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_workflow_id ON tasks(workflow_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_type ON tasks(type)`,
		`CREATE INDEX IF NOT EXISTS idx_workflows_status ON workflows(status)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	return nil
}

func (s *PostgresStore) CreateWorkflow(workflow *core.Workflow) error {
	configJSON, err := json.Marshal(workflow.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	query := `
		INSERT INTO workflows (id, name, description, status, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = s.db.Exec(query,
		workflow.ID,
		workflow.Name,
		workflow.Description,
		workflow.Status,
		configJSON,
		workflow.CreatedAt,
		workflow.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}

	s.logger.Infof("Created workflow: %s", workflow.ID)
	return nil
}

func (s *PostgresStore) GetWorkflow(id string) (*core.Workflow, error) {
	query := `
		SELECT id, name, description, status, config, created_at, updated_at, started_at, completed_at
		FROM workflows WHERE id = $1
	`

	row := s.db.QueryRow(query, id)

	var workflow core.Workflow
	var configJSON []byte
	var startedAt, completedAt sql.NullTime

	err := row.Scan(
		&workflow.ID,
		&workflow.Name,
		&workflow.Description,
		&workflow.Status,
		&configJSON,
		&workflow.CreatedAt,
		&workflow.UpdatedAt,
		&startedAt,
		&completedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workflow not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	if err := json.Unmarshal(configJSON, &workflow.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if startedAt.Valid {
		workflow.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		workflow.CompletedAt = &completedAt.Time
	}

	tasks, err := s.GetTasksByWorkflow(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	workflow.Tasks = tasks
	return &workflow, nil
}

func (s *PostgresStore) UpdateWorkflowStatus(id string, status core.WorkflowStatus) error {
	now := time.Now()
	var query string
	var args []interface{}

	switch status {
	case core.WorkflowStatusRunning:
		query = `UPDATE workflows SET status = $1, started_at = $2, updated_at = $3 WHERE id = $4`
		args = []interface{}{status, now, now, id}
	case core.WorkflowStatusCompleted, core.WorkflowStatusFailed, core.WorkflowStatusCancelled:
		query = `UPDATE workflows SET status = $1, completed_at = $2, updated_at = $3 WHERE id = $4`
		args = []interface{}{status, now, now, id}
	default:
		query = `UPDATE workflows SET status = $1, updated_at = $2 WHERE id = $3`
		args = []interface{}{status, now, id}
	}

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update workflow status: %w", err)
	}

	s.logger.Infof("Updated workflow %s status to %s", id, status)
	return nil
}

func (s *PostgresStore) CreateTask(task *core.Task) error {
	payloadJSON, err := json.Marshal(task.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	dependenciesJSON, err := json.Marshal(task.Dependencies)
	if err != nil {
		return fmt.Errorf("failed to marshal dependencies: %w", err)
	}

	query := `
		INSERT INTO tasks (id, workflow_id, name, type, payload, status, retry_count, max_retries, priority, dependencies, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err = s.db.Exec(query,
		task.ID,
		task.WorkflowID,
		task.Name,
		task.Type,
		payloadJSON,
		task.Status,
		task.RetryCount,
		task.MaxRetries,
		task.Priority,
		dependenciesJSON,
		task.CreatedAt,
		task.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	s.logger.Infof("Created task: %s", task.ID)
	return nil
}

func (s *PostgresStore) GetTask(id string) (*core.Task, error) {
	query := `
		SELECT id, workflow_id, name, type, payload, status, result, error, retry_count, max_retries, priority, dependencies, created_at, updated_at, started_at, completed_at
		FROM tasks WHERE id = $1
	`

	row := s.db.QueryRow(query, id)
	return s.scanTask(row)
}

func (s *PostgresStore) GetTasksByWorkflow(workflowID string) ([]core.Task, error) {
	query := `
		SELECT id, workflow_id, name, type, payload, status, result, error, retry_count, max_retries, priority, dependencies, created_at, updated_at, started_at, completed_at
		FROM tasks WHERE workflow_id = $1 ORDER BY created_at
	`

	rows, err := s.db.Query(query, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []core.Task
	for rows.Next() {
		task, err := s.scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}

	return tasks, nil
}

func (s *PostgresStore) UpdateTaskStatus(id string, status core.TaskStatus, result map[string]interface{}, errorMsg string) error {
	now := time.Now()
	
	var resultJSON []byte
	if result != nil {
		var err error
		resultJSON, err = json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
	}

	var query string
	var args []interface{}

	switch status {
	case core.TaskStatusRunning:
		query = `UPDATE tasks SET status = $1, started_at = $2, updated_at = $3 WHERE id = $4`
		args = []interface{}{status, now, now, id}
	case core.TaskStatusCompleted:
		query = `UPDATE tasks SET status = $1, result = $2, completed_at = $3, updated_at = $4 WHERE id = $5`
		args = []interface{}{status, resultJSON, now, now, id}
	case core.TaskStatusFailed:
		query = `UPDATE tasks SET status = $1, error = $2, completed_at = $3, updated_at = $4 WHERE id = $5`
		args = []interface{}{status, errorMsg, now, now, id}
	case core.TaskStatusRetrying:
		query = `UPDATE tasks SET status = $1, retry_count = retry_count + 1, updated_at = $2 WHERE id = $3`
		args = []interface{}{status, now, id}
	default:
		query = `UPDATE tasks SET status = $1, updated_at = $2 WHERE id = $3`
		args = []interface{}{status, now, id}
	}

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	s.logger.Infof("Updated task %s status to %s", id, status)
	return nil
}

func (s *PostgresStore) GetPendingTasks() ([]core.Task, error) {
	query := `
		SELECT id, workflow_id, name, type, payload, status, result, error, retry_count, max_retries, priority, dependencies, created_at, updated_at, started_at, completed_at
		FROM tasks WHERE status IN ('pending', 'retrying') ORDER BY priority DESC, created_at ASC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending tasks: %w", err)
	}
	defer rows.Close()

	var tasks []core.Task
	for rows.Next() {
		task, err := s.scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}

	return tasks, nil
}

func (s *PostgresStore) scanTask(scanner interface {
	Scan(dest ...interface{}) error
}) (*core.Task, error) {
	var task core.Task
	var payloadJSON, resultJSON, dependenciesJSON []byte
	var result sql.NullString
	var errorMsg sql.NullString
	var startedAt, completedAt sql.NullTime

	err := scanner.Scan(
		&task.ID,
		&task.WorkflowID,
		&task.Name,
		&task.Type,
		&payloadJSON,
		&task.Status,
		&resultJSON,
		&errorMsg,
		&task.RetryCount,
		&task.MaxRetries,
		&task.Priority,
		&dependenciesJSON,
		&task.CreatedAt,
		&task.UpdatedAt,
		&startedAt,
		&completedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan task: %w", err)
	}

	if err := json.Unmarshal(payloadJSON, &task.Payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if err := json.Unmarshal(dependenciesJSON, &task.Dependencies); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dependencies: %w", err)
	}

	if resultJSON != nil {
		if err := json.Unmarshal(resultJSON, &task.Result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	if errorMsg.Valid {
		task.Error = errorMsg.String
	}

	if startedAt.Valid {
		task.StartedAt = &startedAt.Time
	}

	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}

	return &task, nil
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}
