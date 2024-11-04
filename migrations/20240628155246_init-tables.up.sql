-- Create partitions table
CREATE TABLE partitions(
    id TEXT,
    created_at_nano BIGINT NOT NULL,
    deleted_at_nano BIGINT,
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
    PRIMARY KEY (id)
);

-- Create applications table
CREATE TABLE applications(
    id TEXT, -- internal id
    created_at_nano BIGINT NOT NULL,
    deleted_at_nano BIGINT,
    app_id TEXT NOT NULL,
    used_resource JSONB,
    max_used_resource JSONB,
    pending_resource JSONB,
    partition_id TEXT NOT NULL,
    partition TEXT NOT NULL,
    queue_name TEXT NOT NULL,
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
    PRIMARY KEY (id)
);

-- Create queues table
CREATE TABLE queues(
    id TEXT NOT NULL,
    created_at_nano BIGINT NOT NULL,
    deleted_at_nano BIGINT,
    queue_name TEXT NOT NULL,
    parent_id TEXT REFERENCES queues(id),
    parent TEXT,
    status TEXT,
    partition_id TEXT NOT NULL CHECK (partition_id <> ''),
    pending_resource JSONB,
    max_resource JSONB,
    guaranteed_resource JSONB,
    allocated_resource JSONB,
    preempting_resource JSONB,
    head_room JSONB,
    is_leaf BOOLEAN,
    is_managed BOOLEAN,
    properties JSONB,
    template_info JSONB,
    abs_used_capacity JSONB,
    max_running_apps INTEGER,
    running_apps INTEGER NOT NULL,
    current_priority INTEGER,
    allocating_accepted_apps TEXT[],
    PRIMARY KEY (id)
);

-- Create nodes table
CREATE TABLE nodes(
    id TEXT,
    created_at_nano BIGINT NOT NULL,
    deleted_at_nano BIGINT,
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

-- Drop history_type if it exists
DROP TYPE IF EXISTS history_type;

-- Create history_type enum
CREATE TYPE history_type AS ENUM ('container', 'application');

-- Create history table
CREATE TABLE history(
    id TEXT,
    created_at_nano BIGINT NOT NULL,
    deleted_at_nano BIGINT,
    history_type history_type NOT NULL,
    total_number BIGINT NOT NULL,
    timestamp BIGINT NOT NULL,
    UNIQUE (id),
    PRIMARY KEY (id)
);
