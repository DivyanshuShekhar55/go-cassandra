package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/gocql/gocql"
    "github.com/google/uuid"
)

func must(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

func main() {
    cassHost := os.Getenv("CASSANDRA_HOST")
    if cassHost == "" {
        cassHost = "127.0.0.1"
    }

    // Connect to Cassandra
    cluster := gocql.NewCluster(cassHost)
    cluster.Consistency = gocql.Quorum
    sess, err := cluster.CreateSession()
    must(err)
    defer sess.Close()
    ctx := context.Background()

    // Create keyspace and tables
    fmt.Println("Applying schema...")
    schema := []string{
        `CREATE KEYSPACE IF NOT EXISTS chat WITH replication = {'class':'SimpleStrategy', 'replication_factor':1};`,

        `CREATE TABLE IF NOT EXISTS chat.group_messages (
            group_id      text,
            bucket        text,
            msg_id        timeuuid,
            sender_id     text,
            sender_name   text,
            ts            timestamp,
            PRIMARY KEY ((group_id, bucket), msg_id)
        ) WITH CLUSTERING ORDER BY (msg_id DESC);`,

        `CREATE TABLE IF NOT EXISTS chat.group_members (
            group_id uuid,
            user_id uuid,
            joined_at timestamp,
            PRIMARY KEY (group_id, user_id)
        );`,

        `CREATE TABLE IF NOT EXISTS chat.user_groups (
            user_id uuid,
            group_id uuid,
            joined_at timestamp,
            PRIMARY KEY (user_id, group_id)
        );`,
    }
    for _, q := range schema {
        must(sess.Query(q).WithContext(ctx).Exec())
    }

    // Seed users and groups
    fmt.Println("Seeding test data...")

    // Create 2 groups and 3 users
    groupA := uuid.New()
    groupB := uuid.New()
    user1 := uuid.New()
    user2 := uuid.New()
    user3 := uuid.New()
    now := time.Now()

    // user1 and user2 in groupA; user3 alone in groupB
    members := []struct {
        groupID uuid.UUID
        userID  uuid.UUID
    }{
        {groupA, user1},
        {groupA, user2},
        {groupB, user3},
    }
    for _, m := range members {
        must(sess.Query(
            "INSERT INTO chat.group_members (group_id, user_id, joined_at) VALUES (?, ?, ?)",
            m.groupID, m.userID, now,
        ).WithContext(ctx).Exec())
        must(sess.Query(
            "INSERT INTO chat.user_groups (user_id, group_id, joined_at) VALUES (?, ?, ?)",
            m.userID, m.groupID, now,
        ).WithContext(ctx).Exec())
    }

    // Print current seed data
    fmt.Println("\n=== GROUPS ===")
    fmt.Printf("GroupA ID: %v\nGroupB ID: %v\n", groupA, groupB)
    fmt.Println("\n=== USERS ===")
    fmt.Printf("User1 ID: %v\nUser2 ID: %v\nUser3 ID: %v\n", user1, user2, user3)
    fmt.Println("\n=== group_members ===")
    scanMembers := sess.Query("SELECT group_id, user_id, joined_at FROM chat.group_members").Iter().Scanner()
    for scanMembers.Next() {
        var gid, uid gocql.UUID
        var joined time.Time
        must(scanMembers.Scan(&gid, &uid, &joined))
        fmt.Printf("Group %v has member %v (joined %v)\n", gid, uid, joined.Format(time.RFC3339))
    }

    fmt.Println("\n=== user_groups ===")
    scanUsers := sess.Query("SELECT user_id, group_id, joined_at FROM chat.user_groups").Iter().Scanner()
    for scanUsers.Next() {
        var gid, uid gocql.UUID
        var joined time.Time
        must(scanUsers.Scan(&uid, &gid, &joined))
        fmt.Printf("User %v is in group %v (joined %v)\n", uid, gid, joined.Format(time.RFC3339))
    }
}
