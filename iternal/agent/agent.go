package agent

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Tomap-Tomap/go-loyalty-service/iternal/client"
	"github.com/Tomap-Tomap/go-loyalty-service/iternal/logger"
	"github.com/Tomap-Tomap/go-loyalty-service/iternal/storage"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type Agent struct {
	s           *storage.Storage
	c           *client.Client
	getInterval uint
	workerLimit uint
}

func NewAgent(s *storage.Storage, c *client.Client, getInterval, workerLimit uint) Agent {
	return Agent{s, c, getInterval, workerLimit}
}

func (a *Agent) Run(ctx context.Context) error {
	jobs := make(chan func() error, a.workerLimit)
	defer close(jobs)

	for w := uint(1); w <= a.workerLimit; w++ {
		go worker(jobs)
	}

	for {
		select {
		case <-time.After(time.Duration(a.getInterval) * time.Second):
			err := a.processingOrders(ctx, jobs)

			if err != nil {
				return err
			}
		case <-ctx.Done():
			return nil
		}

	}
}

func (a *Agent) processingOrders(ctx context.Context, jobs chan<- func() error) error {
	logger.Log.Info("Get orders from db")
	numbers, err := a.s.GetNotProcessedOrders(ctx)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("get orders: %w", err)
	}

	var wg sync.WaitGroup
	for _, val := range numbers {
		jobs <- func() error {
			wg.Add(1)
			defer wg.Done()
			logger.Log.Info("Get order from service")
			o, err := a.c.GetOrder(ctx, val)

			if err != nil {
				return err
			}

			logger.Log.Info("Update order from service")
			err = a.s.UpdateOrder(ctx, *o)

			if err != nil {
				return err
			}

			return nil
		}
	}

	wg.Wait()
	return nil
}

func worker(jobs <-chan func() error) {
	for j := range jobs {
		err := j()

		if err != nil {
			logger.Log.Warn(
				"Error on sending to service",
				zap.Error(err),
			)
		}
	}
}
