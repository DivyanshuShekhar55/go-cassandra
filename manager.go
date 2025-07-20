package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/DivyanshuShekhar55/go-cassandra.git/model"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
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
	clients                                ClientList
	UnsubscribeServerFromGroupChannelFuncs map[string]func()
	sync.RWMutex
}

func NewManager(ctx context.Context) *Manager {
	m := &Manager{
		clients:                                make(ClientList),
		UnsubscribeServerFromGroupChannelFuncs: make(map[string]func()),
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
	go client.receiveMessage()
}

func (m *Manager) addClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	m.clients[client.userId] = client
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
		// if not already added
		count, err := CheckRemainingGroupMembersOnServer(group, serverId)
		if err != nil {
			// retry
			fmt.Println("Coudn't fetch count of group members on server", serverId)
		}
		if count == 1 {
			err = AddActiveServerToGroup(group, serverId)
			if err != nil {
				fmt.Println("Could not mark server as active for group:", err)
				continue // try for other groups
			}

			// add the group to pubsub
			// so that server can listen to incoming messages
			// add the cancel func returned to map of unsubscribe funcs
			// this map is used when remoing server from pubsub in removeClient func
			groupChannelKey := fmt.Sprintf("group:%s:messages", group)
			cancelFunc, err := SubscribeToGroup(groupChannelKey, broadcast)

			if err != nil {
				// do something
				continue
			}

			m.UnsubscribeServerFromGroupChannelFuncs[groupChannelKey] = cancelFunc
		}
	}
}

func (m *Manager) removeClient(client *Client) {
	// Lock state (assume m.clients and clients are protected concurrently)
	m.Lock()
	defer m.Unlock()

	// Only proceed if client exists
	if _, ok := m.clients[client.userId]; ok {
		// 1. Clean up: close conn, with error handling
		err := client.conn.Close()
		if err != nil {
			fmt.Println("err closing connection:", err)
			// schedule for background retry/cleanup
		}

		// 2. Remove from in-memory maps
		delete(m.clients, client.userId)
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

				// unsubscribe from pubsub
				// so that server doesn't listen to any messages for this group

				groupChannelKey := fmt.Sprintf("group:%s:messages", group)
				cancelFunc, ok := m.UnsubscribeServerFromGroupChannelFuncs[groupChannelKey]
				if !ok {
					fmt.Println("already unsubscribed from server")
					continue
				}

				cancelFunc()
				delete(m.UnsubscribeServerFromGroupChannelFuncs, groupChannelKey)

			}
		}
	}
}

/*
	TODO : PUBSUB LISTENER FUNCTION

1. listen to "msg receive" events
2. once a msg comes check all clients from the same group
3. loop on all clients and send them the msg
*/
func ListenToChannel(broadcast <-chan *redis.Message, m *Manager) {

	for msg := range broadcast { // This loop runs forever

		// extract group-id +e.g. "group:123:messages" => "123"
		groupId := extractGroupIdFromChannel(msg.Channel)

		// acquire lock on the clients map
		// map every group user and send him the message
		m.RLock()
		m.FanOutMessage(msg.Payload, groupId)
		m.RUnlock()
	}
}

func extractGroupIdFromChannel(groupChannelKey string) string {
	parts := strings.Split(groupChannelKey, ":")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

func (m *Manager) FanOutMessage(payload string, groupId string) error {
	clientList, err := GetGroupMembersOnServer(groupId, serverId)
	if err != nil {
		log.Printf("unable to get group members for group %s: %v", groupId, err)
		return err
	}

	for _, userId := range clientList {
		clientObj := m.clients[userId]
		if clientObj == nil {
			// Maybe client disconnected; skip or log if desired
			continue
		}
		err := clientObj.conn.WriteMessage(websocket.TextMessage, []byte(payload))
		if err != nil {
			log.Printf("error sending message to user %s: %v", userId, err)
			// Optionally, close and remove the client if err is serious
			// clientObj.manager.removeClient(clientObj)
			// Or schedule for retry/cleanup
		}
	}
	return nil
}
