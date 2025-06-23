package queue

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	MaxRequestsPerSecond = 7
	MaxEntitiesPerBatch  = 200
)

type QueueService struct {
	logger      *zap.Logger
	rateLimiter *RateLimiter
	batchQueue  chan *BatchRequest
	wg          sync.WaitGroup
}

type BatchRequest struct {
	Type     string
	Entities []interface{}
	Callback func(error)
}

type RateLimiter struct {
	tokens chan struct{}
	ticker *time.Ticker
	mu     sync.Mutex
}

func NewQueueService(logger *zap.Logger) *QueueService {
	qs := &QueueService{
		logger:      logger,
		rateLimiter: NewRateLimiter(MaxRequestsPerSecond),
		batchQueue:  make(chan *BatchRequest, 1000),
	}

	// Start processing goroutine
	go qs.processQueue()

	return qs
}

func NewRateLimiter(rps int) *RateLimiter {
	rl := &RateLimiter{
		tokens: make(chan struct{}, rps),
		ticker: time.NewTicker(time.Second / time.Duration(rps)),
	}

	// Fill initial tokens
	for i := 0; i < rps; i++ {
		rl.tokens <- struct{}{}
	}

	// Refill tokens
	go func() {
		for range rl.ticker.C {
			select {
			case rl.tokens <- struct{}{}:
			default:
			}
		}
	}()

	return rl
}

func (rl *RateLimiter) Wait() {
	<-rl.tokens
}

func (qs *QueueService) AddToQueue(ctx context.Context, requestType string, entities []interface{}) error {
	// Split entities into batches
	for i := 0; i < len(entities); i += MaxEntitiesPerBatch {
		end := i + MaxEntitiesPerBatch
		if end > len(entities) {
			end = len(entities)
		}

		batch := &BatchRequest{
			Type:     requestType,
			Entities: entities[i:end],
			Callback: func(err error) {
				if err != nil {
					qs.logger.Error("Batch processing failed", zap.Error(err))
				}
			},
		}

		select {
		case qs.batchQueue <- batch:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

func (qs *QueueService) processQueue() {
	for batch := range qs.batchQueue {
		qs.rateLimiter.Wait()

		qs.wg.Add(1)
		go func(b *BatchRequest) {
			defer qs.wg.Done()

			// Process batch
			err := qs.processBatch(b)
			if b.Callback != nil {
				b.Callback(err)
			}
		}(batch)
	}
}

func (qs *QueueService) processBatch(batch *BatchRequest) error {
	qs.logger.Info("Processing batch",
		zap.String("type", batch.Type),
		zap.Int("count", len(batch.Entities)))

	// TODO: Implement actual API call to AmoCRM
	return nil
}
