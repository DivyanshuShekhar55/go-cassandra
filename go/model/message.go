package model

import (
	"fmt"
	"time"

	db "github.com/DivyanshuShekhar55/go-cassandra.git/database"
	"github.com/gocql/gocql"
)

type Message struct {
	GroupID    string     `json:"group_id,omitempty"`
	MsgID      gocql.UUID `json:"msg_id,omitempty"`
	SenderID   string     `json:"sender_id"`
	SenderName string     `json:"sender_name,omitempty"`
	Receiver   string     `json:"receiver,omitempty"` // empty if group msg
	Content    string     `json:"content"`
	Timestamp  time.Time  `json:"timestamp,omitempty"`
	Bucket     string     `json:"bucket,omitempty"`
	Group      bool       `json:"group"`
	//SenderAvatar string
}

func SaveMessageGroupChat(message Message) error {
	query := `INSERT INTO group_chat( group_id, bucket, msg_id, sender_id, sender_name, content, timestamp) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`

	err := db.ExecuteQuery(query, message.GroupID, message.Bucket, message.MsgID, message.SenderID, message.SenderName, message.Content, message.Timestamp)

	if err != nil {
		fmt.Println("error", err)
		return err
	}
	return nil
}
