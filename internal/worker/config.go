package worker

import "github.com/redis/go-redis/v9"

func InitializeRedis() *redis.Client {
	rd := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	return rd
}
