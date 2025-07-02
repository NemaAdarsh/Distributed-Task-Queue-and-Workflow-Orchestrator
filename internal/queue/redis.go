package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"flowctl/internal/core"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type RedisQueue struct {
	client *redis.Client
	logger *logrus.Logger
}

func NewRedisQueue(addr, password string, db int, logger *logrus.Logger) (*RedisQueue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisQueue{
		client: client,
		logger: logger,
	}, nil
}

func (q *RedisQueue) EnqueueTask(ctx context.Context, task *core.Task) error {
	taskJSON, err := task.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize task: %w", err)
	}

	queueKey := fmt.Sprintf("queue:%s", task.Type)
	
	err = q.client.LPush(ctx, queueKey, taskJSON).Err()
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	q.logger.Infof("Enqueued task %s to queue %s", task.ID, queueKey)
	return nil
}

func (q *RedisQueue) DequeueTask(ctx context.Context, taskType string, timeout time.Duration) (*core.Task, error) {
	queueKey := fmt.Sprintf("queue:%s", taskType)
	processingKey := fmt.Sprintf("processing:%s", taskType)

	result, err := q.client.BRPopLPush(ctx, queueKey, processingKey, timeout).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to dequeue task: %w", err)
	}

	task, err := core.TaskFromJSON([]byte(result))
	if err != nil {
		q.client.LRem(ctx, processingKey, 1, result)
		return nil, fmt.Errorf("failed to deserialize task: %w", err)
	}

	q.logger.Infof("Dequeued task %s from queue %s", task.ID, queueKey)
	return task, nil
}

func (q *RedisQueue) AckTask(ctx context.Context, task *core.Task) error {
	processingKey := fmt.Sprintf("processing:%s", task.Type)
	
	taskJSON, err := task.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize task: %w", err)
	}

	err = q.client.LRem(ctx, processingKey, 1, string(taskJSON)).Err()
	if err != nil {
		return fmt.Errorf("failed to acknowledge task: %w", err)
	}

	q.logger.Infof("Acknowledged task %s", task.ID)
	return nil
}

func (q *RedisQueue) NackTask(ctx context.Context, task *core.Task) error {
	processingKey := fmt.Sprintf("processing:%s", task.Type)
	retryKey := fmt.Sprintf("retry:%s", task.Type)
	
	taskJSON, err := task.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize task: %w", err)
	}

	pipe := q.client.Pipeline()
	pipe.LRem(ctx, processingKey, 1, string(taskJSON))
	
	if task.RetryCount < task.MaxRetries {
		retryAt := time.Now().Add(q.calculateBackoff(task.RetryCount))
		pipe.ZAdd(ctx, retryKey, &redis.Z{
			Score:  float64(retryAt.Unix()),
			Member: string(taskJSON),
		})
	} else {
		deadLetterKey := fmt.Sprintf("dead_letter:%s", task.Type)
		pipe.LPush(ctx, deadLetterKey, string(taskJSON))
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to nack task: %w", err)
	}

	q.logger.Infof("Nacked task %s (retry count: %d)", task.ID, task.RetryCount)
	return nil
}

func (q *RedisQueue) ProcessRetries(ctx context.Context, taskType string) error {
	retryKey := fmt.Sprintf("retry:%s", taskType)
	queueKey := fmt.Sprintf("queue:%s", taskType)
	
	now := float64(time.Now().Unix())
	
	result, err := q.client.ZRangeByScore(ctx, retryKey, &redis.ZRangeBy{
		Min:   "0",
		Max:   fmt.Sprintf("%f", now),
		Count: 100,
	}).Result()
	
	if err != nil {
		return fmt.Errorf("failed to get retry tasks: %w", err)
	}

	for _, taskJSON := range result {
		task, err := core.TaskFromJSON([]byte(taskJSON))
		if err != nil {
			q.logger.Errorf("Failed to deserialize retry task: %v", err)
			continue
		}

		pipe := q.client.Pipeline()
		pipe.ZRem(ctx, retryKey, taskJSON)
		pipe.LPush(ctx, queueKey, taskJSON)
		
		_, err = pipe.Exec(ctx)
		if err != nil {
			q.logger.Errorf("Failed to requeue retry task %s: %v", task.ID, err)
			continue
		}

		q.logger.Infof("Requeued retry task %s", task.ID)
	}

	return nil
}

func (q *RedisQueue) GetQueueStats(ctx context.Context, taskType string) (map[string]int64, error) {
	queueKey := fmt.Sprintf("queue:%s", taskType)
	processingKey := fmt.Sprintf("processing:%s", taskType)
	retryKey := fmt.Sprintf("retry:%s", taskType)
	deadLetterKey := fmt.Sprintf("dead_letter:%s", taskType)

	pipe := q.client.Pipeline()
	queueLen := pipe.LLen(ctx, queueKey)
	processingLen := pipe.LLen(ctx, processingKey)
	retryLen := pipe.ZCard(ctx, retryKey)
	deadLetterLen := pipe.LLen(ctx, deadLetterKey)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue stats: %w", err)
	}

	return map[string]int64{
		"pending":     queueLen.Val(),
		"processing":  processingLen.Val(),
		"retry":       retryLen.Val(),
		"dead_letter": deadLetterLen.Val(),
	}, nil
}

func (q *RedisQueue) RegisterWorker(ctx context.Context, workerID, address string, taskTypes []string) error {
	workerKey := fmt.Sprintf("worker:%s", workerID)
	
	workerInfo := core.WorkerInfo{
		ID:            workerID,
		Address:       address,
		TaskTypes:     taskTypes,
		Status:        "active",
		LastHeartbeat: time.Now(),
		CurrentTasks:  []string{},
	}

	workerJSON, err := json.Marshal(workerInfo)
	if err != nil {
		return fmt.Errorf("failed to serialize worker info: %w", err)
	}

	err = q.client.Set(ctx, workerKey, workerJSON, time.Minute*5).Err()
	if err != nil {
		return fmt.Errorf("failed to register worker: %w", err)
	}

	for _, taskType := range taskTypes {
		workerSetKey := fmt.Sprintf("workers:%s", taskType)
		err = q.client.SAdd(ctx, workerSetKey, workerID).Err()
		if err != nil {
			q.logger.Errorf("Failed to add worker %s to task type %s: %v", workerID, taskType, err)
		}
	}

	q.logger.Infof("Registered worker %s for task types %v", workerID, taskTypes)
	return nil
}

func (q *RedisQueue) UpdateWorkerHeartbeat(ctx context.Context, workerID string) error {
	workerKey := fmt.Sprintf("worker:%s", workerID)
	
	workerJSON, err := q.client.Get(ctx, workerKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get worker info: %w", err)
	}

	var workerInfo core.WorkerInfo
	if err := json.Unmarshal([]byte(workerJSON), &workerInfo); err != nil {
		return fmt.Errorf("failed to unmarshal worker info: %w", err)
	}

	workerInfo.LastHeartbeat = time.Now()

	updatedJSON, err := json.Marshal(workerInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal updated worker info: %w", err)
	}

	err = q.client.Set(ctx, workerKey, updatedJSON, time.Minute*5).Err()
	if err != nil {
		return fmt.Errorf("failed to update worker heartbeat: %w", err)
	}

	return nil
}

func (q *RedisQueue) GetActiveWorkers(ctx context.Context, taskType string) ([]core.WorkerInfo, error) {
	workerSetKey := fmt.Sprintf("workers:%s", taskType)
	
	workerIDs, err := q.client.SMembers(ctx, workerSetKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get worker IDs: %w", err)
	}

	var workers []core.WorkerInfo
	for _, workerID := range workerIDs {
		workerKey := fmt.Sprintf("worker:%s", workerID)
		workerJSON, err := q.client.Get(ctx, workerKey).Result()
		if err != nil {
			if err == redis.Nil {
				q.client.SRem(ctx, workerSetKey, workerID)
				continue
			}
			q.logger.Errorf("Failed to get worker %s info: %v", workerID, err)
			continue
		}

		var workerInfo core.WorkerInfo
		if err := json.Unmarshal([]byte(workerJSON), &workerInfo); err != nil {
			q.logger.Errorf("Failed to unmarshal worker %s info: %v", workerID, err)
			continue
		}

		if time.Since(workerInfo.LastHeartbeat) > time.Minute*2 {
			q.client.SRem(ctx, workerSetKey, workerID)
			q.client.Del(ctx, workerKey)
			continue
		}

		workers = append(workers, workerInfo)
	}

	return workers, nil
}

func (q *RedisQueue) calculateBackoff(retryCount int) time.Duration {
	base := time.Second * 2
	backoff := base * time.Duration(1<<uint(retryCount))
	if backoff > time.Minute*5 {
		return time.Minute * 5
	}
	return backoff
}

func (q *RedisQueue) Close() error {
	return q.client.Close()
}
