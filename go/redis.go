package main

import (
	"context"
	"fmt"
	"os"
	//"time"

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
	err := RedisConn.SAdd(redisCtx, key, userId).Err()
	if err != nil {
		// TODO : what else can be done here ??
		fmt.Println("couldn't add user to group-server", err)
		return err
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

// remove an active server from list
// use when no more members are active on the server for that group
func RemoveActiveServerForGroup(groupId, serverId string) error {
	key := fmt.Sprintf("group:%s:activeServers", groupId)
	return RedisConn.SRem(redisCtx, key, serverId).Err()
}

// useful during removing server from active servers for group
func CheckRemainingGroupMembersOnServer(groupId, serverId string) (count int64, err error) {
	key := fmt.Sprintf("group:%s:server:%s", groupId, serverId)
	count, err = RedisConn.SCard(redisCtx, key).Result()

	if err != nil {
		return -1, err // returns -1 for error
	}

	return count, nil
}

// func to listen to if any group message comes
// the server subscribes to the group on pubsub and listens for messages
func SubscribeToGroup(groupChannelKey string, broadcast chan<- *redis.Message) (func(), error) {

	subscriber := RedisConn.Subscribe(redisCtx, groupChannelKey)
	// Optionally: Use WaitGroup to coordinate shutdown/cleanup if many goroutines
	quit := make(chan struct{})
	// Run a forever listener for this subscription on a goroutine
	go func() {
		for {
			select {
			case <-quit:
				subscriber.Close()
				return
			default:
				msg, err := subscriber.ReceiveMessage(redisCtx)
				if err != nil {
					fmt.Println("Error receiving message from Redis:", err)
					//time.Sleep(time.Second) // Backoff before retry
					continue
				}
				// Deliver the message to broadcast channel for processing/fanout
				// the listener on broadcast channel will send the
				// msg to all users on this listener server
				broadcast <- msg
			}
		}
	}()
	// Return a cleanup function to unsubscribe and quit
	return func() { close(quit) }, nil
}

// func to publish group messages to redis pubsub
func PublishGroupMessage(group string, payload []byte) error{
	
	// returns count of subscribers, meh i don't need that
	_, err:= RedisConn.Publish(redisCtx, group, payload).Result()

	if err != nil {
		return err
	}
	return nil
}
