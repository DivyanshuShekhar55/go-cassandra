package main

import "time"

type Message struct {
	Payload string `json:"payload"`
	// author will be user-id and name
	// we are de-normalising maybe add profile pic too (Cassandra has de-normalisation)
	AuthorId   string `json:"authorID"`
	AuthorName string `json:"authorName"`
	GroupId    string `json:"groupID"`
	Timestamp  time.Time `json:"timestamp"` 
}