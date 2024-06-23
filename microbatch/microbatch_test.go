package microbatch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockBatchProcessor is a mock implementation of BatchProcessor for testing
type MockBatchProcessor struct {
	ProcessedJobs []Job
}

// ProcessBatch processes a batch of jobs and returns the results.
//
// Parameters:
// - jobs: a slice of Job objects representing the jobs to be processed.
// Return type: a slice of JobResult objects representing the results of the processed jobs.
func (bp *MockBatchProcessor) ProcessBatch(jobs []Job) []JobResult {
	bp.ProcessedJobs = append(bp.ProcessedJobs, jobs...)
	results := make([]JobResult, len(jobs))
	for i, job := range jobs {
		results[i] = JobResult{ID: job.ID, Result: job.Payload}
	}
	return results
}

// TestMicroBatcher_ProcessesJobsImmediatelyWhenBatchSizeReached tests that the MicroBatcher processes jobs immediately when the batch size is reached.
//
// This test creates a new MicroBatcher instance with a mock batch processor and a batch size of 5 and a batch interval of 1 second.
// It then submits jobs to the MicroBatcher until the batch size is reached. After submitting the jobs, it waits for a brief period of time to allow processing.
// Finally, it asserts that all jobs were processed by checking the length of the processed jobs in the mock batch processor and comparing them to the expected batch size.
func TestMicroBatcher_ProcessesJobsImmediatelyWhenBatchSizeReached(t *testing.T) {
	batchProcessor := &MockBatchProcessor{}
	batchSize := 5
	batchInterval := 1 * time.Second

	batcher := NewMicroBatcher(batchProcessor, batchSize, batchInterval)

	// Submit jobs
	for i := 0; i < batchSize; i++ {
		batcher.SubmitJob(Job{ID: i, Payload: i})
	}

	// Wait briefly to allow processing
	time.Sleep(100 * time.Millisecond)

	// Assert that all jobs were processed
	assert.Len(t, batchProcessor.ProcessedJobs, batchSize)
	for i := 0; i < batchSize; i++ {
		assert.Equal(t, Job{ID: i, Payload: i}, batchProcessor.ProcessedJobs[i])
	}

	batcher.Shutdown()
}

// TestMicroBatcher_ShutdownProcessesAllJobs tests that the MicroBatcher processes all jobs when it is shutdown.
//
// This test creates a new MicroBatcher instance with a mock batch processor and a batch size of 5 and a batch interval of 10 seconds.
// It then submits 3 jobs to the MicroBatcher.
// After submitting the jobs, it shuts down the MicroBatcher and asserts that all 3 jobs are processed by checking the length of the processed jobs in the mock batch processor and comparing them to the expected number of jobs.
func TestMicroBatcher_ShutdownProcessesAllJobs(t *testing.T) {
	batchProcessor := &MockBatchProcessor{}
	batchSize := 5
	batchInterval := 10 * time.Second // Set a long interval to ensure jobs don't process automatically

	batcher := NewMicroBatcher(batchProcessor, batchSize, batchInterval)

	// Submit fewer jobs than batch size
	numJobs := 3
	for i := 0; i < numJobs; i++ {
		batcher.SubmitJob(Job{ID: i, Payload: i})
	}

	// Shutdown the batcher and assert all jobs are processed
	batcher.Shutdown()

	assert.Len(t, batchProcessor.ProcessedJobs, numJobs)
	for i := 0; i < numJobs; i++ {
		assert.Equal(t, Job{ID: i, Payload: i}, batchProcessor.ProcessedJobs[i])
	}
}

// TestMicroBatcher_ProcessesJobsWhenIntervalPasses tests the behavior of the MicroBatcher when the interval passes.
//
// It creates a new MicroBatcher instance with a mock batch processor, a batch size of 5, and a batch interval of 1 second.
// It then submits fewer jobs than the batch size.
// After submitting the jobs, it waits for the batch interval to elapse.
// Finally, it asserts that all jobs were processed by checking the length of the processed jobs in the mock batch processor and comparing them to the expected number of jobs.
func TestMicroBatcher_ProcessesJobsWhenIntervalPasses(t *testing.T) {
	batchProcessor := &MockBatchProcessor{}
	batchSize := 5
	batchInterval := 1 * time.Second

	batcher := NewMicroBatcher(batchProcessor, batchSize, batchInterval)

	// Submit fewer jobs than batch size
	numJobs := 3
	for i := 0; i < numJobs; i++ {
		batcher.SubmitJob(Job{ID: i, Payload: i})
	}

	// Wait for batch interval to elapse
	time.Sleep(batchInterval + 500*time.Millisecond) // Adding extra buffer time

	// Assert that all jobs were processed
	assert.Len(t, batchProcessor.ProcessedJobs, numJobs)
	for i := 0; i < numJobs; i++ {
		assert.Equal(t, Job{ID: i, Payload: i}, batchProcessor.ProcessedJobs[i])
	}

	batcher.Shutdown()
}

// TestMicroBatcher_TickerResetsAfterBatchProcessing tests the behavior of the MicroBatcher after a batch is processed.
//
// It creates a new MicroBatcher instance with a mock batch processor, a batch size of 3, and a batch interval of 3 seconds.
// It then waits for 2 seconds before submitting 3 jobs to the MicroBatcher.
// After submitting the jobs, it waits briefly to allow processing.
// It asserts that the batch is run by checking the length of the processed jobs in the mock batch processor and comparing it to the expected batch size of 3.
// It then submits 2 more jobs to the MicroBatcher.
// It waits for 2 seconds (less than the batch interval) to ensure the batch hasn't been run yet.
// It asserts that the batch hasn't been run yet by checking the length of the processed jobs in the mock batch processor and comparing it to the expected batch size of 3.
// It waits for another 1 second (completing the batch interval) to allow the batch to be processed.
// It asserts that the batch is run by checking the length of the processed jobs in the mock batch processor and comparing it to the expected batch size of 5.
// Finally, it shuts down the MicroBatcher.
func TestMicroBatcher_TickerResetsAfterBatchProcessing(t *testing.T) {
	batchProcessor := &MockBatchProcessor{}
	batchSize := 3
	batchInterval := 3 * time.Second

	batcher := NewMicroBatcher(batchProcessor, batchSize, batchInterval)

	// Wait for 2 seconds before submitting jobs
	time.Sleep(2 * time.Second)

	// Submit 3 jobs
	for i := 0; i < 3; i++ {
		batcher.SubmitJob(Job{ID: i, Payload: i})
	}

	// Wait briefly to allow processing
	time.Sleep(100 * time.Millisecond)

	// Assert that the batch is run
	assert.Len(t, batchProcessor.ProcessedJobs, 3)

	// Submit 2 more jobs
	for i := 3; i < 5; i++ {
		batcher.SubmitJob(Job{ID: i, Payload: i})
	}

	// Wait for 2 seconds (less than batch interval)
	time.Sleep(2 * time.Second)

	// Assert that the batch hasn't been run yet
	assert.Len(t, batchProcessor.ProcessedJobs, 3)

	// Wait another 1 second (completing the batch interval)
	time.Sleep(1 * time.Second)

	// Assert that the batch is run
	assert.Len(t, batchProcessor.ProcessedJobs, 5)

	batcher.Shutdown()
}
