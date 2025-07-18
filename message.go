package main

import (
	"time"

	"github.com/gocql/gocql"
)

type Message struct {
    GroupID      string
    MsgID        gocql.UUID
    SenderID     string
    SenderName   string
    //SenderAvatar string
    Content      string
    Timestamp    time.Time
    Bucket       string
}
