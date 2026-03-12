// Package stream provides Redis-backed pub/sub for SSE delivery.
package stream

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Broadcaster publishes events to Redis channels and allows subscription.
type Broadcaster struct {
	redis *redis.Client
}

// NewBroadcaster creates a Broadcaster backed by the given Redis client.
func NewBroadcaster(r *redis.Client) *Broadcaster {
	return &Broadcaster{redis: r}
}

// channelKey returns the Redis pub/sub channel name for a given session.
func channelKey(sessionID string) string {
	return fmt.Sprintf("forge:session:%s", sessionID)
}

// Publish sends an event to all SSE subscribers for a session.
func (b *Broadcaster) Publish(ctx context.Context, sessionID, eventType string, data any) error {
	payload, err := json.Marshal(map[string]any{
		"type": eventType,
		"data": data,
	})
	if err != nil {
		return fmt.Errorf("stream: marshal event: %w", err)
	}
	return b.redis.Publish(ctx, channelKey(sessionID), payload).Err()
}

// Subscribe returns a channel that receives raw JSON event messages for the session.
// The caller must cancel ctx to unsubscribe and close the channel.
func (b *Broadcaster) Subscribe(ctx context.Context, sessionID string) <-chan string {
	ch := make(chan string, 64)
	sub := b.redis.Subscribe(ctx, channelKey(sessionID))

	go func() {
		defer close(ch)
		defer sub.Close()
		msgCh := sub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgCh:
				if !ok {
					return
				}
				ch <- msg.Payload
			}
		}
	}()

	return ch
}
