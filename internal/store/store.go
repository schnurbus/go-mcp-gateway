package store

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	OAuthAccessTokenTTL    = 60 * time.Minute
	OAuthRefreshTokenTTL   = 30 * 24 * time.Hour
	OAuthStateTTL          = 5 * time.Minute
	OAuthClientTTL         = 90 * 24 * time.Hour
	SessionTTL             = 7 * 24 * time.Hour
	ResourceAccessTokenTTL = 30 * 24 * time.Hour
)

type Store struct {
	rdb    *redis.Client
	prefix string
	ttl    time.Duration
}

func NewStore(rdb *redis.Client, prefix string, ttl time.Duration) *Store {
	return &Store{
		rdb:    rdb,
		prefix: prefix + ":",
		ttl:    ttl,
	}
}

func (s *Store) Get(ctx context.Context, key string) (string, error) {
	return s.rdb.Get(ctx, s.prefix+key).Result()
}

func (s *Store) GetDel(ctx context.Context, key string) (string, error) {
	return s.rdb.GetDel(ctx, s.prefix+key).Result()
}

func (s *Store) Set(ctx context.Context, key string, value any) error {
	return s.rdb.Set(ctx, s.prefix+key, value, s.ttl).Err()
}

func (s *Store) Del(ctx context.Context, key string) error {
	return s.rdb.Del(ctx, s.prefix+key).Err()
}
