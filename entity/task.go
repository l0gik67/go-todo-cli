package entity

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// TaskStruct stores tasks by their unique titles.
type TaskStruct struct {
	tasks map[string]*Task
}

// Task describes one todo item.
type Task struct {
	title          string
	text           string
	timeCreation   time.Time
	isDone         bool
	timeCompletion *time.Time
}

func NewTaskStruct() *TaskStruct {
	return &TaskStruct{
		tasks: make(map[string]*Task),
	}
}

func NewTask(title, text string) (*Task, error) {
	if len(strings.Fields(title)) != 1 {
		return nil, errors.New("заголовок задачи должен состоять из одного слова")
	}
	if strings.TrimSpace(text) == "" {
		return nil, errors.New("текст задачи не должен быть пустым")
	}

	return &Task{
		title:        title,
		text:         text,
		timeCreation: time.Now(),
	}, nil
}

func (t *Task) GetTitle() string {
	return t.title
}

func (t *Task) GetText() string {
	return t.text
}

func (t *Task) GetTimeCreation() time.Time {
	return t.timeCreation
}

func (t *Task) IsDone() bool {
	return t.isDone
}

// GetTimeCompletion returns the completion time and whether it was set.
func (t *Task) GetTimeCompletion() (time.Time, bool) {
	if t.timeCompletion == nil {
		return time.Time{}, false
	}
	return *t.timeCompletion, true
}

func (t *TaskStruct) Add(title, text string) error {
	if _, exists := t.tasks[title]; exists {
		return fmt.Errorf("задача с заголовком %s уже существует", title)
	}

	task, err := NewTask(title, text)
	if err != nil {
		return err
	}
	t.tasks[task.GetTitle()] = task
	return nil
}

// List returns all tasks sorted by title.
func (t *TaskStruct) List() []*Task {
	tasks := make([]*Task, 0, len(t.tasks))
	for _, task := range t.tasks {
		tasks = append(tasks, task)
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].GetTitle() < tasks[j].GetTitle()
	})
	return tasks
}

func (t *TaskStruct) Del(title string) error {
	if _, exists := t.tasks[title]; !exists {
		return fmt.Errorf("задача с заголовком %s не найдена", title)
	}
	delete(t.tasks, title)
	return nil
}

func (t *TaskStruct) Done(title string) error {
	task, exists := t.tasks[title]
	if !exists {
		return fmt.Errorf("задача с заголовком %s не найдена", title)
	}
	if task.isDone {
		return fmt.Errorf("задача с заголовком %s уже выполнена", title)
	}

	now := time.Now()
	task.isDone = true
	task.timeCompletion = &now
	return nil
}
