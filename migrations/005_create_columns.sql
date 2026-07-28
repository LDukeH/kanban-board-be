CREATE TABLE columns (
    id UUID PRIMARY KEY,
    board_id UUID NOT NULL,
    name TEXT NOT NULL,
    position INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_columns_board
        FOREIGN KEY (board_id)
        REFERENCES boards(id)
        ON DELETE CASCADE
);

