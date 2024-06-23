package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"microbatch-service/microbatch"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

// MockBatchProcessor is a mock implementation of BatchProcessor for demonstration
type MockBatchProcessor struct{}

// ProcessBatch processes a batch of jobs and returns the results.
//
// Parameters:
// - jobs: a slice of Job objects representing the jobs to be processed.
// Return type: a slice of JobResult objects representing the results of the processed jobs.
func (bp *MockBatchProcessor) ProcessBatch(jobs []microbatch.Job) []microbatch.JobResult {
	results := make([]microbatch.JobResult, len(jobs))
	for i, job := range jobs {
		payload := job.Payload.(map[string]interface{})
		fmt.Printf("Processing job ID: %d, Payload: %v\n", job.ID, payload)
		results[i] = microbatch.JobResult{ID: job.ID, Result: payload}
	}
	return results
}

// main is the entry point of the program.
//
// It initializes the batch size and batch interval from environment variables.
// It creates a MockBatchProcessor and a MicroBatcher with the specified batch size and batch interval.
// It sets up HTTP handlers for submitting jobs, getting job results, and shutting down the server.
// It starts the HTTP server and listens for incoming requests.
//
// No parameters.
// No return type.
func main() {
	batchSize, err := strconv.Atoi(getEnv("BATCH_SIZE", "5"))
	if err != nil {
		log.Fatalf("Invalid BATCH_SIZE: %v", err)
	}
	batchInterval, err := strconv.Atoi(getEnv("BATCH_INTERVAL", "5"))
	if err != nil {
		log.Fatalf("Invalid BATCH_INTERVAL: %v", err)
	}

	processor := &MockBatchProcessor{}
	batcher := microbatch.NewMicroBatcher(processor, batchSize, time.Duration(batchInterval)*time.Second)

	var jobIDCounter int
	var jobIDMutex sync.Mutex

	server := &http.Server{Addr: ":8080"}

	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		jobIDMutex.Lock()
		jobIDCounter++
		jobID := jobIDCounter
		jobIDMutex.Unlock()

		batcher.SubmitJob(microbatch.Job{ID: jobID, Payload: payload})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ID": jobID})
	})

	// add a handler to get job result by job id
	http.HandleFunc("/result", func(w http.ResponseWriter, r *http.Request) {
		jobIDStr := r.URL.Query().Get("id")
		jobID, err := strconv.Atoi(jobIDStr)
		if err != nil {
			http.Error(w, "Invalid job ID", http.StatusBadRequest)
			return
		}
		result, ok := batcher.GetJobResult(jobID)
		if !ok {
			http.Error(w, "Job not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	http.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		go func() {
			batcher.Shutdown()
			if err := server.Shutdown(context.Background()); err != nil {
				log.Fatalf("HTTP server Shutdown: %v", err)
			}
			fmt.Fprintln(w, "MicroBatcher shutdown completed")
		}()
	})

	fmt.Println("Starting server on :8080")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe(): %v", err)
	}
}

// getEnv retrieves the value of an environment variable with the specified key.
//
// Parameters:
// - key: the name of the environment variable to retrieve.
// - defaultValue: the default value to return if the environment variable is not set.
//
// Return type:
// - string: the value of the environment variable, or the default value if the environment variable is not set.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
