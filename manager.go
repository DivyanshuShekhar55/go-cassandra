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

// TODO : get server-id for the current server
var serverId = os.Getenv("SERVERID")

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

	for _, group := range groups {
		err := AddUserToGroupServer(group, serverId, client.userId)
		if err != nil {
			fmt.Println("Could not add user to Redis group-server:", err)
			// Optionally handle/retry/fail here
			continue // try for othe groups
		}

		//  Mark this server as active for the group for routing
		err = AddActiveServerToGroup(group, serverId)
		if err != nil {
			fmt.Println("Could not mark server as active for group:", err)
			continue // try for other groups
		}
	}

}

func (m *Manager) removeClient(client *Client) {
	// Lock state (assume m.clients and clients are protected concurrently)
	m.Lock()
	defer m.Unlock()

	// Only proceed if client exists
	if _, ok := m.clients[client]; ok {
		// 1. Clean up: close conn, with error handling
		err := client.conn.Close()
		if err != nil {
			fmt.Println("err closing connection:", err)
			// schedule for background retry/cleanup
		}

		// 2. Remove from in-memory maps
		delete(m.clients, client)
		delete(clients, client.userId)

		// 3. Fetch all groups to which this user belonged
		groups, err := model.FetchAllUserGroups(client.userId)
		if err != nil {
			fmt.Printf("error fetching groups of user %s for removing: %v\n", client.userId, err)
			// Optionally: retry/fallback, but safe to return
			return
		}

		for _, group := range groups {
			// 4. Remove user from group-server set in Redis
			err := RemoveUserFromGroupServer(group, serverId, client.userId)
			if err != nil {
				// TODO: add robust retry logic (exponential backoff or retry queue)
				fmt.Printf("error removing user %s from group-server (%s): %v\n", client.userId, group, err)
				// Continue attempting to remove from other groups
			}

			// 5. Check if any users left in this group on this server
			count, err := CheckRemainingGroupMembersOnServer(group, serverId)
			if err != nil {
				fmt.Printf("error checking online members for group %s: %v\n", group, err)
				// optionally: retry logic or notification
				continue
			}

			// 6. If no members left, remove this server from activeServers for group
			if count == 0 {
				err := RemoveActiveServerForGroup(group, serverId)
				if err != nil {
					// TODO: possible retry logic
					fmt.Printf("error removing server %s from activeServers for group %s: %v\n", serverId, group, err)
				}
			}
		}
	}
}

/* TODO : PUBSUB LISTENER FUNCTION
1. listen to "msg receive" events
2. once a msg comes check all clients from the same group
3. loop on all clients and send them the msg

*/

/* TODO : PUBSUB SEND EVENT
1. called when user clicks "send" event
2. publish msg to that group's pubsub
3. write to cassandra

*/
