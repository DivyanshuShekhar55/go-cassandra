package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func GetBucketForTime(t time.Time) string {
	year, week := t.ISOWeek()
	return fmt.Sprintf("%d-W%02d", year, week)
}

type GroupListPayload struct {
	Type   string   `json:"type"`
	Groups []string `json:"groups"`
}

// ideally the Client struct should be importable
// and should be passed here instead of ws conn
// but meh :-) will do it later
func SendUserGroupListToUser(conn *websocket.Conn, groupList []string) error {
	if len(groupList) == 0 {
		// no need to write, maybe some error
		// but will be handled in the fetchUserGroup fn only
		return nil
	}

	// create the payload struct
	payload := GroupListPayload{
		Type:   "groups",
		Groups: groupList,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("error in marshalling user group")
		return err
	}

	// send the message to frontend
	err = conn.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		log.Printf("error in sending use groups for user")
		return err
	}

	return nil

}
