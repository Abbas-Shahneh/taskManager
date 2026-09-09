CREATE TABLE tasks (
    id UUID PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    assignee VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT tasks_status_check
        CHECK (
            status IN (
                'pending',
                'in_progress',
                'completed'
            )
        )
);

CREATE INDEX idx_tasks_status
    ON tasks (status);

CREATE INDEX idx_tasks_assignee
    ON tasks (assignee);

CREATE INDEX idx_tasks_created_at
    ON tasks (created_at DESC);