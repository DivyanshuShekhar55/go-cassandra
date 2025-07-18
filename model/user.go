package model

import (
	"fmt"

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

	query := `SELECT group_id FROM user_groups WHERE user_id = ?`
	result, err := db.ExecuteIterableQuery(query, uuid)
	if err != nil {
		fmt.Printf("Error getting user groups")
		// also may need to retry or panic maybe
		// because it's a big error
		return nil, err
	}

	// result is of type any need to convert into uuid
	groups = make([]string, len(result))
	for i, v := range result {
		// If your ExecuteIterableQuery uses string as item, this is safe.
		// If it returns UUID types, use v.(gocql.UUID).String()
		groups[i] = v.(string)
	}
	return groups, nil
}
