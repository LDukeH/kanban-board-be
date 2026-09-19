-- +goose Up

-- Convert tasks.description from TEXT to JSONB so the column can store the
-- Tiptap JSON document the frontend sends and reads back:
--
--   { "type": "doc", "content": [ ... ] }
--
-- The Go API has only ever bound description as *string, so every existing row
-- holds plain text. A plain `description::jsonb` cast would abort on those rows
-- (plain text is not valid JSON), so the USING clause wraps each non-empty
-- value into a single-paragraph Tiptap document, preserving the text exactly:
--
--   'Some legacy text'
--     -> { "type": "doc",
--          "content": [
--            { "type": "paragraph",
--              "content": [ { "type": "text", "text": "Some legacy text" } ] }
--          ] }
--
-- NULL and empty/whitespace-only text stay NULL (no description).
--
-- Re-run guard: rows whose column is already JSONB (migration applied twice)
-- are passed through unchanged instead of being wrapped a second time.
--
-- NOTE: this migration must land together with the Go model change that binds
-- the request/response description as a JSON document (json.RawMessage with a
-- top-level "type": "doc" check). After it is applied, plain-text writes fail
-- the jsonb input parse, and the current *string binding cannot send the doc
-- object the frontend uses. seed_2.sql inserts plain-text descriptions, so it
-- must be run BEFORE this migration (rows get wrapped by the USING clause).
-- Seeding a database that already has this migration requires the seed text to
-- be wrapped as documents too.

ALTER TABLE tasks
    ALTER COLUMN description TYPE JSONB
    USING CASE
        WHEN description IS NULL THEN NULL
        WHEN pg_typeof(description)::text = 'jsonb' THEN description::jsonb
        WHEN btrim(description::text) = '' THEN NULL
        ELSE jsonb_build_object(
            'type', 'doc',
            'content', jsonb_build_array(
                jsonb_build_object(
                    'type', 'paragraph',
                    'content', jsonb_build_array(
                        jsonb_build_object('type', 'text', 'text', description)
                    )
                )
            )
        )
    END;
