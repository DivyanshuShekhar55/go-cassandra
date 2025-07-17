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
		CheckOrigin : checkOrigin,
		WriteBufferSize: 1024,
		ReadBufferSize: 1024,
	}
)


func checkOrigin (r *http.Request) bool {
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

func NewManager(ctx context.Context) *Manager{
	m := &Manager{
		clients:  make(ClientList),
	}
	return m
}

func (m* Manager) serverWS(w http.ResponseWriter, r *http.Request){
	log.Println("New Connection")
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(conn)
}

