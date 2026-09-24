package presence

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// PresenceTTL is how long a user is considered online after last heartbeat
	PresenceTTL = 5 * time.Minute
	// HeartbeatInterval is how often clients should send heartbeats
	HeartbeatInterval = 1 * time.Minute
)

// Presence manages online status using Redis with TTL
type Presence struct {
	client *redis.Client
}

// NewPresence creates a new Presence instance
func NewPresence(client *redis.Client) *Presence {
	return &Presence{
		client: client,
	}
}

// SetOnline marks a user as online
func (p *Presence) SetOnline(ctx context.Context, username string) error {
	key := p.userKey(username)
	return p.client.Set(ctx, key, "1", PresenceTTL).Err()
}

// SetOffline marks a user as offline
func (p *Presence) SetOffline(ctx context.Context, username string) error {
	key := p.userKey(username)
	return p.client.Del(ctx, key).Err()
}

// IsOnline checks if a user is online
func (p *Presence) IsOnline(ctx context.Context, username string) (bool, error) {
	key := p.userKey(username)
	exists, err := p.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// GetOnlineCount returns the number of online users
func (p *Presence) GetOnlineCount(ctx context.Context) (int64, error) {
	pattern := "presence:user:*"
	iter := p.client.Scan(ctx, 0, pattern, 0).Iterator()
	
	count := int64(0)
	for iter.Next(ctx) {
		count++
	}
	if err := iter.Err(); err != nil {
		return 0, err
	}
	return count, nil
}

// Heartbeat refreshes the online status TTL
func (p *Presence) Heartbeat(ctx context.Context, username string) error {
	return p.SetOnline(ctx, username)
}

func (p *Presence) userKey(username string) string {
	return "presence:user:" + username
}
