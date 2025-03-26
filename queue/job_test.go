package queue_test

import (
	"testing"
	"time"

	"github.com/dmitrymomot/gokit/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testPayload struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestNewJob(t *testing.T) {
	// Create a new job
	payload := testPayload{Name: "test", Value: 123}
	job, err := queue.NewJob("test_task", payload)
	
	// Validate job was created successfully
	require.NoError(t, err)
	require.NotNil(t, job)
	
	// Validate job fields
	assert.Equal(t, "test_task", job.TaskName)
	assert.NotEmpty(t, job.Payload)
	assert.Equal(t, queue.JobStatusPending, job.Status)
	assert.False(t, job.CreatedAt.IsZero())
	assert.False(t, job.UpdatedAt.IsZero())
	assert.False(t, job.RunAt.IsZero())
	assert.Equal(t, 0, job.RetryCount)
	assert.Equal(t, 3, job.MaxRetries) // Default max retries
	assert.Empty(t, job.LastError)
}

func TestGetPayload(t *testing.T) {
	// Create a job with a test payload
	originalPayload := testPayload{Name: "test", Value: 123}
	job, err := queue.NewJob("test_task", originalPayload)
	require.NoError(t, err)
	require.NotNil(t, job)
	
	// Deserialize the payload
	var retrievedPayload testPayload
	err = job.GetPayload(&retrievedPayload)
	
	// Validate payload was deserialized correctly
	require.NoError(t, err)
	assert.Equal(t, originalPayload.Name, retrievedPayload.Name)
	assert.Equal(t, originalPayload.Value, retrievedPayload.Value)
}

func TestGetPayload_Error(t *testing.T) {
	// Create a job with a test payload
	originalPayload := testPayload{Name: "test", Value: 123}
	job, err := queue.NewJob("test_task", originalPayload)
	require.NoError(t, err)
	require.NotNil(t, job)
	
	// Intentionally cause an error by using a non-pointer receiver
	var wrongType int
	
	// This should return an error
	err = job.GetPayload(&wrongType)
	require.Error(t, err)
	assert.ErrorIs(t, err, queue.ErrInvalidJobPayload)
}

func TestClone(t *testing.T) {
	// Create a job
	original, err := queue.NewJob("test_task", testPayload{Name: "test", Value: 123})
	require.NoError(t, err)
	original.ID = "test-id"
	original.Status = queue.JobStatusProcessing
	original.LastError = "test error"
	
	// Clone the job
	clone := original.Clone()
	
	// Validate the clone has the same values
	assert.Equal(t, original.ID, clone.ID)
	assert.Equal(t, original.TaskName, clone.TaskName)
	assert.Equal(t, original.Status, clone.Status)
	assert.Equal(t, original.CreatedAt, clone.CreatedAt)
	assert.Equal(t, original.UpdatedAt, clone.UpdatedAt)
	assert.Equal(t, original.RunAt, clone.RunAt)
	assert.Equal(t, original.RetryCount, clone.RetryCount)
	assert.Equal(t, original.MaxRetries, clone.MaxRetries)
	assert.Equal(t, original.LastError, clone.LastError)
	
	// Ensure the payload is a deep copy
	assert.Equal(t, len(original.Payload), len(clone.Payload))
	assert.NotSame(t, &original.Payload, &clone.Payload)
	
	// Modify the clone and ensure the original is unchanged
	clone.Status = queue.JobStatusCompleted
	assert.NotEqual(t, original.Status, clone.Status)
}

func TestShouldRetry(t *testing.T) {
	// Create a job
	job, err := queue.NewJob("test_task", testPayload{})
	require.NoError(t, err)
	
	// Test with different retry counts
	testCases := []struct {
		retryCount int
		maxRetries int
		expected   bool
	}{
		{0, 3, true},  // 0 retries, max 3
		{1, 3, true},  // 1 retry, max 3
		{2, 3, true},  // 2 retries, max 3
		{3, 3, false}, // 3 retries, max 3
		{4, 3, false}, // 4 retries, max 3
		{0, 0, false}, // 0 retries, max 0
	}
	
	for _, tc := range testCases {
		job.RetryCount = tc.retryCount
		job.MaxRetries = tc.maxRetries
		
		assert.Equal(t, tc.expected, job.ShouldRetry(), 
			"ShouldRetry() failed with retryCount=%d, maxRetries=%d", 
			tc.retryCount, tc.maxRetries)
	}
}

func TestIsReady(t *testing.T) {
	// Create a job
	job, err := queue.NewJob("test_task", testPayload{})
	require.NoError(t, err)
	
	// Test with different statuses and run times
	now := time.Now()
	
	testCases := []struct {
		status queue.JobStatus
		runAt  time.Time
		expected bool
	}{
		{queue.JobStatusPending, now.Add(-1 * time.Hour), true},  // Pending and in the past
		{queue.JobStatusPending, now.Add(1 * time.Hour), false},  // Pending but in the future
		{queue.JobStatusProcessing, now.Add(-1 * time.Hour), false}, // Not pending
		{queue.JobStatusCompleted, now.Add(-1 * time.Hour), false},  // Not pending
		{queue.JobStatusFailed, now.Add(-1 * time.Hour), false},     // Not pending
		{queue.JobStatusRetrying, now.Add(-1 * time.Hour), false},   // Not pending
	}
	
	for _, tc := range testCases {
		job.Status = tc.status
		job.RunAt = tc.runAt
		
		assert.Equal(t, tc.expected, job.IsReady(), 
			"IsReady() failed with status=%s, runAt=%v", 
			tc.status, tc.runAt)
	}
}
