CREATE TABLE IF NOT EXISTS operators (
    id uuid PRIMARY KEY,
    name text NOT NULL,
    role text NOT NULL CHECK (role IN ('drone_operator', 'telemetry_operator', 'quality_reviewer', 'safety_supervisor')),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS drone_missions (
    id uuid PRIMARY KEY,
    code text NOT NULL UNIQUE,
    name text NOT NULL,
    status text NOT NULL CHECK (status IN ('draft', 'scheduled', 'active', 'closed')),
    timezone text NOT NULL,
    starts_at timestamptz NOT NULL,
    ends_at timestamptz NOT NULL,
    version bigint NOT NULL DEFAULT 1,
    created_by uuid NOT NULL REFERENCES operators(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (ends_at > starts_at)
);

CREATE TABLE IF NOT EXISTS mission_drones (
    id uuid PRIMARY KEY,
    mission_id uuid NOT NULL REFERENCES drone_missions(id) ON DELETE CASCADE,
    code text NOT NULL,
    room_label text NOT NULL,
    required_tasks integer NOT NULL CHECK (required_tasks > 0),
    completed_tasks integer NOT NULL DEFAULT 0 CHECK (completed_tasks >= 0),
    UNIQUE (mission_id, code)
);

CREATE TABLE IF NOT EXISTS assignments (
    id uuid PRIMARY KEY,
    mission_id uuid NOT NULL REFERENCES drone_missions(id) ON DELETE CASCADE,
    drone_id uuid NOT NULL REFERENCES mission_drones(id) ON DELETE CASCADE,
    operator_id uuid NOT NULL REFERENCES operators(id),
    status text NOT NULL CHECK (status IN ('queued','active','completed','cancelled')),
    starts_at timestamptz NOT NULL,
    ends_at timestamptz NOT NULL,
    version bigint NOT NULL DEFAULT 1,
    UNIQUE (drone_id, starts_at),
    CHECK (ends_at > starts_at)
);

CREATE TABLE IF NOT EXISTS drone_tasks (
    id uuid PRIMARY KEY,
    mission_id uuid NOT NULL REFERENCES drone_missions(id),
    drone_id uuid NOT NULL REFERENCES mission_drones(id),
    task_code text NOT NULL UNIQUE,
    status text NOT NULL CHECK (status IN ('queued', 'completed', 'device_transfer_pending', 'accepted', 'in_progress', 'verified', 'rejected', 'archived')),
    completed_at timestamptz,
    accepted_at timestamptz,
    expires_at timestamptz NOT NULL,
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS device_transfer_events (
    id uuid PRIMARY KEY,
    drone_task_id uuid NOT NULL REFERENCES drone_tasks(id) ON DELETE CASCADE,
    from_operator uuid REFERENCES operators(id),
    to_operator uuid NOT NULL REFERENCES operators(id),
    location text NOT NULL,
    recorded_at timestamptz NOT NULL,
    note text NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS mission_batches (
    id uuid PRIMARY KEY,
    code text NOT NULL UNIQUE,
    status text NOT NULL CHECK (status IN ('queued', 'running', 'completed', 'cancelled')),
    method text NOT NULL,
    capacity integer NOT NULL CHECK (capacity > 0),
    started_at timestamptz,
    completed_at timestamptz,
    version bigint NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS mission_batch_tasks (
    mission_batch_id uuid NOT NULL REFERENCES mission_batches(id) ON DELETE CASCADE,
    drone_task_id uuid NOT NULL REFERENCES drone_tasks(id),
    PRIMARY KEY (mission_batch_id, drone_task_id)
);

CREATE TABLE IF NOT EXISTS telemetry_events (
    id uuid PRIMARY KEY,
    drone_task_id uuid NOT NULL REFERENCES drone_tasks(id),
    mission_batch_id uuid NOT NULL REFERENCES mission_batches(id),
    recorded_by uuid NOT NULL REFERENCES operators(id),
    status text NOT NULL CHECK (status IN ('pending', 'verified', 'rejected')),
    value numeric(18,6) NOT NULL,
    unit text NOT NULL,
    limit_value numeric(18,6) NOT NULL,
    measured_at timestamptz NOT NULL,
    reviewed_at timestamptz,
    version bigint NOT NULL DEFAULT 1,
    UNIQUE (drone_task_id, mission_batch_id)
);

CREATE TABLE IF NOT EXISTS safety_alerts (
    id uuid PRIMARY KEY,
    drone_task_id uuid NOT NULL REFERENCES drone_tasks(id),
    kind text NOT NULL CHECK (kind IN ('retask', 'repeat_drone', 'safety_adjustment', 'close_record')),
    status text NOT NULL CHECK (status IN ('open', 'in_progress', 'closed')),
    reason text NOT NULL,
    due_at timestamptz NOT NULL,
    closed_at timestamptz,
    UNIQUE (drone_task_id, kind, status)
);

CREATE TABLE IF NOT EXISTS audit_events (
    id uuid PRIMARY KEY,
    request_id text NOT NULL,
    operator_id uuid REFERENCES operators(id),
    object_type text NOT NULL,
    object_id uuid NOT NULL,
    action text NOT NULL,
    outcome text NOT NULL,
    detail jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS idempotency_keys (
    key text PRIMARY KEY,
    request_hash text NOT NULL,
    response_code integer NOT NULL,
    response_body jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS drone_tasks_mission_status_idx ON drone_tasks(mission_id, status);
CREATE INDEX IF NOT EXISTS drone_tasks_expiry_idx ON drone_tasks(status, expires_at);
CREATE INDEX IF NOT EXISTS device_transfer_task_time_idx ON device_transfer_events(drone_task_id, recorded_at DESC);
CREATE INDEX IF NOT EXISTS telemetry_events_status_idx ON telemetry_events(status, measured_at);
CREATE INDEX IF NOT EXISTS safety_alerts_due_idx ON safety_alerts(status, due_at);
CREATE INDEX IF NOT EXISTS audit_object_idx ON audit_events(object_type, object_id, created_at DESC);
