-- Migration: Create available_resources table
-- Filename: 005_create_available_resources_table.cql

-- Create the available_resources table
CREATE TABLE IF NOT EXISTS task_scheduler.available_resources (
    id text PRIMARY KEY,
    cpu int,
    memory int,
    storage int
);

-- Insert initial values
INSERT INTO task_scheduler.available_resources (id, cpu, memory, storage)
VALUES ('global', 100, 1024000, 1024000);
