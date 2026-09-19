-- +goose Up

CREATE TABLE labels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    board_id UUID NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_labels_board
        FOREIGN KEY (board_id)
        REFERENCES boards(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_labels_board_name
        UNIQUE (board_id, name)
);