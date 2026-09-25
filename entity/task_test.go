package entity_test

import (
	"strings"
	"testing"
	"time"

	"github.com/l0gik67/go-todo-cli/entity"
)

func TestNewTask(t *testing.T) {
	beforeCreation := time.Now()
	task, err := entity.NewTask("study", "изучить основы Go")
	afterCreation := time.Now()
	if err != nil {
		t.Fatalf("NewTask() returned an unexpected error: %v", err)
	}

	if task.GetTitle() != "study" {
		t.Errorf("GetTitle() = %q, want %q", task.GetTitle(), "study")
	}
	if task.GetText() != "изучить основы Go" {
		t.Errorf("GetText() = %q, want %q", task.GetText(), "изучить основы Go")
	}
	if createdAt := task.GetTimeCreation(); createdAt.Before(beforeCreation) || createdAt.After(afterCreation) {
		t.Errorf("GetTimeCreation() = %v, want a value between %v and %v", createdAt, beforeCreation, afterCreation)
	}
	if task.IsDone() {
		t.Error("a new task must not be completed")
	}
	if completedAt, completed := task.GetTimeCompletion(); completed || !completedAt.IsZero() {
		t.Errorf("GetTimeCompletion() = (%v, %v), want a zero time and false", completedAt, completed)
	}
}

func TestNewTaskValidation(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		text        string
		wantMessage string
	}{
		{name: "empty title", title: "", text: "text", wantMessage: "заголовок"},
		{name: "whitespace title", title: " \t ", text: "text", wantMessage: "заголовок"},
		{name: "multiword title", title: "two words", text: "text", wantMessage: "заголовок"},
		{name: "empty text", title: "title", text: "", wantMessage: "текст"},
		{name: "whitespace text", title: "title", text: " \t ", wantMessage: "текст"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := entity.NewTask(tt.title, tt.text)
			if err == nil {
				t.Fatal("NewTask() error = nil, want a validation error")
			}
			if task != nil {
				t.Errorf("NewTask() task = %#v, want nil on validation failure", task)
			}
			if !strings.Contains(err.Error(), tt.wantMessage) {
				t.Errorf("NewTask() error = %q, want it to mention %q", err, tt.wantMessage)
			}
		})
	}
}

func TestTaskStructAddListAndRejectDuplicate(t *testing.T) {
	tasks := entity.NewTaskStruct()
	if err := tasks.Add("zeta", "последняя задача"); err != nil {
		t.Fatalf("Add(zeta) returned an unexpected error: %v", err)
	}
	if err := tasks.Add("alpha", "исходный текст"); err != nil {
		t.Fatalf("Add(alpha) returned an unexpected error: %v", err)
	}

	got := tasks.List()
	if len(got) != 2 {
		t.Fatalf("List() returned %d tasks, want 2", len(got))
	}
	if got[0].GetTitle() != "alpha" || got[1].GetTitle() != "zeta" {
		t.Errorf("List() titles = [%q, %q], want sorted titles [alpha, zeta]", got[0].GetTitle(), got[1].GetTitle())
	}
	originalCreationTime := got[0].GetTimeCreation()

	if err := tasks.Add("alpha", "заменяющий текст"); err == nil {
		t.Fatal("duplicate Add(alpha) error = nil, want an error")
	}

	got = tasks.List()
	if len(got) != 2 {
		t.Fatalf("List() returned %d tasks after duplicate add, want 2", len(got))
	}
	if got[0].GetText() != "исходный текст" {
		t.Errorf("duplicate add replaced task text with %q", got[0].GetText())
	}
	if !got[0].GetTimeCreation().Equal(originalCreationTime) {
		t.Error("duplicate add replaced the original task")
	}
}

func TestTaskStructListReturnsSliceCopy(t *testing.T) {
	tasks := entity.NewTaskStruct()
	if err := tasks.Add("study", "learn Go"); err != nil {
		t.Fatalf("Add() returned an unexpected error: %v", err)
	}

	firstList := tasks.List()
	firstList[0] = nil
	secondList := tasks.List()
	if len(secondList) != 1 || secondList[0] == nil || secondList[0].GetTitle() != "study" {
		t.Errorf("modifying a List() result changed the task store: %#v", secondList)
	}
}

func TestTaskStructDone(t *testing.T) {
	tasks := entity.NewTaskStruct()
	if err := tasks.Add("study", "learn Go"); err != nil {
		t.Fatalf("Add() returned an unexpected error: %v", err)
	}

	beforeCompletion := time.Now()
	if err := tasks.Done("study"); err != nil {
		t.Fatalf("Done() returned an unexpected error: %v", err)
	}
	afterCompletion := time.Now()

	task := tasks.List()[0]
	if !task.IsDone() {
		t.Error("Done() did not mark the task as completed")
	}
	completedAt, completed := task.GetTimeCompletion()
	if !completed {
		t.Fatal("GetTimeCompletion() reports that completion time is not set")
	}
	if completedAt.Before(beforeCompletion) || completedAt.After(afterCompletion) {
		t.Errorf("completion time = %v, want a value between %v and %v", completedAt, beforeCompletion, afterCompletion)
	}

	if err := tasks.Done("study"); err == nil {
		t.Fatal("second Done() error = nil, want an already-completed error")
	}
	afterSecondDone, completed := task.GetTimeCompletion()
	if !completed || !afterSecondDone.Equal(completedAt) {
		t.Errorf("second Done() changed completion time from %v to %v", completedAt, afterSecondDone)
	}
}

func TestTaskStructDoneMissing(t *testing.T) {
	tasks := entity.NewTaskStruct()
	if err := tasks.Done("missing"); err == nil {
		t.Fatal("Done(missing) error = nil, want an error")
	}
	if got := tasks.List(); len(got) != 0 {
		t.Errorf("Done(missing) changed the task store: %#v", got)
	}
}

func TestTaskStructDelete(t *testing.T) {
	tasks := entity.NewTaskStruct()
	if err := tasks.Add("keep", "keep this task"); err != nil {
		t.Fatalf("Add(keep) returned an unexpected error: %v", err)
	}
	if err := tasks.Add("remove", "remove this task"); err != nil {
		t.Fatalf("Add(remove) returned an unexpected error: %v", err)
	}

	if err := tasks.Del("remove"); err != nil {
		t.Fatalf("Del(remove) returned an unexpected error: %v", err)
	}
	got := tasks.List()
	if len(got) != 1 || got[0].GetTitle() != "keep" {
		t.Errorf("List() after delete = %#v, want only task %q", got, "keep")
	}
	if err := tasks.Del("remove"); err == nil {
		t.Fatal("second Del(remove) error = nil, want a not-found error")
	}
	if got := tasks.List(); len(got) != 1 || got[0].GetTitle() != "keep" {
		t.Errorf("failed delete changed the task store: %#v", got)
	}
}
