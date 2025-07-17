package main

import "github.com/gorilla/websocket"

type Client struct {
	conn *websocket.Conn

	manager Manager 

	chatroom string
}

type ClientList map[*Client]bool