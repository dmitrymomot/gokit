package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/dmitrymomot/gokit/queue"
	queueredis "github.com/dmitrymomot/gokit/queue/redis"
	"github.com/redis/go-redis/v9"
)

// EmailPayload represents the data needed to send an email
type EmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// NotificationPayload represents the data for a notification
type NotificationPayload struct {
	UserID  string `json:"user_id"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

// LoggingMiddleware is a middleware that logs job processing
func LoggingMiddleware(next queue.Handler) queue.Handler {
	return func(ctx context.Context, job *queue.Job) error {
		// Log job start
		var payloadMap map[string]interface{}
		if err := json.Unmarshal(job.Payload, &payloadMap); err != nil {
			payloadMap = map[string]interface{}{"error": "failed to unmarshal payload"}
		}
		
		log.Printf("Processing job %s (task: %s) with payload: %+v", job.ID, job.TaskName, payloadMap)
		
		startTime := time.Now()
		
		// Call the next handler
		err := next(ctx, job)
		
		// Log job completion
		duration := time.Since(startTime)
		if err != nil {
			log.Printf("Job %s failed after %v: %v", job.ID, duration, err)
		} else {
			log.Printf("Job %s completed successfully in %v", job.ID, duration)
		}
		
		return err
	}
}

// TimingMiddleware is a middleware that measures the execution time of jobs
func TimingMiddleware(next queue.Handler) queue.Handler {
	return func(ctx context.Context, job *queue.Job) error {
		startTime := time.Now()
		err := next(ctx, job)
		log.Printf("Job %s (%s) execution time: %v", job.ID, job.TaskName, time.Since(startTime))
		return err
	}
}

func main() {
	// Setup miniredis (In-memory Redis server for testing)
	mr, err := miniredis.Run()
	if err != nil {
		log.Fatalf("Failed to start miniredis: %v", err)
	}
	defer mr.Close()

	// Create a Redis client connected to miniredis
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer redisClient.Close()

	// Create Redis storage for the queue
	storage := queueredis.New(redisClient)

	// Create the queue with middleware
	q := queue.New(storage, 
		queue.WithConcurrency(5),
		queue.WithMiddleware(LoggingMiddleware, TimingMiddleware),
	)

	// Register handlers
	if err := q.AddHandler("send_email", handleEmail); err != nil {
		log.Fatalf("Failed to register email handler: %v", err)
	}

	if err := q.AddHandler("send_notification", handleNotification); err != nil {
		log.Fatalf("Failed to register notification handler: %v", err)
	}

	if err := q.AddHandler("process_with_error", handleWithError); err != nil {
		log.Fatalf("Failed to register error handler: %v", err)
	}

	// Create context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the queue in a goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Starting queue processor...")
		if err := q.Run(ctx); err != nil && err != context.Canceled {
			log.Printf("Queue processor error: %v", err)
		}
		log.Println("Queue processor stopped")
	}()

	// Allow time for queue to start
	time.Sleep(100 * time.Millisecond)

	// Enqueue an immediate email job
	emailJob := EmailPayload{
		To:      "user@example.com",
		Subject: "Welcome to our service",
		Body:    "Thank you for registering!",
	}
	jobID, err := q.Enqueue(ctx, "send_email", emailJob)
	if err != nil {
		log.Fatalf("Failed to enqueue email job: %v", err)
	}
	log.Printf("Enqueued email job with ID: %s", jobID)

	// Enqueue a delayed notification job (run after 2 seconds)
	notificationJob := NotificationPayload{
		UserID:  "user123",
		Message: "You have a new message",
		Type:    "message",
	}
	delayJobID, err := q.EnqueueIn(ctx, "send_notification", notificationJob, 2*time.Second)
	if err != nil {
		log.Fatalf("Failed to enqueue delayed notification job: %v", err)
	}
	log.Printf("Enqueued delayed notification job with ID: %s", delayJobID)

	// Enqueue a job that will fail and be retried
	errorJobID, err := q.Enqueue(ctx, "process_with_error", map[string]string{"data": "this will fail"})
	if err != nil {
		log.Fatalf("Failed to enqueue error job: %v", err)
	}
	log.Printf("Enqueued error job with ID: %s", errorJobID)

	// Let the queue process the jobs (including the delayed one)
	log.Println("Waiting for jobs to be processed...")
	time.Sleep(5 * time.Second)

	// Gracefully stop the queue
	log.Println("Stopping queue...")
	ctxStop, cancelStop := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelStop()
	if err := q.Stop(ctxStop); err != nil {
		log.Printf("Error stopping queue: %v", err)
	} else {
		log.Println("Queue stopped successfully")
	}

	// Wait for queue processor to finish
	wg.Wait()
	log.Println("Example completed successfully")
}

// Handler for email jobs
func handleEmail(ctx context.Context, payload EmailPayload) error {
	log.Printf("Sending email to %s with subject: %s", payload.To, payload.Subject)
	// Simulate actual email sending
	time.Sleep(200 * time.Millisecond)
	return nil
}

// Handler for notification jobs
func handleNotification(ctx context.Context, payload NotificationPayload) error {
	log.Printf("Sending %s notification to user %s: %s", payload.Type, payload.UserID, payload.Message)
	// Simulate actual notification sending
	time.Sleep(150 * time.Millisecond)
	return nil
}

// Handler that will fail and trigger retries
func handleWithError(ctx context.Context, payload map[string]string) error {
	log.Printf("Processing job that will fail with payload: %v", payload)
	retryCount, ok := payload["retry_count"]
	if !ok {
		payload["retry_count"] = "0"
		retryCount = "0"
	}
	count, err := strconv.Atoi(retryCount)
	if err != nil {
		return err
	}
	if count >= 3 {
		log.Println("Job will succeed after 3 attempts")
		return nil
	}
	log.Printf("Job will fail (attempt %d)", count+1)
	payload["retry_count"] = strconv.Itoa(count + 1)
	return fmt.Errorf("simulated error to demonstrate retry mechanism")
}
