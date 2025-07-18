package main

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

var redisCtx = context.Background()
var RedisConn redis.Client

func NewRedisConnPool() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0, // use default db 0-th index/number
	})

	RedisConn = *rdb
}

func PubSub() {
	SERVERID := os.Getenv("SERVERID")

	// TODO : find out what the server id here refers to

	subscriber := RedisConn.Subscribe(redisCtx, SERVERID)

	for {
		msg, err := subscriber.ReceiveMessage(redisCtx)

		if err != nil {
			fmt.Println("Couldnt connect to redis")
			panic(err)
		}

		// fmt.Printf("message from pub/sub : %v", msg.Payload)
		broadcast <- msg

	}
}
