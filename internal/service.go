package internal

import (
	"github.com/go-redis/redis/v7"
	log "github.com/micro/go-micro/v2/logger"
	"tpush/options"
)

func NewCache(options options.RedisOptions) (ret *redis.Client) {
	ret = redis.NewClient(
		&redis.Options{
			Addr:            options.Address,
			Password:        options.Password, // no password set
			DB:              0,                // use default DB
			MaxRetries:      1,
			MinRetryBackoff: 1,
			MaxRetryBackoff: 64,
		},
	)
	_, err := ret.Ping().Result()
	if err != nil {
		log.Fatalf("cache.Ping() err, %v", err)
	}
	return ret
}
