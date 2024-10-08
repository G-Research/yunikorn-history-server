-- Create partitions table
CREATE TABLE partitions(
    id TEXT,
    cluster_id TEXT NOT NULL,
    name TEXT NOT NULL,
    capacity JSONB,
    used_capacity JSONB,
    utilization JSONB,
    total_nodes INTEGER,
    applications JSONB,
    total_containers INTEGER,
    state TEXT,
    last_state_transition_time BIGINT,
    deleted_at BIGINT,
    UNIQUE (id),
    UNIQUE (name),
    PRIMARY KEY (id)
);

-- Create applications table
CREATE TABLE applications(
    id TEXT,
    app_id TEXT NOT NULL,
    used_resource JSONB,
    max_used_resource JSONB,
    pending_resource JSONB,
    partition TEXT NOT NULL,
    queue_name TEXT NOT NULL,
    queue_id UUID NOT NULL,
    submission_time BIGINT,
    finished_time BIGINT,
    requests JSONB,
    allocations JSONB,
    state TEXT,
    "user" TEXT,
    groups TEXT[],
    rejected_message TEXT,
    state_log JSONB,
    place_holder_data JSONB,
    has_reserved BOOLEAN,
    reservations JSONB,
    max_request_priority INTEGER,
    UNIQUE (id),
    PRIMARY KEY (id)
);

-- Create unique index on applications
CREATE UNIQUE INDEX idx_partition_queue_app_id ON applications (partition, queue_name, app_id);

-- Create queues table
CREATE TABLE queues(
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    parent_id UUID,
    created_at BIGINT NOT NULL,
    deleted_at BIGINT,
    queue_name TEXT NOT NULL,
    status TEXT,
    partition TEXT NOT NULL CHECK (partition <> ''),
    pending_resource JSONB,
    max_resource JSONB,
    guaranteed_resource JSONB,
    allocated_resource JSONB,
    preempting_resource JSONB,
    head_room JSONB,
    is_leaf BOOLEAN,
    is_managed BOOLEAN,
    properties JSONB,
    parent TEXT,
    template_info JSONB,
    abs_used_capacity JSONB,
    max_running_apps INTEGER,
    running_apps INTEGER NOT NULL,
    current_priority INTEGER,
    allocating_accepted_apps TEXT[],
    UNIQUE (id),
    PRIMARY KEY (id)
);

-- Create unique index on queues
CREATE UNIQUE INDEX idx_partition_queue_name ON queues (partition, queue_name);

-- Create nodes table
CREATE TABLE nodes(
    id TEXT,
    node_id TEXT NOT NULL,
    partition TEXT NOT NULL,
    host_name TEXT NOT NULL,
    rack_name TEXT,
    attributes JSONB,
    capacity JSONB,
    allocated JSONB,
    occupied JSONB,
    available JSONB,
    utilized JSONB,
    allocations JSONB,
    schedulable BOOLEAN,
    is_reserved BOOLEAN,
    reservations TEXT[],
    UNIQUE (id),
    UNIQUE (node_id),
    PRIMARY KEY (id)
);

-- Create partition_nodes_util table
CREATE TABLE partition_nodes_util(
    id TEXT,
    cluster_id TEXT NOT NULL,
    partition TEXT NOT NULL,
    nodes_util_list JSONB,
    UNIQUE (id),
    PRIMARY KEY (id)
);

-- Drop history_type if it exists
DROP TYPE IF EXISTS history_type;

-- Create history_type enum
CREATE TYPE history_type AS ENUM ('container', 'application');

-- Create history table
CREATE TABLE history(
    id TEXT,
    history_type history_type NOT NULL,
    total_number BIGINT NOT NULL,
    timestamp BIGINT NOT NULL,
    UNIQUE (id),
    PRIMARY KEY (id)
);
