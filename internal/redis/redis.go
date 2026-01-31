// Package redis provides Redis client wrapper for pub/sub and caching.
package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// Client wraps the Redis client with convenience methods
type Client struct {
	rdb *redis.Client
}

// New creates a new Redis client
func New(redisURL string) (*Client, error) {
	if redisURL == "" {
		return nil, nil
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Client{rdb: rdb}, nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	if c.rdb != nil {
		return c.rdb.Close()
	}
	return nil
}

// IsActive returns true if Redis is connected
func (c *Client) IsActive() bool {
	return c.rdb != nil
}

// Set stores a value with TTL
func (c *Client) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

// Get retrieves a value
func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

// Delete removes a key
func (c *Client) Delete(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, key).Err()
}

// Publish publishes a message to a channel
func (c *Client) Publish(ctx context.Context, channel string, message []byte) error {
	return c.rdb.Publish(ctx, channel, message).Err()
}

// Subscribe subscribes to a channel and returns a channel for messages
func (c *Client) Subscribe(ctx context.Context, channel string) (<-chan []byte, func()) {
	pubsub := c.rdb.Subscribe(ctx, channel)
	msgChan := make(chan []byte, 100)

	go func() {
		defer close(msgChan)
		ch := pubsub.Channel()
		for msg := range ch {
			select {
			case msgChan <- []byte(msg.Payload):
			case <-ctx.Done():
				return
			}
		}
	}()

	cleanup := func() {
		_ = pubsub.Close()
	}

	return msgChan, cleanup
}

// SetTicket stores a WebSocket ticket with 10s TTL
func (c *Client) SetTicket(ctx context.Context, ticket string, userID uint) error {
	return c.Set(ctx, "ws:ticket:"+ticket, []byte{byte(userID)}, 10*time.Second)
}

// GetAndDeleteTicket retrieves and deletes a ticket (one-time use)
func (c *Client) GetAndDeleteTicket(ctx context.Context, ticket string) (uint, error) {
	key := "ws:ticket:" + ticket
	val, err := c.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	if val == nil {
		return 0, nil
	}

	// Delete ticket (one-time use)
	_ = c.Delete(ctx, key)

	return uint(val[0]), nil
}

