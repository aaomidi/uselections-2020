package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

type Redis struct {
	options *redis.Options
	client  *redis.Client
	log     *log.Entry
}

func New(url string) (*Redis, error) {
	options, err := redis.ParseURL(url)
	if err != nil {
		return nil, NewError(err, "unable to parse redis url")
	}

	r := redis.NewClient(options)

	self := &Redis{
		options: options,
		client:  r,
		log:     log.WithField("source", "redis"),
	}

	self.log.Infof("redis connected to %s", url)

	return self, nil
}

func (r *Redis) GetMessageIdForState(channelId int64, state string) (int64, error) {
	val := r.client.Get(context.Background(),
		fmt.Sprintf("state-%d-%s", channelId, strings.ToUpper(state)),
	)

	messageId, err := val.Int64()

	if err != nil {
		return 0, NewError(err, "Message ID did not convert to int64 or something")
	}

	return messageId, nil
}

func (r *Redis) SaveMessageIdForState(channelId int64, state string, messageId int64) error {
	err := r.client.Set(context.Background(),
		fmt.Sprintf("state-%d-%s", channelId, strings.ToUpper(state)),
		messageId,
		0,
	).Err()

	if err != nil {
		return NewError(err, "Could not set value in redis")
	}

	return nil
}

func (r *Redis) SaveInlineMessageId(state string, inlineMessageId string) error {
	err := r.client.RPush(context.Background(),
		fmt.Sprintf("inline-state-%s", strings.ToUpper(state)),
		inlineMessageId,
	).Err()

	if err != nil {
		return NewError(err, "Could not save a new inline message")
	}

	return nil
}

func (r *Redis) GetInlineMessageId(state string) ([]int64, error) {
	result := r.client.LRange(context.Background(),
		fmt.Sprintf("inline-state-%s", strings.ToUpper(state)),
		0, -1,
	)

	if result.Err() != nil {
		return nil, NewError(result.Err(), "Could not save a new inline message")
	}

	result.Val()
	finalResult := make([]int64, 0, len(result.Val()))
	for _, r := range result.Val() {
		v, err := strconv.ParseInt(r, 10, 64)
		if err != nil {
			return nil, NewError(err, "Error when parsing one of the values from results")
		}
		finalResult = append(finalResult, v)
	}
	return finalResult, nil
}
