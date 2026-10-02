package task

import (
	"encoding/json"
	"testing"
	"time"
)

func mustParseDate(t *testing.T, s string) DateOnly {
	t.Helper()
	parsed, err := time.Parse(DateLayout, s)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return DateOnly{Time: parsed}
}

func TestOptionalUnmarshalAbsentStaysUnset(t *testing.T) {
	var req TaskUpdateRequest
	if err := json.Unmarshal([]byte(`{"title":"keep"}`), &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.DueDate.State() != OptionalUnset {
		t.Fatalf("due_date state = %v, want OptionalUnset", req.DueDate.State())
	}
	if req.DueDate.IsSet() {
		t.Fatal("IsSet = true for an absent key, want false")
	}
}

func TestOptionalUnmarshalNullClears(t *testing.T) {
	var req TaskUpdateRequest
	if err := json.Unmarshal([]byte(`{"due_date":null}`), &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.DueDate.State() != OptionalNull {
		t.Fatalf("due_date state = %v, want OptionalNull", req.DueDate.State())
	}
	if !req.DueDate.IsSet() || !req.DueDate.IsNull() {
		t.Fatal("IsSet/IsNull flags wrong for explicit null")
	}
	if _, ok := req.DueDate.Get(); ok {
		t.Fatal("Get returned a value for explicit null")
	}
}

func TestOptionalUnmarshalValueSets(t *testing.T) {
	var req TaskUpdateRequest
	if err := json.Unmarshal([]byte(`{"due_date":"2026-10-02"}`), &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.DueDate.State() != OptionalValue {
		t.Fatalf("due_date state = %v, want OptionalValue", req.DueDate.State())
	}
	got, ok := req.DueDate.Get()
	if !ok {
		t.Fatal("Get ok = false for a value, want true")
	}
	if got.Format(DateLayout) != "2026-10-02" {
		t.Fatalf("due_date = %q, want 2026-10-02", got.Format(DateLayout))
	}
}

func TestOptionalUnmarshalRejectsBadDate(t *testing.T) {
	var req TaskUpdateRequest
	if err := json.Unmarshal([]byte(`{"due_date":"10/02/2026"}`), &req); err == nil {
		t.Fatal("expected an error for a non-YYYY-MM-DD date, got nil")
	}

	var reqTime TaskUpdateRequest
	if err := json.Unmarshal([]byte(`{"due_date":"2026-10-02T00:00:00Z"}`), &reqTime); err == nil {
		t.Fatal("expected an error for an RFC3339 timestamp, got nil")
	}
}

func TestOptionalMarshalShapes(t *testing.T) {
	render := func(o Optional[DateOnly]) string {
		b, err := json.Marshal(o)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		return string(b)
	}

	if got := render(Unset[DateOnly]()); got != "null" {
		t.Fatalf("Unset marshals to %s, want null", got)
	}
	if got := render(Null[DateOnly]()); got != "null" {
		t.Fatalf("Null marshals to %s, want null", got)
	}
	if got := render(Some(mustParseDate(t, "2026-10-02"))); got != `"2026-10-02"` {
		t.Fatalf("Value marshals to %s, want \"2026-10-02\"", got)
	}
}

func TestOptionalRoundTripValue(t *testing.T) {
	in := Some(mustParseDate(t, "2026-10-02"))
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out Optional[DateOnly]
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got, ok := out.Get()
	if !ok || got.Format(DateLayout) != "2026-10-02" {
		t.Fatalf("round trip = %v, %v; want 2026-10-02, true", got, ok)
	}
}
