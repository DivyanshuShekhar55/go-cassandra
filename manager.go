package main

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	wsUpgrader = websocket.Upgrader{
		CheckOrigin:     checkOrigin,
		WriteBufferSize: 1024,
		ReadBufferSize:  1024,
	}
)

func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")

	switch origin {
	case "http://localhost:8000":
		return true
	default:
		return false
	}
}

type Manager struct {
	clients ClientList

	sync.RWMutex
}

func NewManager(ctx context.Context) *Manager {
	m := &Manager{
		clients: make(ClientList),
	}
	return m
}

func (m *Manager) serverWS(w http.ResponseWriter, r *http.Request) {

	log.Println("New Connection")

	userId := r.URL.Query().Get("id")

	if userId == "" {
		http.Error(w, "userID not provided", http.StatusBadRequest)
		return
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := NewClient(conn, m, userId)
	m.addClient(client)
}

func (m *Manager) addClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	m.clients[client]=true
}

func (m *Manager) removeClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	// check if client exists, if yes delete
	if _, ok := m.clients[client]; ok {
		//clean-up with closing connection
		client.conn.Close()

		// remove from active list
		delete(m.clients, client)
	}
}

