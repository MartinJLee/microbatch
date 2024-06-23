package microbatch

import (
	"sync"
	"time"
)

// Job represents a unit of work to be processed by the MicroBatcher.
type Job struct {
	ID      int
	Payload interface{}
}

// JobResult represents the result of processing a Job.
type JobResult struct {
	ID     int
	Result interface{}
}

// BatchProcessor is an interface that defines the method for processing a batch of jobs.
type BatchProcessor interface {
	ProcessBatch(jobs []Job) []JobResult
}

type MicroBatcher struct {
	batchProcessor BatchProcessor
	batchSize      int
	batchInterval  time.Duration
	jobs           []Job
	jobsMutex      sync.Mutex
	results        map[int]JobResult
	resultsMutex  sync.Mutex
	stopCh         chan struct{}
	wg             sync.WaitGroup
	ticker         *time.Ticker
}

// NewMicroBatcher creates a new MicroBatcher with the specified batch processor, batch size, and batch interval.
// It initializes the MicroBatcher fields, starts the processing routine, and returns the created MicroBatcher instance.
func NewMicroBatcher(batchProcessor BatchProcessor, batchSize int, batchInterval time.Duration) *MicroBatcher {
	mb := &MicroBatcher{
		batchProcessor: batchProcessor,
		batchSize:      batchSize,
		batchInterval:  batchInterval,
		results:        make(map[int]JobResult),
		stopCh:         make(chan struct{}),
		ticker:         time.NewTicker(batchInterval),
	}
	mb.start()
	return mb
}

// start initializes the MicroBatcher, starts a goroutine to handle batch processing,
// and listens for ticker and stop channel events to trigger the batch processing.
func (mb *MicroBatcher) start() {
	mb.wg.Add(1)
	go func() {
		defer mb.wg.Done()
		defer mb.ticker.Stop()

		for {
			select {
			case <-mb.ticker.C:
				mb.processBatch()
			case <-mb.stopCh:
				mb.processBatch()
				return
			}
		}
	}()
}

// SubmitJob adds a job to the MicroBatcher's job list and processes the batch if the
// number of jobs reaches the batch size.
//
// Parameters:
// - mb: a pointer to the MicroBatcher instance.
// - job: a Job object representing the job to be added.
//
// Return type: None.
func (mb *MicroBatcher) SubmitJob(job Job) {
	mb.jobsMutex.Lock()
	mb.jobs = append(mb.jobs, job)
	mb.jobsMutex.Unlock()

	if len(mb.jobs) >= mb.batchSize {
		mb.processBatch()
	}
}

// GetJobResult retrieves the job result and a boolean indicating whether the result exists or not.
//
// Parameters:
// - id: The ID of the job.
//
// Returns:
// - JobResult: The job result.
// - bool: A boolean indicating whether the result exists or not.
func (mb *MicroBatcher) GetJobResult(id int) (JobResult, bool) {
	mb.resultsMutex.Lock()
	defer mb.resultsMutex.Unlock()
	result, ok := mb.results[id]
	return result, ok
}

// processBatch processes the batch of jobs in the MicroBatcher.
//
// No parameters.
func (mb *MicroBatcher) processBatch() {
	mb.ticker.Reset(mb.batchInterval)
	mb.jobsMutex.Lock()
	defer mb.jobsMutex.Unlock()

	if len(mb.jobs) == 0 {
		return
	}

	jobsToProcess := mb.jobs
	mb.jobs = nil

	mb.wg.Add(1)
	go func() {
		defer mb.wg.Done()
		results := mb.batchProcessor.ProcessBatch(jobsToProcess)
		mb.resultsMutex.Lock()
		defer mb.resultsMutex.Unlock()
		for _, result := range results {
			mb.results[result.ID] = result
		}
	}()
}

// Shutdown stops the MicroBatcher by closing the stop channel and waiting for all goroutines to finish.
//
// No parameters.
// No return values.
func (mb *MicroBatcher) Shutdown() {
	close(mb.stopCh)
	mb.wg.Wait()
}
