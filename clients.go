package main

import "github.com/gorilla/websocket"

type Client struct {
	conn *websocket.Conn

	manager *Manager 

}

type ClientList map[*Client]bool

func NewClient(conn *websocket.Conn, m *Manager) *Client{
	return &Client{
		conn:conn,
		manager:m,
	}
}



