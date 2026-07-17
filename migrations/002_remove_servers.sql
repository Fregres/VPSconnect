DROP INDEX IF EXISTS metric_samples_server_time_idx;

ALTER TABLE metric_samples
    DROP COLUMN server_id;

DROP TABLE servers;

CREATE INDEX metric_samples_collected_at_idx
    ON metric_samples (collected_at);
