package repo

import (
	"LTICore/internal/config"
	"LTICore/internal/core/domain"
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
)

type RedisLoginSessionRepo struct {
	client *redis.Client
	cfg    *config.RedisConfig
}

func NewRedisClient(cfg *config.RedisConfig) *RedisLoginSessionRepo {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.MaxSize,
	})
	return &RedisLoginSessionRepo{client, cfg}
}
func (c *RedisLoginSessionRepo) GetAndDelete(ctx context.Context, state string) (*domain.LoginSession, error) {
	key := c.cfg.LoginSessionKey + state

	data, err := c.client.GetDel(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var session domain.LoginSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	return &session, nil
}
func (c *RedisLoginSessionRepo) Save(ctx context.Context, session *domain.LoginSession) error {

	key := c.cfg.LoginSessionKey + session.State

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, c.cfg.TTL).Err()
}
