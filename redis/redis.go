package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"strings"
)

type Redis struct {
	options *redis.Options
	client  *redis.Client
}

func New(url string) (*Redis, error) {
	options, err := redis.ParseURL(url)
	if err != nil {
		return nil, NewError(err, "unable to parse redis url")
	}

	r := redis.NewClient(options)

	return &Redis{
		options: options,
		client:  r,
	}, nil
}

func (r *Redis) GetMessageIdForState(state string) (int64, error) {
	val := r.client.Get(context.Background(),
		fmt.Sprintf("state-%s", strings.ToUpper(state)),
	)

	messageId, err := val.Int64()

	if err != nil {
		return 0, NewError(err, "Message ID did not convert to int64 or something")
	}

	return messageId, nil
}

func (r *Redis) SaveMessageIdForState(state string, messageId int64) error {
	err := r.client.Set(context.Background(),
		fmt.Sprintf("state-%s", strings.ToUpper(state)),
		messageId,
		0,
	).Err()

	if err != nil {
		return NewError(err, "Could not set value in redis")
	}

	return nil
}
