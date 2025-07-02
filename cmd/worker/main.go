package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"flowctl/internal/core"
	"flowctl/internal/queue"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Worker struct {
	id           string
	address      string
	taskTypes    []string
	queue        *queue.RedisQueue
	logger       *logrus.Logger
	stopCh       chan struct{}
	schedulerURL string
}

func NewWorker(address string, taskTypes []string, redisQueue *queue.RedisQueue, schedulerURL string, logger *logrus.Logger) *Worker {
	return &Worker{
		id:           uuid.New().String(),
		address:      address,
		taskTypes:    taskTypes,
		queue:        redisQueue,
		logger:       logger,
		stopCh:       make(chan struct{}),
		schedulerURL: schedulerURL,
	}
}

func (w *Worker) Start(ctx context.Context) {
	w.logger.Infof("Starting worker %s on %s for task types %v", w.id, w.address, w.taskTypes)

	if err := w.queue.RegisterWorker(ctx, w.id, w.address, w.taskTypes); err != nil {
		w.logger.Errorf("Failed to register worker: %v", err)
		return
	}

	go w.heartbeat(ctx)

	for _, taskType := range w.taskTypes {
		go w.processTaskType(ctx, taskType)
	}

	<-w.stopCh
	w.logger.Info("Worker stopped")
}

func (w *Worker) Stop() {
	close(w.stopCh)
}

func (w *Worker) heartbeat(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case <-ticker.C:
			if err := w.queue.UpdateWorkerHeartbeat(ctx, w.id); err != nil {
				w.logger.Errorf("Failed to update heartbeat: %v", err)
			}
		}
	}
}

func (w *Worker) processTaskType(ctx context.Context, taskType string) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		default:
			task, err := w.queue.DequeueTask(ctx, taskType, time.Second*30)
			if err != nil {
				w.logger.Errorf("Failed to dequeue task: %v", err)
				time.Sleep(time.Second * 5)
				continue
			}

			if task == nil {
				continue
			}

			w.executeTask(ctx, task)
		}
	}
}

func (w *Worker) executeTask(ctx context.Context, task *core.Task) {
	w.logger.Infof("Executing task %s of type %s", task.ID, task.Type)

	w.notifyTaskStatus(task.ID, "running", nil, "")

	result, err := w.runTask(task)
	if err != nil {
		w.logger.Errorf("Task %s failed: %v", task.ID, err)
		
		if task.RetryCount < task.MaxRetries {
			w.queue.NackTask(ctx, task)
			w.notifyTaskStatus(task.ID, "retrying", nil, err.Error())
		} else {
			w.queue.AckTask(ctx, task)
			w.notifyTaskStatus(task.ID, "failed", nil, err.Error())
		}
		return
	}

	w.queue.AckTask(ctx, task)
	w.notifyTaskStatus(task.ID, "completed", result, "")
	w.logger.Infof("Task %s completed successfully", task.ID)
}

func (w *Worker) runTask(task *core.Task) (map[string]interface{}, error) {
	switch task.Type {
	case "etl":
		return w.runETLTask(task)
	case "ml_training":
		return w.runMLTrainingTask(task)
	case "ci":
		return w.runCITask(task)
	case "generic":
		return w.runGenericTask(task)
	default:
		return nil, fmt.Errorf("unknown task type: %s", task.Type)
	}
}

func (w *Worker) runETLTask(task *core.Task) (map[string]interface{}, error) {
	sourceURL, ok := task.Payload["source_url"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid source_url")
	}

	targetURL, ok := task.Payload["target_url"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid target_url")
	}

	w.logger.Infof("Processing ETL task: %s -> %s", sourceURL, targetURL)
	
	time.Sleep(time.Second * 5)

	return map[string]interface{}{
		"records_processed": 1000,
		"processing_time":   "5s",
		"source":           sourceURL,
		"target":           targetURL,
	}, nil
}

func (w *Worker) runMLTrainingTask(task *core.Task) (map[string]interface{}, error) {
	modelName, ok := task.Payload["model_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid model_name")
	}

	datasetURL, ok := task.Payload["dataset_url"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid dataset_url")
	}

	w.logger.Infof("Training ML model: %s with dataset: %s", modelName, datasetURL)
	
	time.Sleep(time.Second * 10)

	return map[string]interface{}{
		"model_name":      modelName,
		"accuracy":        0.95,
		"training_time":   "10s",
		"model_size_mb":   25.6,
	}, nil
}

func (w *Worker) runCITask(task *core.Task) (map[string]interface{}, error) {
	repoURL, ok := task.Payload["repo_url"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid repo_url")
	}

	command, ok := task.Payload["command"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid command")
	}

	w.logger.Infof("Running CI task: %s on %s", command, repoURL)
	
	time.Sleep(time.Second * 8)

	return map[string]interface{}{
		"repo_url":     repoURL,
		"command":      command,
		"exit_code":    0,
		"build_time":   "8s",
		"tests_passed": 42,
		"tests_failed": 0,
	}, nil
}

func (w *Worker) runGenericTask(task *core.Task) (map[string]interface{}, error) {
	command, ok := task.Payload["command"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid command")
	}

	w.logger.Infof("Running generic task: %s", command)
	
	sleepDuration := time.Second * 3
	if duration, ok := task.Payload["sleep_duration"].(float64); ok {
		sleepDuration = time.Duration(duration) * time.Second
	}

	time.Sleep(sleepDuration)

	return map[string]interface{}{
		"command":      command,
		"exit_code":    0,
		"output":       "Task completed successfully",
		"duration":     sleepDuration.String(),
	}, nil
}

func (w *Worker) notifyTaskStatus(taskID, status string, result map[string]interface{}, errorMsg string) {
	if w.schedulerURL == "" {
		return
	}

	payload := map[string]interface{}{
		"task_id": taskID,
		"status":  status,
		"result":  result,
		"error":   errorMsg,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		w.logger.Errorf("Failed to marshal task status: %v", err)
		return
	}

	url := fmt.Sprintf("%s/api/v1/tasks/%s/status", w.schedulerURL, taskID)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		w.logger.Errorf("Failed to notify task status: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		w.logger.Errorf("Failed to notify task status, status code: %d", resp.StatusCode)
	}
}

func main() {
	var (
		redisAddr    = flag.String("redis", "localhost:6379", "Redis address")
		redisPass    = flag.String("redis-pass", "", "Redis password")
		redisDB      = flag.Int("redis-db", 0, "Redis database")
		workerAddr   = flag.String("addr", "localhost:9000", "Worker address")
		schedulerURL = flag.String("scheduler", "http://localhost:8080", "Scheduler URL")
		taskTypes    = flag.String("types", "generic", "Comma-separated task types")
	)
	flag.Parse()

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	redisQueue, err := queue.NewRedisQueue(*redisAddr, *redisPass, *redisDB, logger)
	if err != nil {
		logger.Fatalf("Failed to create Redis queue: %v", err)
	}
	defer redisQueue.Close()

	var types []string
	if *taskTypes != "" {
		types = []string{*taskTypes}
	} else {
		types = []string{"generic"}
	}

	worker := NewWorker(*workerAddr, types, redisQueue, *schedulerURL, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go worker.Start(ctx)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	logger.Info("Received shutdown signal")

	worker.Stop()
	cancel()
}
