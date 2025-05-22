package rtcm

import (
	"context"
	"sync"
)

// WorkerPool represents a pool of workers for processing RTCM messages
type WorkerPool struct {
	numWorkers int
	jobQueue   chan *RTCMMessage
	results    chan interface{}
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewWorkerPool creates a new worker pool with the specified number of workers
func NewWorkerPool(numWorkers int, queueSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	
	pool := &WorkerPool{
		numWorkers: numWorkers,
		jobQueue:   make(chan *RTCMMessage, queueSize),
		results:    make(chan interface{}, queueSize),
		ctx:        ctx,
		cancel:     cancel,
	}
	
	// Start workers
	pool.Start()
	
	return pool
}

// Start starts the worker pool
func (p *WorkerPool) Start() {
	for i := 0; i < p.numWorkers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// Stop stops the worker pool
func (p *WorkerPool) Stop() {
	p.cancel()
	close(p.jobQueue)
	p.wg.Wait()
	close(p.results)
}

// Submit submits a message for processing
func (p *WorkerPool) Submit(msg *RTCMMessage) {
	select {
	case p.jobQueue <- msg:
		// Message submitted successfully
	case <-p.ctx.Done():
		// Worker pool is shutting down
	}
}

// Results returns the results channel
func (p *WorkerPool) Results() <-chan interface{} {
	return p.results
}

// worker processes messages from the job queue
func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()
	
	for {
		select {
		case msg, ok := <-p.jobQueue:
			if !ok {
				// Job queue is closed
				return
			}
			
			// Process message
			result, err := DecodeRTCMMessage(msg)
			if err == nil {
				// Send result to results channel
				select {
				case p.results <- result:
					// Result sent successfully
				case <-p.ctx.Done():
					// Worker pool is shutting down
					return
				}
			}
			
		case <-p.ctx.Done():
			// Worker pool is shutting down
			return
		}
	}
}
