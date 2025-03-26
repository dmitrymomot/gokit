package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/dmitrymomot/gokit/queue"
	"github.com/dmitrymomot/gokit/queue/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	storage := memory.New()
	require.NotNil(t, storage)
}

func TestPing(t *testing.T) {
	storage := memory.New()
	err := storage.Ping(context.Background())
	require.NoError(t, err)

	// Close the storage and ensure Ping returns an error
	err = storage.Close(context.Background())
	require.NoError(t, err)

	err = storage.Ping(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, queue.ErrStorageUnavailable)
}

func TestPut(t *testing.T) {
	storage := memory.New()
	
	// Create a job
	job, err := queue.NewJob("test_task", map[string]string{"key": "value"})
	require.NoError(t, err)
	
	// Put the job in storage
	err = storage.Put(context.Background(), job)
	require.NoError(t, err)
	
	// ID should have been generated
	assert.NotEmpty(t, job.ID)
	
	// Try to retrieve the job
	retrieved, err := storage.Get(context.Background(), job.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	
	// Job should match
	assert.Equal(t, job.ID, retrieved.ID)
	assert.Equal(t, job.TaskName, retrieved.TaskName)
	assert.Equal(t, job.Status, retrieved.Status)
}

func TestGet(t *testing.T) {
	storage := memory.New()
	
	// Try to get a non-existent job
	_, err := storage.Get(context.Background(), "non-existent-id")
	require.Error(t, err)
	assert.ErrorIs(t, err, queue.ErrJobNotFound)
	
	// Create and store a job
	job, err := queue.NewJob("test_task", map[string]string{"key": "value"})
	require.NoError(t, err)
	
	err = storage.Put(context.Background(), job)
	require.NoError(t, err)
	
	// Get the job
	retrieved, err := storage.Get(context.Background(), job.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	
	// Job should match but be a different instance
	assert.Equal(t, job.ID, retrieved.ID)
	assert.Equal(t, job.TaskName, retrieved.TaskName)
	assert.NotSame(t, job, retrieved)
}

func TestUpdate(t *testing.T) {
	storage := memory.New()
	
	// Try to update a non-existent job
	job, err := queue.NewJob("test_task", map[string]string{"key": "value"})
	require.NoError(t, err)
	job.ID = "non-existent-id"
	
	err = storage.Update(context.Background(), job)
	require.Error(t, err)
	assert.ErrorIs(t, err, queue.ErrJobNotFound)
	
	// Create and store a job
	job, err = queue.NewJob("test_task", map[string]string{"key": "value"})
	require.NoError(t, err)
	
	err = storage.Put(context.Background(), job)
	require.NoError(t, err)
	
	// Update the job
	job.Status = queue.JobStatusProcessing
	job.LastError = "test error"
	
	err = storage.Update(context.Background(), job)
	require.NoError(t, err)
	
	// Get the job and verify updates
	retrieved, err := storage.Get(context.Background(), job.ID)
	require.NoError(t, err)
	
	assert.Equal(t, queue.JobStatusProcessing, retrieved.Status)
	assert.Equal(t, "test error", retrieved.LastError)
}

func TestDelete(t *testing.T) {
	storage := memory.New()
	
	// Try to delete a non-existent job
	err := storage.Delete(context.Background(), "non-existent-id")
	require.Error(t, err)
	assert.ErrorIs(t, err, queue.ErrJobNotFound)
	
	// Create and store a job
	job, err := queue.NewJob("test_task", map[string]string{"key": "value"})
	require.NoError(t, err)
	
	err = storage.Put(context.Background(), job)
	require.NoError(t, err)
	
	// Delete the job
	err = storage.Delete(context.Background(), job.ID)
	require.NoError(t, err)
	
	// Try to get the deleted job
	_, err = storage.Get(context.Background(), job.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, queue.ErrJobNotFound)
}

func TestFetchDue(t *testing.T) {
	storage := memory.New()
	
	// Create jobs with different statuses and run times
	now := time.Now()
	
	// Job due now (should be fetched)
	job1, _ := queue.NewJob("task1", nil)
	job1.Status = queue.JobStatusPending
	job1.RunAt = now.Add(-1 * time.Hour)
	storage.Put(context.Background(), job1)
	
	// Job due in the future (should not be fetched)
	job2, _ := queue.NewJob("task2", nil)
	job2.Status = queue.JobStatusPending
	job2.RunAt = now.Add(1 * time.Hour)
	storage.Put(context.Background(), job2)
	
	// Job with non-pending status (should not be fetched)
	job3, _ := queue.NewJob("task3", nil)
	job3.Status = queue.JobStatusCompleted
	job3.RunAt = now.Add(-1 * time.Hour)
	storage.Put(context.Background(), job3)
	
	// Fetch due jobs
	jobs, err := storage.FetchDue(context.Background(), 10)
	require.NoError(t, err)
	
	// Should only get the first job
	assert.Len(t, jobs, 1)
	assert.Equal(t, job1.ID, jobs[0].ID)
	
	// The status should be updated to processing
	assert.Equal(t, queue.JobStatusProcessing, jobs[0].Status)
	
	// Original job in storage should also be updated
	updatedJob, err := storage.Get(context.Background(), job1.ID)
	require.NoError(t, err)
	assert.Equal(t, queue.JobStatusProcessing, updatedJob.Status)
}

func TestFetchByStatus(t *testing.T) {
	storage := memory.New()
	
	// Create jobs with different statuses
	job1, _ := queue.NewJob("task1", nil)
	job1.Status = queue.JobStatusPending
	storage.Put(context.Background(), job1)
	
	job2, _ := queue.NewJob("task2", nil)
	job2.Status = queue.JobStatusProcessing
	storage.Put(context.Background(), job2)
	
	job3, _ := queue.NewJob("task3", nil)
	job3.Status = queue.JobStatusCompleted
	storage.Put(context.Background(), job3)
	
	job4, _ := queue.NewJob("task4", nil)
	job4.Status = queue.JobStatusFailed
	storage.Put(context.Background(), job4)
	
	job5, _ := queue.NewJob("task5", nil)
	job5.Status = queue.JobStatusCompleted
	storage.Put(context.Background(), job5)
	
	// Fetch by status
	jobs, err := storage.FetchByStatus(context.Background(), queue.JobStatusCompleted, 10)
	require.NoError(t, err)
	
	// Should get 2 completed jobs
	assert.Len(t, jobs, 2)
	
	// Check if all returned jobs have the correct status
	for _, job := range jobs {
		assert.Equal(t, queue.JobStatusCompleted, job.Status)
	}
	
	// Test with limit
	jobs, err = storage.FetchByStatus(context.Background(), queue.JobStatusCompleted, 1)
	require.NoError(t, err)
	
	// Should only get 1 job
	assert.Len(t, jobs, 1)
	assert.Equal(t, queue.JobStatusCompleted, jobs[0].Status)
}

func TestPurgeCompleted(t *testing.T) {
	storage := memory.New()
	
	// Create completed jobs with different update times
	now := time.Now()
	
	// Old job (should be purged)
	job1, _ := queue.NewJob("task1", nil)
	job1.Status = queue.JobStatusCompleted
	job1.UpdatedAt = now.Add(-2 * time.Hour)
	storage.Put(context.Background(), job1)
	
	// Recent job (should not be purged)
	job2, _ := queue.NewJob("task2", nil)
	job2.Status = queue.JobStatusCompleted
	job2.UpdatedAt = now.Add(-30 * time.Minute)
	storage.Put(context.Background(), job2)
	
	// Job with different status (should not be purged)
	job3, _ := queue.NewJob("task3", nil)
	job3.Status = queue.JobStatusFailed
	job3.UpdatedAt = now.Add(-2 * time.Hour)
	storage.Put(context.Background(), job3)
	
	// Purge completed jobs older than 1 hour
	err := storage.PurgeCompleted(context.Background(), 1*time.Hour)
	require.NoError(t, err)
	
	// Job1 should be purged
	_, err = storage.Get(context.Background(), job1.ID)
	assert.ErrorIs(t, err, queue.ErrJobNotFound)
	
	// Job2 should still exist
	j2, err := storage.Get(context.Background(), job2.ID)
	require.NoError(t, err)
	assert.Equal(t, job2.ID, j2.ID)
	
	// Job3 should still exist
	j3, err := storage.Get(context.Background(), job3.ID)
	require.NoError(t, err)
	assert.Equal(t, job3.ID, j3.ID)
}

func TestPurgeFailed(t *testing.T) {
	storage := memory.New()
	
	// Create failed jobs with different update times
	now := time.Now()
	
	// Old job (should be purged)
	job1, _ := queue.NewJob("task1", nil)
	job1.Status = queue.JobStatusFailed
	job1.UpdatedAt = now.Add(-2 * time.Hour)
	storage.Put(context.Background(), job1)
	
	// Recent job (should not be purged)
	job2, _ := queue.NewJob("task2", nil)
	job2.Status = queue.JobStatusFailed
	job2.UpdatedAt = now.Add(-30 * time.Minute)
	storage.Put(context.Background(), job2)
	
	// Job with different status (should not be purged)
	job3, _ := queue.NewJob("task3", nil)
	job3.Status = queue.JobStatusCompleted
	job3.UpdatedAt = now.Add(-2 * time.Hour)
	storage.Put(context.Background(), job3)
	
	// Purge failed jobs older than 1 hour
	err := storage.PurgeFailed(context.Background(), 1*time.Hour)
	require.NoError(t, err)
	
	// Job1 should be purged
	_, err = storage.Get(context.Background(), job1.ID)
	assert.ErrorIs(t, err, queue.ErrJobNotFound)
	
	// Job2 should still exist
	j2, err := storage.Get(context.Background(), job2.ID)
	require.NoError(t, err)
	assert.Equal(t, job2.ID, j2.ID)
	
	// Job3 should still exist
	j3, err := storage.Get(context.Background(), job3.ID)
	require.NoError(t, err)
	assert.Equal(t, job3.ID, j3.ID)
}

func TestSize(t *testing.T) {
	storage := memory.New()
	
	// Empty storage
	size, err := storage.Size(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, size)
	
	// Add some jobs
	for i := 0; i < 5; i++ {
		job, _ := queue.NewJob("task", nil)
		storage.Put(context.Background(), job)
	}
	
	// Check size
	size, err = storage.Size(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 5, size)
	
	// Delete a job
	jobs, _ := storage.FetchByStatus(context.Background(), queue.JobStatusPending, 1)
	storage.Delete(context.Background(), jobs[0].ID)
	
	// Check size again
	size, err = storage.Size(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 4, size)
}

func TestClose(t *testing.T) {
	storage := memory.New()
	
	// Add a job
	job, _ := queue.NewJob("task", nil)
	err := storage.Put(context.Background(), job)
	require.NoError(t, err)
	
	// Close the storage
	err = storage.Close(context.Background())
	require.NoError(t, err)
	
	// All operations should now fail
	_, err = storage.Get(context.Background(), job.ID)
	assert.ErrorIs(t, err, queue.ErrStorageUnavailable)
	
	err = storage.Put(context.Background(), job)
	assert.ErrorIs(t, err, queue.ErrStorageUnavailable)
	
	err = storage.Update(context.Background(), job)
	assert.ErrorIs(t, err, queue.ErrStorageUnavailable)
	
	err = storage.Delete(context.Background(), job.ID)
	assert.ErrorIs(t, err, queue.ErrStorageUnavailable)
	
	_, err = storage.FetchDue(context.Background(), 10)
	assert.ErrorIs(t, err, queue.ErrStorageUnavailable)
	
	_, err = storage.FetchByStatus(context.Background(), queue.JobStatusPending, 10)
	assert.ErrorIs(t, err, queue.ErrStorageUnavailable)
	
	err = storage.PurgeCompleted(context.Background(), 1*time.Hour)
	assert.ErrorIs(t, err, queue.ErrStorageUnavailable)
	
	err = storage.PurgeFailed(context.Background(), 1*time.Hour)
	assert.ErrorIs(t, err, queue.ErrStorageUnavailable)
	
	_, err = storage.Size(context.Background())
	assert.ErrorIs(t, err, queue.ErrStorageUnavailable)
}
