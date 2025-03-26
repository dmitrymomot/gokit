// Package redis provides a Redis-based implementation of the queue.Storage interface.
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dmitrymomot/gokit/queue"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	// Default key prefixes
	defaultPrefix      = "queue:"
	defaultJobsPrefix  = "jobs:"
	defaultPendingSet  = "pending"
	defaultProcessing  = "processing"
	defaultCompletedSet = "completed"
	defaultFailedSet   = "failed"
	defaultRetryingSet = "retrying"
)

// Storage implements the queue.Storage interface using Redis.
// It is thread-safe and can be used across multiple instances of the application
// for a distributed queue system.
type Storage struct {
	client         redis.UniversalClient
	prefix         string
	jobsPrefix     string
	pendingSet     string
	processingSet  string
	completedSet   string
	failedSet      string
	retryingSet    string
	lockTimeout    time.Duration
}

// Options configures the Redis storage.
type Options struct {
	// Prefix is the prefix used for all Redis keys.
	// Default: "queue:"
	Prefix string

	// LockTimeout is the duration after which a lock on a job is considered expired.
	// Default: 30s
	LockTimeout time.Duration
}

// DefaultOptions returns the default options for the Redis storage.
func DefaultOptions() Options {
	return Options{
		Prefix:      defaultPrefix,
		LockTimeout: 30 * time.Second,
	}
}

// New creates a new Redis storage using the provided Redis client.
// Options are optional and can be omitted to use the default values.
func New(client redis.UniversalClient, opts ...Options) *Storage {
	options := DefaultOptions()
	if len(opts) > 0 {
		options = opts[0]
	}

	// Make sure the prefix ends with ":"
	if options.Prefix != "" && options.Prefix[len(options.Prefix)-1] != ':' {
		options.Prefix += ":"
	}

	prefix := options.Prefix
	return &Storage{
		client:         client,
		prefix:         prefix,
		jobsPrefix:     prefix + defaultJobsPrefix,
		pendingSet:     prefix + defaultPendingSet,
		processingSet:  prefix + defaultProcessing,
		completedSet:   prefix + defaultCompletedSet,
		failedSet:      prefix + defaultFailedSet,
		retryingSet:    prefix + defaultRetryingSet,
		lockTimeout:    options.LockTimeout,
	}
}

// jobKey returns the Redis key for a job.
func (s *Storage) jobKey(id string) string {
	return s.jobsPrefix + id
}

// Ping checks if the Redis connection is available.
func (s *Storage) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

// Put stores a job in Redis.
func (s *Storage) Put(ctx context.Context, job *queue.Job) error {
	// Generate a new ID if not set
	if job.ID == "" {
		job.ID = uuid.New().String()
	}

	// Serialize the job
	jobData, err := json.Marshal(job)
	if err != nil {
		return err
	}

	// Store the job in a transaction
	pipe := s.client.Pipeline()
	
	// Store the job data
	pipe.Set(ctx, s.jobKey(job.ID), jobData, 0)
	
	// Add to the appropriate set based on status
	switch job.Status {
	case queue.JobStatusPending:
		// Add to pending set with score as the Unix timestamp for RunAt
		pipe.ZAdd(ctx, s.pendingSet, redis.Z{
			Score:  float64(job.RunAt.Unix()),
			Member: job.ID,
		})
	case queue.JobStatusProcessing:
		pipe.SAdd(ctx, s.processingSet, job.ID)
	case queue.JobStatusCompleted:
		pipe.SAdd(ctx, s.completedSet, job.ID)
	case queue.JobStatusFailed:
		pipe.SAdd(ctx, s.failedSet, job.ID)
	case queue.JobStatusRetrying:
		// Add to retrying set with score as the Unix timestamp for RunAt
		pipe.ZAdd(ctx, s.retryingSet, redis.Z{
			Score:  float64(job.RunAt.Unix()),
			Member: job.ID,
		})
	default:
		return queue.ErrUnknownJobStatus
	}

	// Execute the transaction
	_, err = pipe.Exec(ctx)
	return err
}

// Get retrieves a job by ID.
func (s *Storage) Get(ctx context.Context, id string) (*queue.Job, error) {
	data, err := s.client.Get(ctx, s.jobKey(id)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, queue.ErrJobNotFound
		}
		return nil, err
	}

	var job queue.Job
	if err := json.Unmarshal(data, &job); err != nil {
		return nil, err
	}

	return &job, nil
}

// Update updates a job in Redis.
func (s *Storage) Update(ctx context.Context, job *queue.Job) error {
	// Get current job to check if status has changed
	currentJob, err := s.Get(ctx, job.ID)
	if err != nil {
		return err
	}

	// Serialize the job
	jobData, err := json.Marshal(job)
	if err != nil {
		return err
	}

	// Update the job in a transaction
	pipe := s.client.Pipeline()
	
	// Update the job data
	pipe.Set(ctx, s.jobKey(job.ID), jobData, 0)
	
	// If status has changed, update the sets
	if currentJob.Status != job.Status {
		// Remove from the old status set
		switch currentJob.Status {
		case queue.JobStatusPending:
			pipe.ZRem(ctx, s.pendingSet, job.ID)
		case queue.JobStatusProcessing:
			pipe.SRem(ctx, s.processingSet, job.ID)
		case queue.JobStatusCompleted:
			pipe.SRem(ctx, s.completedSet, job.ID)
		case queue.JobStatusFailed:
			pipe.SRem(ctx, s.failedSet, job.ID)
		case queue.JobStatusRetrying:
			pipe.ZRem(ctx, s.retryingSet, job.ID)
		}

		// Add to the new status set
		switch job.Status {
		case queue.JobStatusPending:
			pipe.ZAdd(ctx, s.pendingSet, redis.Z{
				Score:  float64(job.RunAt.Unix()),
				Member: job.ID,
			})
		case queue.JobStatusProcessing:
			pipe.SAdd(ctx, s.processingSet, job.ID)
		case queue.JobStatusCompleted:
			pipe.SAdd(ctx, s.completedSet, job.ID)
		case queue.JobStatusFailed:
			pipe.SAdd(ctx, s.failedSet, job.ID)
		case queue.JobStatusRetrying:
			pipe.ZAdd(ctx, s.retryingSet, redis.Z{
				Score:  float64(job.RunAt.Unix()),
				Member: job.ID,
			})
		default:
			return queue.ErrUnknownJobStatus
		}
	} else if job.Status == queue.JobStatusPending || job.Status == queue.JobStatusRetrying {
		// If RunAt changed for pending or retrying jobs, update the score
		switch job.Status {
		case queue.JobStatusPending:
			pipe.ZAdd(ctx, s.pendingSet, redis.Z{
				Score:  float64(job.RunAt.Unix()),
				Member: job.ID,
			})
		case queue.JobStatusRetrying:
			pipe.ZAdd(ctx, s.retryingSet, redis.Z{
				Score:  float64(job.RunAt.Unix()),
				Member: job.ID,
			})
		}
	}

	// Execute the transaction
	_, err = pipe.Exec(ctx)
	return err
}

// Delete removes a job from Redis.
func (s *Storage) Delete(ctx context.Context, id string) error {
	// Get job to determine which sets it belongs to
	job, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	// Delete the job in a transaction
	pipe := s.client.Pipeline()
	
	// Delete the job data
	pipe.Del(ctx, s.jobKey(id))
	
	// Remove from the appropriate set based on status
	switch job.Status {
	case queue.JobStatusPending:
		pipe.ZRem(ctx, s.pendingSet, id)
	case queue.JobStatusProcessing:
		pipe.SRem(ctx, s.processingSet, id)
	case queue.JobStatusCompleted:
		pipe.SRem(ctx, s.completedSet, id)
	case queue.JobStatusFailed:
		pipe.SRem(ctx, s.failedSet, id)
	case queue.JobStatusRetrying:
		pipe.ZRem(ctx, s.retryingSet, id)
	}

	// Execute the transaction
	_, err = pipe.Exec(ctx)
	return err
}

// FetchDue retrieves due jobs ready for processing,
// up to the specified limit, marking them as processing.
func (s *Storage) FetchDue(ctx context.Context, limit int) ([]*queue.Job, error) {
	now := time.Now().Unix()
	
	// Get pending jobs that are due
	pendingIDs, err := s.client.ZRangeByScore(ctx, s.pendingSet, &redis.ZRangeBy{
		Min:    "-inf",
		Max:    fmt.Sprintf("%d", now),
		Offset: 0,
		Count:  int64(limit),
	}).Result()
	
	if err != nil {
		return nil, err
	}
	
	// Also get retrying jobs that are due
	retryingIDs, err := s.client.ZRangeByScore(ctx, s.retryingSet, &redis.ZRangeBy{
		Min:    "-inf",
		Max:    fmt.Sprintf("%d", now),
		Offset: 0,
		Count:  int64(limit - len(pendingIDs)), // Adjust limit based on pending jobs found
	}).Result()

	if err != nil {
		return nil, err
	}

	// Combine both lists, limited to the requested limit
	allIDs := append(pendingIDs, retryingIDs...)
	if len(allIDs) > limit {
		allIDs = allIDs[:limit]
	}

	if len(allIDs) == 0 {
		return []*queue.Job{}, nil
	}

	var jobs []*queue.Job

	// Use Lua script to atomically fetch and update jobs
	// This script does the following:
	// 1. Fetch the job data
	// 2. Update the job status to processing
	// 3. Move the job from pending/retrying to processing set
	// 4. Return the original job data (before the update)
	script := `
	local jobs = {}
	for i, id in ipairs(ARGV) do
		local jobKey = KEYS[1] .. id
		local jobData = redis.call('GET', jobKey)
		if jobData then
			table.insert(jobs, jobData)
			local job = cjson.decode(jobData)
			
			-- Check which set the job is in
			local isPending = redis.call('ZREM', KEYS[2], id)
			local isRetrying = 0
			if isPending == 0 then
				isRetrying = redis.call('ZREM', KEYS[3], id)
			end
			
			if isPending == 1 or isRetrying == 1 then
				-- Update job status
				job.status = 'processing'
				job.updated_at = ARGV[#ARGV]
				local updatedData = cjson.encode(job)
				
				-- Update job and add to processing set
				redis.call('SET', jobKey, updatedData)
				redis.call('SADD', KEYS[4], id)
			end
		end
	end
	return jobs
	`

	// Execute the Lua script
	args := make([]interface{}, len(allIDs)+1)
	for i, id := range allIDs {
		args[i] = id
	}
	args[len(allIDs)] = time.Now().Format(time.RFC3339)
	
	res, err := s.client.Eval(ctx, script, []string{
		s.jobsPrefix,
		s.pendingSet,
		s.retryingSet,
		s.processingSet,
	}, args...).Result()
	
	if err != nil {
		return nil, err
	}

	// Parse results
	for _, data := range res.([]interface{}) {
		var job queue.Job
		if err := json.Unmarshal([]byte(data.(string)), &job); err != nil {
			continue
		}
		// Update status to processing
		job.Status = queue.JobStatusProcessing
		job.UpdatedAt = time.Now()
		jobs = append(jobs, &job)
	}

	return jobs, nil
}

// FetchByStatus retrieves jobs with the specified status,
// up to the specified limit.
func (s *Storage) FetchByStatus(ctx context.Context, status queue.JobStatus, limit int) ([]*queue.Job, error) {
	var ids []string
	var err error
	
	switch status {
	case queue.JobStatusPending:
		// Get all pending jobs, sorted by RunAt
		result, err := s.client.ZRange(ctx, s.pendingSet, 0, int64(limit-1)).Result()
		if err != nil {
			return nil, err
		}
		ids = result
	case queue.JobStatusProcessing:
		// Get all processing jobs
		result, err := s.client.SRandMemberN(ctx, s.processingSet, int64(limit)).Result()
		if err != nil {
			return nil, err
		}
		ids = result
	case queue.JobStatusCompleted:
		// Get all completed jobs
		result, err := s.client.SRandMemberN(ctx, s.completedSet, int64(limit)).Result()
		if err != nil {
			return nil, err
		}
		ids = result
	case queue.JobStatusFailed:
		// Get all failed jobs
		result, err := s.client.SRandMemberN(ctx, s.failedSet, int64(limit)).Result()
		if err != nil {
			return nil, err
		}
		ids = result
	case queue.JobStatusRetrying:
		// Get all retrying jobs, sorted by RunAt
		result, err := s.client.ZRange(ctx, s.retryingSet, 0, int64(limit-1)).Result()
		if err != nil {
			return nil, err
		}
		ids = result
	default:
		return nil, queue.ErrUnknownJobStatus
	}

	if len(ids) == 0 {
		return []*queue.Job{}, nil
	}

	// Fetch job data for all IDs
	var keys []string
	for _, id := range ids {
		keys = append(keys, s.jobKey(id))
	}

	// Fetch all jobs in a single call
	results, err := s.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	var jobs []*queue.Job
	for _, result := range results {
		if result == nil {
			continue
		}
		
		data, ok := result.(string)
		if !ok {
			continue
		}
		
		var job queue.Job
		if err := json.Unmarshal([]byte(data), &job); err != nil {
			continue
		}
		
		jobs = append(jobs, &job)
	}

	return jobs, nil
}

// PurgeCompleted removes completed jobs older than the specified duration.
func (s *Storage) PurgeCompleted(ctx context.Context, olderThan time.Duration) error {
	return s.purgeByStatus(ctx, queue.JobStatusCompleted, olderThan)
}

// PurgeFailed removes failed jobs older than the specified duration.
func (s *Storage) PurgeFailed(ctx context.Context, olderThan time.Duration) error {
	return s.purgeByStatus(ctx, queue.JobStatusFailed, olderThan)
}

// purgeByStatus is a helper method to purge jobs of a specific status.
func (s *Storage) purgeByStatus(ctx context.Context, status queue.JobStatus, olderThan time.Duration) error {
	// Calculate cutoff time
	cutoff := time.Now().Add(-olderThan).Unix()
	
	var setKey string
	switch status {
	case queue.JobStatusCompleted:
		setKey = s.completedSet
	case queue.JobStatusFailed:
		setKey = s.failedSet
	default:
		return queue.ErrUnknownJobStatus
	}

	// Use Lua script to safely purge jobs
	script := `
	local jobsToDelete = {}
	local members = redis.call('SMEMBERS', KEYS[1])
	
	for _, id in ipairs(members) do
		local jobKey = KEYS[2] .. id
		local jobData = redis.call('GET', jobKey)
		
		if jobData then
			local job = cjson.decode(jobData)
			if tonumber(job.updated_at) < tonumber(ARGV[1]) then
				redis.call('DEL', jobKey)
				redis.call('SREM', KEYS[1], id)
			end
		else
			-- Job data missing, but ID is in set - clean up
			redis.call('SREM', KEYS[1], id)
		end
	end
	
	return #jobsToDelete
	`

	// Execute the script
	_, err := s.client.Eval(ctx, script, []string{
		setKey,
		s.jobsPrefix,
	}, cutoff).Result()

	return err
}

// Size returns the total number of jobs in the storage.
func (s *Storage) Size(ctx context.Context) (int, error) {
	pipe := s.client.Pipeline()
	
	pendingCount := pipe.ZCard(ctx, s.pendingSet)
	processingCount := pipe.SCard(ctx, s.processingSet)
	completedCount := pipe.SCard(ctx, s.completedSet)
	failedCount := pipe.SCard(ctx, s.failedSet)
	retryingCount := pipe.ZCard(ctx, s.retryingSet)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	
	totalCount := pendingCount.Val() + processingCount.Val() + completedCount.Val() + failedCount.Val() + retryingCount.Val()
	return int(totalCount), nil
}

// Close closes the Redis connection.
func (s *Storage) Close(ctx context.Context) error {
	return s.client.Close()
}

// CleanStaleJobs moves jobs that have been in the processing state for longer than
// the lock timeout back to the pending state so they can be reprocessed.
// This is useful for recovering from worker crashes.
func (s *Storage) CleanStaleJobs(ctx context.Context) error {
	// Get all jobs in the processing set
	processingIDs, err := s.client.SMembers(ctx, s.processingSet).Result()
	if err != nil {
		return err
	}

	if len(processingIDs) == 0 {
		return nil
	}

	now := time.Now()
	lockTimeoutSeconds := int64(s.lockTimeout / time.Second)

	// Get jobs data
	var keys []string
	for _, id := range processingIDs {
		keys = append(keys, s.jobKey(id))
	}

	// Use Lua script to identify and move stale jobs atomically
	script := `
	local staleJobs = {}
	local now = tonumber(ARGV[1])
	local timeout = tonumber(ARGV[2])
	
	for i, key in ipairs(KEYS) do
		local id = ARGV[i+2]
		local jobKey = KEYS[1] .. id
		local jobData = redis.call('GET', jobKey)
		
		if jobData then
			local job = cjson.decode(jobData)
			
			-- Check if the job has been processing for too long
			local updatedAt = job.updated_at and tonumber(job.updated_at) or 0
			if now - updatedAt > timeout then
				-- Move back to pending
				job.status = 'pending'
				job.updated_at = now
				
				redis.call('SET', jobKey, cjson.encode(job))
				redis.call('SREM', KEYS[#KEYS], id)
				redis.call('ZADD', KEYS[#KEYS-1], now, id)
				
				table.insert(staleJobs, id)
			end
		else
			-- Job data missing, but ID is in set - clean up
			redis.call('SREM', KEYS[#KEYS], id)
		end
	end
	
	return staleJobs
	`

	// Add the processing set and pending set keys to the end of the keys list
	keys = append(keys, s.pendingSet, s.processingSet)
	
	// Prepare arguments: now, timeout, and all job IDs
	args := make([]interface{}, len(processingIDs)+2)
	args[0] = now.Unix()
	args[1] = lockTimeoutSeconds
	for i, id := range processingIDs {
		args[i+2] = id
	}
	
	// Execute the script
	_, err = s.client.Eval(ctx, script, keys, args...).Result()
	return err
}
