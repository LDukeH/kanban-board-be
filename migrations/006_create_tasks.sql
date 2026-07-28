CREATE TABLE tasks (
    id UUID PRIMARY KEY,
    column_id UUID NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    position INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_tasks_column
        FOREIGN KEY (column_id)
        REFERENCES columns(id)
        ON DELETE CASCADE
);