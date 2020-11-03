package redis

import "github.com/go-redis/redis/v8"

type Redis struct {
	options *redis.Options
}

func New(url string) {
}
