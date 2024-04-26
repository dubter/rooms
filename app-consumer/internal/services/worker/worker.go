package worker

import (
	"app-consumer/internal/domain"
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"time"
)

type PersistentStorage interface {
	PushMessage(ctx context.Context, msg *domain.Message) error
}

type CacheStorage interface {
	AddToList(ctx context.Context, msg *domain.Message) error
}

type Consumer interface {
	Consume(ctx context.Context, handler domain.MessageHandler) error
}

type Worker struct {
	logger            *slog.Logger
	consumer          Consumer
	persistentStorage PersistentStorage
	cache             CacheStorage
}

func New(logger *slog.Logger, consumer Consumer, persistentStorage PersistentStorage, cache CacheStorage) *Worker {
	return &Worker{
		logger:            logger,
		consumer:          consumer,
		persistentStorage: persistentStorage,
		cache:             cache,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	attempt := 0

	go func() {
		w.logger.Info("Worker is started")

		for {
			err := w.consumer.Consume(ctx, func(msg domain.Message) error {
				w.logger.Info("Consume message", slog.Any("msg", msg))

				err := w.persistentStorage.PushMessage(ctx, &msg)
				if err != nil {
					return fmt.Errorf("services.worker.Run: %w", err)
				}

				err = w.cache.AddToList(ctx, &msg)
				if err != nil {
					return fmt.Errorf("services.worker.Run: %w", err)
				}

				return nil
			})
			if err != nil {
				w.logger.Error("failed to consume message:", slog.String("error", err.Error()))

				time.Sleep(expBackoff(attempt))
				attempt++
			}
		}
	}()

	<-ctx.Done()

	return ctx.Err()
}

func expBackoff(attempt int) time.Duration {
	maxDelay := 30 * time.Second
	backoff := math.Pow(2, float64(attempt))
	delay := time.Duration(backoff) * time.Second
	if delay > maxDelay {
		delay = maxDelay
	}

	jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
	return delay + jitter
}
