package entity_test

import (
	"errors"
	"testing"
	"time"

	"github.com/l0gik67/go-todo-cli/entity"
)

func TestNewEventWithoutError(t *testing.T) {
	const rawInput = "  add   work   написать отчёт  "
	beforeCreation := time.Now()
	event := entity.NewEvent(rawInput, nil)
	afterCreation := time.Now()

	if event.GetText() != rawInput {
		t.Errorf("GetText() = %q, want the complete raw input %q", event.GetText(), rawInput)
	}
	if event.GetDescription() != "" {
		t.Errorf("GetDescription() = %q, want an empty string", event.GetDescription())
	}
	if createdAt := event.GetCreationTime(); createdAt.Before(beforeCreation) || createdAt.After(afterCreation) {
		t.Errorf("GetCreationTime() = %v, want a value between %v and %v", createdAt, beforeCreation, afterCreation)
	}
}

func TestNewEventWithError(t *testing.T) {
	inputErr := errors.New("example failure")
	event := entity.NewEvent("unknown argument", inputErr)

	if event.GetText() != "unknown argument" {
		t.Errorf("GetText() = %q, want %q", event.GetText(), "unknown argument")
	}
	if event.GetDescription() != inputErr.Error() {
		t.Errorf("GetDescription() = %q, want %q", event.GetDescription(), inputErr.Error())
	}
	if event.GetCreationTime().IsZero() {
		t.Error("GetCreationTime() returned a zero time")
	}
}

func TestEventStructListPreservesOrderAndReturnsCopy(t *testing.T) {
	events := entity.NewEventStruct()
	first := entity.NewEvent("first command", nil)
	second := entity.NewEvent("second command", errors.New("failed"))
	events.Add(first)
	events.Add(second)

	got := events.List()
	if len(got) != 2 {
		t.Fatalf("List() returned %d events, want 2", len(got))
	}
	if got[0] != first || got[1] != second {
		t.Errorf("List() did not preserve insertion order: %#v", got)
	}

	got[0] = nil
	again := events.List()
	if len(again) != 2 || again[0] != first || again[1] != second {
		t.Errorf("modifying a List() result changed the event log: %#v", again)
	}
}
