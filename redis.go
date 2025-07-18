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

// for sending group messages keep a set in redis
// use key = group:GROUPID:server:SERVERID
// and corresponding value as user-id
// when user sends msg, check for the group id and send
// msg to all members on all servers with same group-id
func AddUserToGroupServer(groupId, serverId, userId string) error {
	key := fmt.Sprintf("group:%s:server:%s", groupId, serverId)
	err := RedisConn.SAdd(redisCtx, key, userId)
	if err != nil {
		// TODO : what else can be done here ??
		fmt.Println("couldn't add user to group-server")
		return err.Err()
	}
	return nil
}

func RemoveUserFromGroupServer(groupId, serverId, userId string) error {
	key := fmt.Sprintf("group:%s:server:%s", groupId, serverId)
	return RedisConn.SRem(redisCtx, key, userId).Err()
}

func GetGroupMembersOnServer(groupId, serverId string) ([]string, error) {
	key := fmt.Sprintf("group:%s:server:%s", groupId, serverId)

	// as the redis set docs put it ...
	// SMEMBERS is O(n) for set sizes in lakhs
	// look for better ways to handle this then
	return RedisConn.SMembers(redisCtx, key).Result()
}

// following fn is useful for routing purposes
// add to list of all servers where people belonging to a group are present
func AddActiveServerToGroup(groupId, serverId string) error {
	key := fmt.Sprintf("group:%s:activeServers", groupId)
	return RedisConn.SAdd(redisCtx, key, serverId).Err()
}

// get list of all servers where people belonging to a group are present
func GetActiveServersForGroup(groupId string) ([]string, error) {
	key := fmt.Sprintf("group:%s:activeServers", groupId)
	return RedisConn.SMembers(redisCtx, key).Result()
}
