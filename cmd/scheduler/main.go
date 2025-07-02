package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"flowctl/internal/api"
	"flowctl/internal/core"
	"flowctl/internal/queue"
	"flowctl/internal/storage"

	"github.com/sirupsen/logrus"
)

func main() {
	var (
		postgresURL = flag.String("postgres", "postgres://user:password@localhost/flowctl?sslmode=disable", "PostgreSQL connection string")
		redisAddr   = flag.String("redis", "localhost:6379", "Redis address")
		redisPass   = flag.String("redis-pass", "", "Redis password")
		redisDB     = flag.Int("redis-db", 0, "Redis database")
		apiAddr     = flag.String("api", ":8080", "API server address")
	)
	flag.Parse()

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store, err := storage.NewPostgresStore(*postgresURL, logger)
	if err != nil {
		logger.Fatalf("Failed to create PostgreSQL store: %v", err)
	}
	defer store.Close()

	redisQueue, err := queue.NewRedisQueue(*redisAddr, *redisPass, *redisDB, logger)
	if err != nil {
		logger.Fatalf("Failed to create Redis queue: %v", err)
	}
	defer redisQueue.Close()

	scheduler := core.NewScheduler(store, redisQueue, logger)
	server := api.NewServer(scheduler, logger)

	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		scheduler.Start(ctx)
	}()

	go func() {
		defer wg.Done()
		if err := server.Start(*apiAddr); err != nil {
			logger.Errorf("API server failed: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	logger.Info("Received shutdown signal")

	cancel()
	scheduler.Stop()

	wg.Wait()
	logger.Info("Scheduler stopped")
}
