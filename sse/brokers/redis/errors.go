package redis

import "errors"

var (
	// ErrNoRedisClient is returned when attempting to create a broker without a Redis client
	ErrNoRedisClient = errors.New("no Redis client provided")

	// ErrRedisConnectionFailed is returned when the broker cannot connect to Redis
	ErrRedisConnectionFailed = errors.New("failed to connect to Redis")
)
