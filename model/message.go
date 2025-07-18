package model

import (
	"fmt"
	"time"

	db "github.com/DivyanshuShekhar55/go-cassandra.git/database"
	"github.com/gocql/gocql"
)

type Message struct {
	GroupID    string
	MsgID      gocql.UUID
	SenderID   string
	SenderName string
	//SenderAvatar string
	Receiver string `json:"receiver,omitempty"` // empty if group msg
	Content   string
	Timestamp time.Time
	Bucket    string
	Group     bool
}

func SaveMessageGroupChat(message Message) {
	query := `INSERT INTO group_chat( group_id, bucket, msg_id, sender_id, sender_name, content, timestamp) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`

	err := db.ExecuteQuery(query, message.GroupID, message.Bucket, message.MsgID, message.SenderID, message.SenderName, message.Content, message.Timestamp)

	if err != nil {
		fmt.Println("error", err)
	}
}
