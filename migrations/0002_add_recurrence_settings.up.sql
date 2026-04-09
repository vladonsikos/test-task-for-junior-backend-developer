CREATE TABLE IF NOT EXISTS recurrence_settings (
    id          BIGSERIAL PRIMARY KEY,
    task_id     BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    type        TEXT NOT NULL,
    interval    INT,
    day_of_month INT,
    dates       DATE[],
    parity      TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_recurrence_settings_task_id
    ON recurrence_settings (task_id);
