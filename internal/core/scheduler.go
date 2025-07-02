package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"flowctl/internal/queue"
	"flowctl/internal/storage"

	"github.com/sirupsen/logrus"
)

type Scheduler struct {
	store    *storage.PostgresStore
	queue    *queue.RedisQueue
	logger   *logrus.Logger
	stopCh   chan struct{}
	wg       sync.WaitGroup
	interval time.Duration
}

func NewScheduler(store *storage.PostgresStore, queue *queue.RedisQueue, logger *logrus.Logger) *Scheduler {
	return &Scheduler{
		store:    store,
		queue:    queue,
		logger:   logger,
		stopCh:   make(chan struct{}),
		interval: time.Second * 10,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	s.logger.Info("Starting scheduler")
	
	s.wg.Add(3)
	go s.scheduleWorkflows(ctx)
	go s.processRetries(ctx)
	go s.monitorWorkflows(ctx)
}

func (s *Scheduler) Stop() {
	s.logger.Info("Stopping scheduler")
	close(s.stopCh)
	s.wg.Wait()
}

func (s *Scheduler) scheduleWorkflows(ctx context.Context) {
	defer s.wg.Done()
	
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			if err := s.schedulePendingTasks(ctx); err != nil {
				s.logger.Errorf("Failed to schedule pending tasks: %v", err)
			}
		}
	}
}

func (s *Scheduler) schedulePendingTasks(ctx context.Context) error {
	tasks, err := s.store.GetPendingTasks()
	if err != nil {
		return fmt.Errorf("failed to get pending tasks: %w", err)
	}

	workflowTasks := make(map[string][]Task)
	for _, task := range tasks {
		workflowTasks[task.WorkflowID] = append(workflowTasks[task.WorkflowID], task)
	}

	for workflowID, tasks := range workflowTasks {
		if err := s.scheduleWorkflowTasks(ctx, workflowID, tasks); err != nil {
			s.logger.Errorf("Failed to schedule tasks for workflow %s: %v", workflowID, err)
		}
	}

	return nil
}

func (s *Scheduler) scheduleWorkflowTasks(ctx context.Context, workflowID string, tasks []Task) error {
	workflow, err := s.store.GetWorkflow(workflowID)
	if err != nil {
		return fmt.Errorf("failed to get workflow: %w", err)
	}

	if workflow.Status != WorkflowStatusPending && workflow.Status != WorkflowStatusRunning {
		return nil
	}

	completedTasks := make(map[string]bool)
	for _, task := range workflow.Tasks {
		if task.Status == TaskStatusCompleted {
			completedTasks[task.ID] = true
		}
	}

	var tasksToSchedule []Task
	for _, task := range tasks {
		if task.CanExecute(completedTasks) {
			tasksToSchedule = append(tasksToSchedule, task)
		}
	}

	if len(tasksToSchedule) == 0 {
		return nil
	}

	if workflow.Status == WorkflowStatusPending {
		if err := s.store.UpdateWorkflowStatus(workflowID, WorkflowStatusRunning); err != nil {
			return fmt.Errorf("failed to update workflow status: %w", err)
		}
	}

	for _, task := range tasksToSchedule {
		if err := s.queue.EnqueueTask(ctx, &task); err != nil {
			s.logger.Errorf("Failed to enqueue task %s: %v", task.ID, err)
			continue
		}
		
		if err := s.store.UpdateTaskStatus(task.ID, TaskStatusPending, nil, ""); err != nil {
			s.logger.Errorf("Failed to update task status %s: %v", task.ID, err)
		}
	}

	s.logger.Infof("Scheduled %d tasks for workflow %s", len(tasksToSchedule), workflowID)
	return nil
}

func (s *Scheduler) processRetries(ctx context.Context) {
	defer s.wg.Done()
	
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			taskTypes := []string{"etl", "ml_training", "ci", "generic"}
			for _, taskType := range taskTypes {
				if err := s.queue.ProcessRetries(ctx, taskType); err != nil {
					s.logger.Errorf("Failed to process retries for task type %s: %v", taskType, err)
				}
			}
		}
	}
}

func (s *Scheduler) monitorWorkflows(ctx context.Context) {
	defer s.wg.Done()
	
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			if err := s.checkWorkflowCompletion(ctx); err != nil {
				s.logger.Errorf("Failed to check workflow completion: %v", err)
			}
		}
	}
}

func (s *Scheduler) checkWorkflowCompletion(ctx context.Context) error {
	return nil
}

func (s *Scheduler) SubmitWorkflow(ctx context.Context, workflow *Workflow) error {
	if err := s.store.CreateWorkflow(workflow); err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}

	for _, task := range workflow.Tasks {
		if err := s.store.CreateTask(&task); err != nil {
			s.logger.Errorf("Failed to create task %s: %v", task.ID, err)
		}
	}

	s.logger.Infof("Submitted workflow %s with %d tasks", workflow.ID, len(workflow.Tasks))
	return nil
}

func (s *Scheduler) CancelWorkflow(ctx context.Context, workflowID string) error {
	if err := s.store.UpdateWorkflowStatus(workflowID, WorkflowStatusCancelled); err != nil {
		return fmt.Errorf("failed to cancel workflow: %w", err)
	}

	s.logger.Infof("Cancelled workflow %s", workflowID)
	return nil
}

func (s *Scheduler) GetWorkflow(workflowID string) (*Workflow, error) {
	return s.store.GetWorkflow(workflowID)
}

func (s *Scheduler) GetTask(taskID string) (*Task, error) {
	return s.store.GetTask(taskID)
}

func (s *Scheduler) GetWorkflowTasks(workflowID string) ([]Task, error) {
	return s.store.GetTasksByWorkflow(workflowID)
}
