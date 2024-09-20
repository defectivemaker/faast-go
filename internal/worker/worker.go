package worker

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"

	"faast-go/internal/curl"

	"github.com/schollz/progressbar/v3"
)

type CurlResult struct {
	Payload  []string
	Response *http.Response
	Err      error
}

type WorkerPool struct {
	config      *curl.CurlConfig
	permChan    <-chan []string
	resultChan  chan<- CurlResult
	progressBar *progressbar.ProgressBar
	numWorkers  int
	wg          sync.WaitGroup
	workerCount int32
}

func NewWorkerPool(config *curl.CurlConfig, permChan <-chan []string, resultChan chan<- CurlResult, progressBar *progressbar.ProgressBar) *WorkerPool {
	return &WorkerPool{
		config:      config,
		permChan:    permChan,
		resultChan:  resultChan,
		progressBar: progressBar,
		numWorkers:  10, // Adjust based on your needs and rate limits
	}
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go func() {
			defer wp.wg.Done()
			atomic.AddInt32(&wp.workerCount, 1)
			wp.worker()
		}()
	}
}

func (wp *WorkerPool) Wait() {
	wp.wg.Wait()
}

func (wp *WorkerPool) worker() {
	for perm := range wp.permChan {
		payload, err := wp.config.ConstructPayload(perm)
		if err != nil {
			wp.resultChan <- CurlResult{Payload: perm, Err: err}
			continue
		}
		res, err := wp.config.SendCurl(context.Background(), payload)
		wp.progressBar.Add(1)
		wp.resultChan <- CurlResult{Payload: perm, Response: res, Err: err}
	}
}
