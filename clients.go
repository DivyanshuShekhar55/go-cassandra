package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/DivyanshuShekhar55/go-cassandra.git/model"
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
// TODO : also make a map for server id and group members connected to the server
var clients = make(map[string]*websocket.Conn)

type Client struct {
	conn    *websocket.Conn
	userId  string
	manager *Manager
}

// map userId to connection
type ClientList map[*Client]bool

func NewClient(conn *websocket.Conn, m *Manager, userId string) *Client {
	return &Client{
		conn:    conn,
		manager: m,
		userId:  userId,
	}
}

func (client *Client) receiveMessage() {

	defer func() {
		// gracefully shutdown the connection
		// once the func is done (user disconnects, phone switches off, etc)
		client.manager.removeClient(client)
	}()

	// set max size of the connection in bytes
	// calculate the message size of content user can send
	client.conn.SetReadLimit(512)

	// TODO : add pong response time deadline
	// see percy's client.go

	// loop forever
	for {
		_, msg, errConn := client.conn.ReadMessage()

		if errConn != nil {
			log.Println("read err", errConn)
			return
		}

		var r model.Message
		if err := json.Unmarshal(msg, &r); err != nil {
			log.Println("error unmarshalling read msg", err.Error())
			continue
		}

		// TODO : validate the message 'r'

		// next if the msg was sent in a group send to all members
		if r.Group {
			// TODO :
		}

	}
}

func (client *Client) Send(msg string) {

	// get msg from frontend via json
	// called only when user clicks "send", no braodcasts here
	//msg := <-broadcast
	message := model.Message{}
	if err := json.Unmarshal([]byte(msg), &message); err != nil {
		fmt.Println("message type not correct")
		// should we panic ??
		panic(err)
	}

	if message.Group {
		groupMessage(message)
		return // skip rest of body for this msg
	}

	// private message
	receiver := clients[message.Receiver]
	if client == nil {
		fmt.Println("receiver offline")
		return
	}

	privateMessage(message, receiver)

}

func groupMessage(message model.Message) {
	// TODO :
	/*
		1. call the manager's (ie., the server's) send event to that group topic
	*/

}

// TODO :
// create a listener on server that when group msg comes
// get all the users on the server of that group and send them message

// TODO : private message chat feature
func privateMessage(message model.Message, client *websocket.Conn) {}
