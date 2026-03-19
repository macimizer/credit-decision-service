package redispkg

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	client *redis.Client
}

func New(addr, password string, db int) *Client {
	return &Client{
		client: redis.NewClient(&redis.Options{
			Addr:         addr,
			Password:     password,
			DB:           db,
			PoolSize:     20,
			MinIdleConns: 5,
			ReadTimeout:  2 * time.Second,
			WriteTimeout: 2 * time.Second,
		}),
	}
}

func (c *Client) Native() *redis.Client {
	return c.client
}

func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *Client) Get(ctx context.Context, key string, destination interface{}) (bool, error) {
	payload, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if err := json.Unmarshal(payload, destination); err != nil {
		return false, err
	}

	return true, nil
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, payload, ttl).Err()
}

func (c *Client) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
