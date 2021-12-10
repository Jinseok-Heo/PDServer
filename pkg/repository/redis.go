package repository

import "github.com/go-redis/redis"

func NewClient() (*redis.Client, error) {
	var client *redis.Client

	client = redis.NewClient(&redis.Options{
		// Addr: "redis:6379", //for docker
		Addr: "localhost:6379",
	})
	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}
	return client, nil
}

func CloseRedisDB(client *redis.Client) error {
	if client != nil {
		return client.Close()
	}
	return nil
}
