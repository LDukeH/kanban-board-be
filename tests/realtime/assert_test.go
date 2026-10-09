package realtime_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

type broadcastEvent struct {
	Topic   string
	Type    string
	Payload map[string]any
}

func testContext() context.Context { return context.Background() }

func capturedEvents(t *testing.T, pool *pgxpool.Pool) []broadcastEvent {
	t.Helper()
	rows, err := pool.Query(testContext(),
		`SELECT topic, event, payload FROM realtime.captured ORDER BY created_at, payload ->> 'entity_id', event`)
	if err != nil {
		t.Fatalf("read captured: %v", err)
	}
	defer rows.Close()

	var events []broadcastEvent
	for rows.Next() {
		var e broadcastEvent
		var raw json.RawMessage
		if err := rows.Scan(&e.Topic, &e.Type, &raw); err != nil {
			t.Fatalf("scan captured: %v", err)
		}
		if err := json.Unmarshal(raw, &e.Payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate captured: %v", err)
	}
	return events
}

func strField(t *testing.T, e broadcastEvent, key string) string {
	t.Helper()
	v, ok := e.Payload[key].(string)
	if !ok {
		t.Fatalf("payload %q = %v, want a string (event %+v)", key, e.Payload[key], e)
	}
	return v
}

func wantEnvelope(t *testing.T, pool *pgxpool.Pool, f fixture, wantType, entityID string) broadcastEvent {
	t.Helper()
	events := capturedEvents(t, pool)
	if len(events) != 1 {
		t.Fatalf("%s: captured %d events, want exactly 1: %+v", wantType, len(events), events)
	}
	e := events[0]
	if e.Topic != "board:"+f.boardID.String() {
		t.Fatalf("topic = %q, want %q", e.Topic, "board:"+f.boardID.String())
	}
	if e.Type != wantType || strField(t, e, "type") != wantType {
		t.Fatalf("type = %q (payload %q), want %q", e.Type, e.Payload["type"], wantType)
	}
	if strField(t, e, "board_id") != f.boardID.String() {
		t.Fatalf("board_id = %v, want %s", e.Payload["board_id"], f.boardID)
	}
	if strField(t, e, "entity_id") != entityID {
		t.Fatalf("entity_id = %v, want %s", e.Payload["entity_id"], entityID)
	}
	if strField(t, e, "actor_id") != f.userID.String() {
		t.Fatalf("actor_id = %v, want %s", e.Payload["actor_id"], f.userID)
	}
	for _, key := range []string{"id", "version", "occurred_at"} {
		if _, ok := e.Payload[key]; !ok {
			t.Fatalf("payload is missing %q: %+v", key, e.Payload)
		}
	}
	return e
}
