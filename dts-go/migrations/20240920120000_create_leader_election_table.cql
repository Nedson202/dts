CREATE TABLE IF NOT EXISTS leader_election (
    key text PRIMARY KEY,
    instance_id text,
    last_heartbeat timestamp
);
