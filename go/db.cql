CREATE KEYSPACE IF NOT EXISTS chat
WITH replication = {'class':'SimpleStrategy', 'replication_factor':1};

CREATE TABLE IF NOT EXISTS chat.group_messages (
    group_id      text,
    bucket        text,
    msg_id        timeuuid,
    sender_id     text,
    sender_name   text,
    ts            timestamp,           -- renamed from "timestamp"
    PRIMARY KEY ((group_id, bucket), msg_id)
) WITH CLUSTERING ORDER BY (msg_id DESC);

CREATE TABLE IF NOT EXISTS chat.group_members (
    group_id uuid,
    user_id uuid,
    joined_at timestamp,
    PRIMARY KEY (group_id, user_id)
);

CREATE TABLE IF NOT EXISTS chat.user_groups (
    user_id uuid,
    group_id uuid,
    joined_at timestamp,
    PRIMARY KEY (user_id, group_id)
);
