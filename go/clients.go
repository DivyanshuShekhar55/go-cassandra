package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	
	"github.com/DivyanshuShekhar55/go-cassandra.git/model"
	"github.com/DivyanshuShekhar55/go-cassandra.git/utils"
	"github.com/gocql/gocql"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// this broadcast or as percy puts 'egress'
// will hold multiple messages which are to be sent
// remember that gorilla allows for only one write goroutine at a time, so need to store all the remaining messages
// think from the view of a spammer, per second maybe 10 messages, single goroutine can't handle
var broadcast = make(chan *redis.Message)

// 'clients' map user-id to ws connection
// TODO : what is the ClientList map then
var clients = make(map[string]*websocket.Conn)

type Client struct {
	conn    *websocket.Conn
	userId  string
	manager *Manager
}

// map userId to connection
type ClientList map[string]*Client

func NewClient(conn *websocket.Conn, m *Manager, userId string) *Client {
	return &Client{
		conn:    conn,
		manager: m,
		userId:  userId,
	}
}

// listen to messages from frontend from user
// do some validations
// send the message for further processing
func (client *Client) receiveMessage() {

	defer func() {
		// gracefully shutdown the connection
		// once the func is done (user disconnects, phone switches off, etc)
		client.manager.removeClient(client)
	}()

	// set max size of the connection in bytes
	// calculate the message size of content user can send
	client.conn.SetReadLimit(1024 * 2) // 2Mb

	// TODO : add pong response time deadline
	// see percy's client.go

	// loop forever
	for {
		_, msg, errConn := client.conn.ReadMessage()

		if errConn != nil {
			log.Printf("ws read err for user %s : %s", client.userId, errConn)
			return
		}

		// Optional: Validate message size again at application level
		if len(msg) == 0 || len(msg) > 2048 {
			log.Printf("Invalid message size from user %s\n", client.userId)
			continue
		}

		client.Send(string(msg))

	}
}

// logic for processing sending of messages by a user
func (client *Client) Send(msg string) {

	message := model.Message{}
	if err := json.Unmarshal([]byte(msg), &message); err != nil {
		fmt.Println("message format not correct")
		return
	}

	// fill in remaining fields of msg struct
	message.Bucket = utils.GetBucketForTime(time.Now().UTC())
	message.MsgID = gocql.TimeUUID()
	message.Timestamp = time.Now().UTC()

	if message.Group {
		err := groupMessage(message)

		if err != nil {
			fmt.Println("error sending message to group")
		}
		return // skip rest of body for this msg
	}

	// private message
	receiver := clients[message.Receiver]
	if client == nil {
		fmt.Println("receiver offline")
		return
	}

	privateMessage(message, receiver)

	/* TODO :
	1. take care of same user online from multiple devices
	2. check for rate-limiting and message size
	3. look more on cassandra batching
	*/

}

func groupMessage(message model.Message) error {
	// 1. persist to cassandra
	err := model.SaveMessageGroupChat(message)
	if err != nil {
		fmt.Println("error while writing to db")
		return err
	}

	// 2. write to pub sub channel made for that group
	// maybe i should handle this group to groupTopic key conversion and marshal
	// in a diff func for cleanliness
	groupTopic := fmt.Sprintf("group:%s:messages", message.GroupID)
	payload, err := json.Marshal(message)

	if err != nil {
		fmt.Println("error parsing group message")
		return err
	}

	err = PublishGroupMessage(groupTopic, payload)

	if err != nil {
		fmt.Println("error publishing to pubsub")
	}

	return nil

}

// TODO : private message chat feature
func privateMessage(message model.Message, client *websocket.Conn) {}
