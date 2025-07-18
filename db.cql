CREATE KEYSPACE chat
WITH replication = {'class':'SimpleStrategy', 'replication_factor':1};

-- partition by group-id and bucket
CREATE TABLE group_messages (
    group_id      text,
    bucket        text,  -- bucket by week, "2025-W29"
    msg_id        timeuuid, 
    sender_id     text,
    sender_name   text,
    --sender_avatar text,  
    timestamp     timestamp,
    PRIMARY KEY ((group_id, bucket), msg_id)
) WITH CLUSTERING ORDER BY (msg_id DESC);
