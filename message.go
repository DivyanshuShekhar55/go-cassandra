package main

import "time"

type Message struct {
	ID      string
	Payload string `json:"payload"`
	// author will be user-id and name
	// we are de-normalising maybe add profile pic too (Cassandra has de-normalisation)
	AuthorId   string    `json:"authorID"`
	AuthorName string    `json:"authorName"`
	Group      bool      `json:"isGroup"`
	GroupId    string    `json:"groupID,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}
