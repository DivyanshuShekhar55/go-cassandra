package main

import (
	"context"
	"log"
	"net/http"
	"os"

	db "github.com/DivyanshuShekhar55/go-cassandra.git/database"
	"github.com/google/uuid"
)

func main() {
	db.SetupDBConnection()
	NewRedisConnPool()

	// generate a random id for current server
	serverID := uuid.New().String()
	os.Setenv("SERVERID", serverID)
	log.Println("SERVER ID:", serverID)

	rootCtx := context.Background()
	ctx, cancel := context.WithCancel(rootCtx)

	defer cancel()

	manager := NewManager(ctx)

	// run the listen channel which keeps running
	// listens to any group message and sends the message
	// to online members of the group on that particular server
	go ListenToChannel(broadcast, manager)

	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/ws", manager.serverWS)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}

}
