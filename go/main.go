package main

import (
	"context"
	"net/http"
)

func main() {
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
