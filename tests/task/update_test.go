package task_test

import (
	"encoding/json"
	"strings"
	"testing"

	. "kanban-board-be/internal/feature/task"
)

func strp(s string) *string { return &s }

func boolp(b bool) *bool { return &b }

func jsonRaw(s string) json.RawMessage { return json.RawMessage(s) }

func datep(t *testing.T, s string) Optional[DateOnly] {
	t.Helper()
	return Some(mustParseDate(t, s))
}

func TestBuildTaskUpdateSetEmpty(t *testing.T) {
	clauses, args := BuildTaskUpdateSet(&Task{})
	if len(clauses) != 0 {
		t.Fatalf("clauses = %v, want none", clauses)
	}
	if len(args) != 0 {
		t.Fatalf("args = %v, want none", args)
	}
}

func TestBuildTaskUpdateSetLegacyFieldsRenumber(t *testing.T) {
	// Old code hardcoded description to $2; with only description set the
	// placeholder must be $1 now.
	desc := jsonRaw(`{"type":"doc","content":[]}`)
	clauses, args := BuildTaskUpdateSet(&Task{Description: &desc})
	if len(clauses) != 1 || clauses[0] != "description = $1" {
		t.Fatalf("clauses = %v, want [description = $1]", clauses)
	}
	if len(args) != 1 {
		t.Fatalf("args = %v, want one arg", args)
	}
}

func TestBuildTaskUpdateSetTitleOnlyKeepsLegacyNumbering(t *testing.T) {
	clauses, _ := BuildTaskUpdateSet(&Task{Title: strp("renamed")})
	if len(clauses) != 1 || clauses[0] != "title = $1" {
		t.Fatalf("clauses = %v, want [title = $1]", clauses)
	}
}

func TestBuildTaskUpdateSetAllFields(t *testing.T) {
	desc := jsonRaw(`{"type":"doc","content":[]}`)
	completed := true
	task := &Task{
		Title:       strp("t"),
		Description: &desc,
		Priority:    strp(PriorityHigh),
		Completed:   boolp(completed),
		DueDate:     datep(t, "2026-10-02"),
	}
	clauses, args := BuildTaskUpdateSet(task)

	want := []string{
		"title = $1",
		"description = $2",
		"priority = $3",
		"completed = $4",
		"due_date = $5",
	}
	if strings.Join(clauses, ",") != strings.Join(want, ",") {
		t.Fatalf("clauses = %v, want %v", clauses, want)
	}
	if len(args) != 5 {
		t.Fatalf("args = %v, want five args", args)
	}
}

func TestBuildTaskUpdateSetDueDateClearEmitsNoArg(t *testing.T) {
	task := &Task{DueDate: Null[DateOnly]()}
	clauses, args := BuildTaskUpdateSet(task)
	if len(clauses) != 1 || clauses[0] != "due_date = NULL" {
		t.Fatalf("clauses = %v, want [due_date = NULL]", clauses)
	}
	if len(args) != 0 {
		t.Fatalf("args = %v, want no args", args)
	}
}

func TestBuildTaskUpdateSetDueDateAbsentEmitsNothing(t *testing.T) {
	task := &Task{Title: strp("t"), DueDate: Unset[DateOnly]()}
	clauses, args := BuildTaskUpdateSet(task)
	if len(clauses) != 1 || clauses[0] != "title = $1" {
		t.Fatalf("clauses = %v, want [title = $1]", clauses)
	}
	if len(args) != 1 {
		t.Fatalf("args = %v, want one arg", args)
	}
}

func TestBuildTaskUpdateSetCompletedFalseIsAnUpdate(t *testing.T) {
	// A false pointer must still render: completed=false is a real write.
	task := &Task{Completed: boolp(false)}
	clauses, args := BuildTaskUpdateSet(task)
	if len(clauses) != 1 || clauses[0] != "completed = $1" {
		t.Fatalf("clauses = %v, want [completed = $1]", clauses)
	}
	if len(args) != 1 || args[0] != false {
		t.Fatalf("args = %v, want [false]", args)
	}
}

func TestIsValidTaskPriority(t *testing.T) {
	for _, p := range []string{PriorityNone, PriorityLow, PriorityMedium, PriorityHigh} {
		if !IsValidTaskPriority(p) {
			t.Fatalf("priority %q rejected, want accepted", p)
		}
	}
	for _, p := range []string{"", "urgent", "HIGH", " none", "none "} {
		if IsValidTaskPriority(p) {
			t.Fatalf("priority %q accepted, want rejected", p)
		}
	}
}
