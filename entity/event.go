package entity

import "time"

// EventStruct stores events in the order in which they occurred.
type EventStruct struct {
	events []*Event
}

func NewEventStruct() *EventStruct {
	return &EventStruct{
		events: make([]*Event, 0, 10),
	}
}

func (e *EventStruct) Add(event *Event) {
	e.events = append(e.events, event)
}

// List returns a copy of the event slice so callers cannot modify the log.
func (e *EventStruct) List() []*Event {
	events := make([]*Event, len(e.events))
	copy(events, e.events)
	return events
}

// Event describes one complete line entered by the user.
type Event struct {
	text         string
	description  string
	creationTime time.Time
}

func NewEvent(text string, err error) *Event {
	description := ""
	if err != nil {
		description = err.Error()
	}

	return &Event{
		text:         text,
		description:  description,
		creationTime: time.Now(),
	}
}

func (e *Event) GetText() string {
	return e.text
}

func (e *Event) GetDescription() string {
	return e.description
}

func (e *Event) GetCreationTime() time.Time {
	return e.creationTime
}
