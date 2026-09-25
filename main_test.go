package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/k0kubun/pp"
	"github.com/l0gik67/go-todo-cli/entity"
)

func executeCommand(
	t *testing.T,
	rawInput string,
	tasks *entity.TaskStruct,
	events *entity.EventStruct,
) (string, bool) {
	t.Helper()
	pp.ColoringEnabled = false
	var output bytes.Buffer
	shouldExit := checkCommand(rawInput, tasks, events, &output)
	return output.String(), shouldExit
}

func TestCheckCommandAddStoresTaskAndCompleteInput(t *testing.T) {
	tasks := entity.NewTaskStruct()
	events := entity.NewEventStruct()
	const rawInput = "  add   work   написать отчёт  "

	output, shouldExit := executeCommand(t, rawInput, tasks, events)
	if shouldExit {
		t.Error("add command requested exit")
	}
	if output != "" {
		t.Errorf("successful add output = %q, want no output", output)
	}

	gotTasks := tasks.List()
	if len(gotTasks) != 1 {
		t.Fatalf("add stored %d tasks, want 1", len(gotTasks))
	}
	if gotTasks[0].GetTitle() != "work" || gotTasks[0].GetText() != "написать отчёт" {
		t.Errorf("stored task = (%q, %q), want (%q, %q)", gotTasks[0].GetTitle(), gotTasks[0].GetText(), "work", "написать отчёт")
	}

	gotEvents := events.List()
	if len(gotEvents) != 1 {
		t.Fatalf("add stored %d events, want 1", len(gotEvents))
	}
	if gotEvents[0].GetText() != rawInput {
		t.Errorf("event text = %q, want complete raw input %q", gotEvents[0].GetText(), rawInput)
	}
	if gotEvents[0].GetDescription() != "" {
		t.Errorf("successful add description = %q, want empty", gotEvents[0].GetDescription())
	}
}

func TestCheckCommandRejectsMalformedInputWithoutChangingTasks(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "empty", input: ""},
		{name: "whitespace", input: " \t "},
		{name: "add without arguments", input: "add"},
		{name: "add without text", input: "add title"},
		{name: "delete without title", input: "del"},
		{name: "delete with extra argument", input: "del title extra"},
		{name: "done without title", input: "done"},
		{name: "done with extra argument", input: "done title extra"},
		{name: "help with argument", input: "help extra"},
		{name: "list with argument", input: "list extra"},
		{name: "events with argument", input: "events extra"},
		{name: "exit with argument", input: "exit extra"},
		{name: "unknown", input: "unknown argument"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks := entity.NewTaskStruct()
			events := entity.NewEventStruct()

			output, shouldExit := executeCommand(t, tt.input, tasks, events)
			if shouldExit {
				t.Error("invalid command requested exit")
			}
			if !strings.Contains(output, "Ошибка:") {
				t.Errorf("output = %q, want a user-visible error", output)
			}
			if got := tasks.List(); len(got) != 0 {
				t.Errorf("invalid command changed task store: %#v", got)
			}

			gotEvents := events.List()
			if len(gotEvents) != 1 {
				t.Fatalf("invalid input stored %d events, want 1", len(gotEvents))
			}
			if gotEvents[0].GetText() != tt.input {
				t.Errorf("event text = %q, want raw input %q", gotEvents[0].GetText(), tt.input)
			}
			if gotEvents[0].GetDescription() == "" {
				t.Error("invalid input event has an empty description")
			}
		})
	}
}

func TestCheckCommandDuplicateAndMissingTaskErrors(t *testing.T) {
	tasks := entity.NewTaskStruct()
	events := entity.NewEventStruct()

	if output, _ := executeCommand(t, "add work original text", tasks, events); output != "" {
		t.Fatalf("initial add output = %q, want no output", output)
	}
	output, shouldExit := executeCommand(t, "add work replacement text", tasks, events)
	if shouldExit || !strings.Contains(output, "Ошибка:") {
		t.Errorf("duplicate add output = %q, exit = %v, want an error and no exit", output, shouldExit)
	}
	gotTasks := tasks.List()
	if len(gotTasks) != 1 || gotTasks[0].GetText() != "original text" {
		t.Errorf("duplicate add changed stored task: %#v", gotTasks)
	}

	for _, input := range []string{"done missing", "del missing"} {
		output, shouldExit = executeCommand(t, input, tasks, events)
		if shouldExit || !strings.Contains(output, "Ошибка:") {
			t.Errorf("%q output = %q, exit = %v, want an error and no exit", input, output, shouldExit)
		}
	}
	if gotTasks := tasks.List(); len(gotTasks) != 1 || gotTasks[0].GetText() != "original text" {
		t.Errorf("missing-task commands changed stored tasks: %#v", gotTasks)
	}

	gotEvents := events.List()
	if len(gotEvents) != 4 {
		t.Fatalf("stored %d events, want one for each of 4 inputs", len(gotEvents))
	}
	for i, event := range gotEvents[1:] {
		if event.GetDescription() == "" {
			t.Errorf("failed event %d has an empty description", i+2)
		}
	}
}

func TestCheckCommandDoneStoresCompletionTimeOnce(t *testing.T) {
	tasks := entity.NewTaskStruct()
	events := entity.NewEventStruct()
	_, _ = executeCommand(t, "add work finish tests", tasks, events)

	beforeCompletion := time.Now()
	output, shouldExit := executeCommand(t, "done work", tasks, events)
	afterCompletion := time.Now()
	if shouldExit || output != "" {
		t.Errorf("successful done output = %q, exit = %v, want no output and no exit", output, shouldExit)
	}
	task := tasks.List()[0]
	completedAt, completed := task.GetTimeCompletion()
	if !task.IsDone() || !completed {
		t.Fatal("done command did not store completed state and time")
	}
	if completedAt.Before(beforeCompletion) || completedAt.After(afterCompletion) {
		t.Errorf("completion time = %v, want a value between %v and %v", completedAt, beforeCompletion, afterCompletion)
	}

	output, shouldExit = executeCommand(t, "done work", tasks, events)
	if shouldExit || !strings.Contains(output, "Ошибка:") {
		t.Errorf("second done output = %q, exit = %v, want an error and no exit", output, shouldExit)
	}
	afterSecondDone, completed := task.GetTimeCompletion()
	if !completed || !afterSecondDone.Equal(completedAt) {
		t.Errorf("second done changed completion time from %v to %v", completedAt, afterSecondDone)
	}
}

func TestCheckCommandListShowsSortedTasksAndCompletion(t *testing.T) {
	tasks := entity.NewTaskStruct()
	events := entity.NewEventStruct()
	_, _ = executeCommand(t, "add zeta last task", tasks, events)
	_, _ = executeCommand(t, "add alpha first task", tasks, events)
	_, _ = executeCommand(t, "done alpha", tasks, events)

	completedAt, completed := tasks.List()[0].GetTimeCompletion()
	if !completed {
		t.Fatal("test setup did not complete alpha task")
	}
	output, shouldExit := executeCommand(t, "list", tasks, events)
	if shouldExit {
		t.Error("list command requested exit")
	}
	if !strings.Contains(output, "Список задач:") ||
		!strings.Contains(output, "first task") ||
		!strings.Contains(output, "last task") ||
		!strings.Contains(output, "выполнена") ||
		!strings.Contains(output, formatTime(completedAt)) {
		t.Errorf("list output does not contain complete task information:\n%s", output)
	}
	alphaPosition := strings.Index(output, `"alpha"`)
	zetaPosition := strings.Index(output, `"zeta"`)
	if alphaPosition < 0 || zetaPosition < 0 || alphaPosition >= zetaPosition {
		t.Errorf("list output is not sorted by title:\n%s", output)
	}
}

func TestEventsCommandIncludesItself(t *testing.T) {
	tasks := entity.NewTaskStruct()
	events := entity.NewEventStruct()
	const addInput = "add work write tests"
	const badInput = "unknown with arguments"
	_, _ = executeCommand(t, addInput, tasks, events)
	_, _ = executeCommand(t, badInput, tasks, events)

	output, shouldExit := executeCommand(t, "events", tasks, events)
	if shouldExit {
		t.Error("events command requested exit")
	}
	gotEvents := events.List()
	if len(gotEvents) != 3 {
		t.Fatalf("events command left %d events, want 3 including itself", len(gotEvents))
	}
	if gotEvents[2].GetText() != "events" || gotEvents[2].GetDescription() != "" {
		t.Errorf("last event = (%q, %q), want successful events command", gotEvents[2].GetText(), gotEvents[2].GetDescription())
	}
	if gotEvents[1].GetText() != badInput || gotEvents[1].GetDescription() == "" {
		t.Errorf("failed event = (%q, %q), want complete input and error description", gotEvents[1].GetText(), gotEvents[1].GetDescription())
	}
	for _, want := range []string{addInput, badInput, "неизвестная команда", `"events"`} {
		if !strings.Contains(output, want) {
			t.Errorf("events output does not contain %q:\n%s", want, output)
		}
	}
}

func TestCheckCommandExitIsRecordedBeforeExit(t *testing.T) {
	tasks := entity.NewTaskStruct()
	events := entity.NewEventStruct()
	const rawInput = "  exit  "

	output, shouldExit := executeCommand(t, rawInput, tasks, events)
	if !shouldExit {
		t.Error("valid exit command did not request exit")
	}
	if output != "" {
		t.Errorf("exit output = %q, want no output", output)
	}
	gotEvents := events.List()
	if len(gotEvents) != 1 || gotEvents[0].GetText() != rawInput || gotEvents[0].GetDescription() != "" {
		t.Errorf("exit event = %#v, want one successful event with exact raw input", gotEvents)
	}
}

func TestRunEndToEnd(t *testing.T) {
	input := strings.Join([]string{
		"add study изучить основы Go",
		"list",
		"done study",
		"list",
		"unknown command",
		"events",
		"exit",
		"",
	}, "\n")
	var output bytes.Buffer

	if err := run(strings.NewReader(input), &output); err != nil {
		t.Fatalf("run() returned an unexpected error: %v", err)
	}
	got := output.String()
	if strings.Count(got, "Список задач:") != 2 {
		t.Errorf("run output contains %d task lists, want 2:\n%s", strings.Count(got, "Список задач:"), got)
	}
	for _, want := range []string{
		"study",
		"изучить основы Go",
		"не выполнена",
		"выполнена",
		"Ошибка:",
		"unknown command",
		"Список событий:",
		`"events"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("run output does not contain %q:\n%s", want, got)
		}
	}
}

func TestRunProcessesFinalLineWithoutNewline(t *testing.T) {
	var output bytes.Buffer
	if err := run(strings.NewReader("help"), &output); err != nil {
		t.Fatalf("run() returned an unexpected error: %v", err)
	}
	if !strings.Contains(output.String(), "Доступные команды:") {
		t.Errorf("run output = %q, want help text", output.String())
	}
}
