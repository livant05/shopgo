package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct{ rdb *redis.Client }

func New(addr, password string) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("conectar Redis: %w", err)
	}
	return &Client{rdb: rdb}, nil
}

func (c *Client) Ping(ctx context.Context) error { return c.rdb.Ping(ctx).Err() }

func (c *Client) Get(ctx context.Context, key string, dest any) error {
	data, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil { return err }
	return json.Unmarshal(data, dest)
}

func (c *Client) Set(ctx context.Context, key string, val any, ttlSec int) error {
	data, err := json.Marshal(val)
	if err != nil { return err }
	return c.rdb.Set(ctx, key, data, time.Duration(ttlSec)*time.Second).Err()
}

func (c *Client) Delete(ctx context.Context, key string) error { return c.rdb.Del(ctx, key).Err() }

func (c *Client) Invalidate(ctx context.Context, pattern string) error {
	iter := c.rdb.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) { c.rdb.Del(ctx, iter.Val()) }
	return iter.Err()
}

func (c *Client) Publish(ctx context.Context, ch string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil { return err }
	return c.rdb.Publish(ctx, ch, data).Err()
}

func (c *Client) Subscribe(ctx context.Context, ch string, fn func([]byte)) error {
	sub := c.rdb.Subscribe(ctx, ch)
	defer sub.Close()
	for {
		select {
		case msg := <-sub.Channel():
			go fn([]byte(msg.Payload))
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
