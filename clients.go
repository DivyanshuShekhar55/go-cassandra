package main

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

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

		var r Message
		if err := json.Unmarshal(msg, &r); err != nil {
			log.Println("error unmarshalling read msg", err.Error())
			continue
		}

		// TODO : validate the message 'r'

		// next if the msg was sent in a group send to all members
		if r.Group {
			// send to db
			// fetch members of group
			// send to members
			
		}

	}
}
