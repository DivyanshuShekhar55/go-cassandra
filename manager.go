package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/DivyanshuShekhar55/go-cassandra.git/model"
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

	m.clients[client] = true
	clients[client.userId] = client.conn

	groups, err := model.FetchAllUserGroups(client.userId)

	if err != nil {
		fmt.Println("error getting user groups")
		// do something cause its an issue
		// maybe, handle (retry or disconnect the client, etc)
		return
	}

	// TODO : get server-id for the current server
	serverId := os.Getenv("SERVERID")

	for _, group := range groups {
		err := AddUserToGroupServer(group, serverId, client.userId)
		if err != nil {
			fmt.Println("Could not add user to Redis group-server:", err)
			// Optionally handle/retry/fail here
		}

		//  Mark this server as active for the group for routing
		err = AddActiveServerToGroup(group, serverId)
		if err != nil {
			fmt.Println("Could not mark server as active for group:", err)
		}
	}

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

	// TODO : whenever user disconnects remove from group-server on redis also
	// also if the user is last online member of this group on the server
	// then remove the server from active group server
}
