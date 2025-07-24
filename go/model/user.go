package model

import (
	"fmt"
	"log"

	db "github.com/DivyanshuShekhar55/go-cassandra.git/database"
	"github.com/gocql/gocql"
)

// TODO : func to add a user to a group
func AddUserToGroup() {}

// TODO : remove from group
func RemoveUserFromGroup() {}

// TODO : fetch all members of a group
func FetchAllGroupMembers() {}

// fetch all groups to which a user-id has joined
func FetchAllUserGroups(userId string) (groups []string, err error) {

	uuid, err := gocql.ParseUUID(userId)
	if err != nil {
		fmt.Println("Invalid user id")
		return nil, err
	}

	log.Println("here coming")

	query := `SELECT group_id FROM user_groups WHERE user_id = ?`

	// do an iteration with an iterator
	iter := db.Connection.Session.Query(query, uuid).Iter()

	var groupID gocql.UUID

	for iter.Scan(&groupID) {
		groups = append(groups, groupID.String())
	}

	if err := iter.Close(); err != nil {
		log.Println("Error closing iterator:", err)
		return nil, err
	}

	return groups, nil
}
