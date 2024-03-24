package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"syscall"
	"time"

	"github.com/Tomap-Tomap/go-loyalty-service/iternal/logger"
	"github.com/Tomap-Tomap/go-loyalty-service/iternal/models"
	"github.com/Tomap-Tomap/go-loyalty-service/iternal/storage"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type Client struct {
	addr        string
	restyClient *resty.Client
	storage     *storage.Storage
}

func NewClient(s *storage.Storage, addr string) *Client {
	client := resty.New().
		AddRetryCondition(func(r *resty.Response, err error) bool {
			if r.StatusCode() == http.StatusTooManyRequests {
				retryAfterHeader := r.Header().Get("Retry-After")
				if retryAfterHeader != "" {
					retryAfter, err := time.ParseDuration(retryAfterHeader + "s")
					if err != nil {
						logger.Log.Warn("Parse Retry-After", zap.Error(err))
						return false
					}

					time.Sleep(retryAfter)
					return true
				}
			}
			return errors.Is(err, syscall.ECONNREFUSED)
		})
	return &Client{addr, client, s}
}

func (c *Client) GetOrder(ctx context.Context, number int64) (*models.Order, error) {
	numberStr := strconv.Itoa(int(number))

	req := c.restyClient.R().
		SetHeader("Content-Encoding", "gzip").
		SetContext(ctx)
	resp, err := req.Get("http://" + c.addr + "/api/orders/" + numberStr)

	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("request return %d status code", resp.StatusCode())
	}

	var order models.Order

	err = json.Unmarshal(resp.Body(), &order)

	if err != nil {
		return nil, fmt.Errorf("unmarsall body: %w", err)
	}

	return &order, nil
}
