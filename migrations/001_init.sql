CREATE TABLE servers (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE metric_samples (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    server_id BIGINT NOT NULL
        REFERENCES servers(id)
        ON DELETE CASCADE,

    collected_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    cpu_usage_percent DOUBLE PRECISION NOT NULL
        CHECK (
            cpu_usage_percent >= 0
            AND cpu_usage_percent <= 100
        ),

    memory_total_bytes BIGINT NOT NULL
        CHECK (memory_total_bytes >= 0),

    memory_used_bytes BIGINT NOT NULL
        CHECK (
            memory_used_bytes >= 0
            AND memory_used_bytes <= memory_total_bytes
        ),

    disk_total_bytes BIGINT NOT NULL
        CHECK (disk_total_bytes >= 0),

    disk_used_bytes BIGINT NOT NULL
        CHECK (
            disk_used_bytes >= 0
            AND disk_used_bytes <= disk_total_bytes
        ),

    uptime_seconds BIGINT NOT NULL
        CHECK (uptime_seconds >= 0)
);

CREATE INDEX metric_samples_server_time_idx
    ON metric_samples (server_id, collected_at);
