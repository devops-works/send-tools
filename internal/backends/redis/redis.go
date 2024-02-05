package redis

import (
	"context"
	"fmt"

	rd "github.com/redis/go-redis/v9"
)

type Redis struct {
	client *rd.Client
	Server string
}

func NewRedis(addr string) *Redis {
	rdb := rd.NewClient(&rd.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return &Redis{
		client: rdb,
		Server: addr,
	}
}

func (r *Redis) ListObjects(ctx context.Context) ([]string, error) {
	// iterate over all keys
	iter := r.client.ScanType(ctx, 0, "", 0, "hash").Iterator()
	var keys []string
	for iter.Next(ctx) {
		// hgetall key
		val, err := r.client.HGetAll(ctx, iter.Val()).Result()
		if err != nil {
			// logger.Error("unable to get key", "key", iter.Val(), "error", err)
			continue
		}
		keys = append(keys, val["prefix"]+"-"+iter.Val())
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return keys, nil
}

func (r *Redis) RemoveObject(ctx context.Context, key string) error {
	// remove key from redis
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("unable to remove key %s from redis: %w", key, err)
	}

	return nil
}
