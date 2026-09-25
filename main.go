package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/k0kubun/pp"
	"github.com/l0gik67/go-todo-cli/entity"
)

const helpText = `Доступные команды:

  help
    Показать доступные команды и их формат.

  add {заголовок} {текст}
    Добавить новую задачу.
    Заголовок должен состоять из одного слова.

  list
    Показать полный список задач.

  del {заголовок}
    Удалить задачу по заголовку.

  done {заголовок}
    Отметить задачу как выполненную.

  events
    Показать список всех событий.

  exit
    Завершить выполнение программы.
`

type taskOutput struct {
	Title       string
	Text        string
	CreatedAt   string
	Status      string
	CompletedAt string
}

type eventOutput struct {
	Text        string
	Description string
	CreatedAt   string
}

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка чтения ввода:", err)
	}
}

func run(input io.Reader, output io.Writer) error {
	taskStruct := entity.NewTaskStruct()
	eventStruct := entity.NewEventStruct()
	reader := bufio.NewReader(input)

	// pp is used for readable structured output without terminal escape codes.
	pp.ColoringEnabled = false

	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			line = strings.TrimSuffix(line, "\n")
			line = strings.TrimSuffix(line, "\r")
			if checkCommand(line, taskStruct, eventStruct, output) {
				return nil
			}
		}

		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// checkCommand handles one input line and returns true only for a valid exit command.
func checkCommand(
	rawInput string,
	taskStruct *entity.TaskStruct,
	eventStruct *entity.EventStruct,
	output io.Writer,
) bool {
	params := strings.Fields(rawInput)
	var commandErr error
	shouldExit := false
	printEventList := false

	if len(params) == 0 {
		commandErr = errors.New("команда не указана")
	} else {
		switch params[0] {
		case "help":
			commandErr = requireArgumentCount(params, 1, "help")
			if commandErr == nil {
				fmt.Fprint(output, helpText)
			}
		case "add":
			if len(params) < 3 {
				commandErr = errors.New("формат команды: add {заголовок} {текст}")
			} else {
				commandErr = taskStruct.Add(params[1], strings.Join(params[2:], " "))
			}
		case "list":
			commandErr = requireArgumentCount(params, 1, "list")
			if commandErr == nil {
				printTasks(output, taskStruct.List())
			}
		case "del":
			commandErr = requireArgumentCount(params, 2, "del {заголовок}")
			if commandErr == nil {
				commandErr = taskStruct.Del(params[1])
			}
		case "done":
			commandErr = requireArgumentCount(params, 2, "done {заголовок}")
			if commandErr == nil {
				commandErr = taskStruct.Done(params[1])
			}
		case "events":
			commandErr = requireArgumentCount(params, 1, "events")
			printEventList = commandErr == nil
		case "exit":
			commandErr = requireArgumentCount(params, 1, "exit")
			shouldExit = commandErr == nil
		default:
			commandErr = fmt.Errorf("неизвестная команда: %s", params[0])
		}
	}

	eventStruct.Add(entity.NewEvent(rawInput, commandErr))
	if commandErr != nil {
		fmt.Fprintln(output, "Ошибка:", commandErr)
	}
	if printEventList {
		printEvents(output, eventStruct.List())
	}

	return shouldExit
}

func requireArgumentCount(params []string, expected int, format string) error {
	if len(params) != expected {
		return fmt.Errorf("формат команды: %s", format)
	}
	return nil
}

func printTasks(output io.Writer, tasks []*entity.Task) {
	if len(tasks) == 0 {
		fmt.Fprintln(output, "Список задач пуст.")
		return
	}

	fmt.Fprintln(output, "Список задач:")
	for _, task := range tasks {
		completedAt := "—"
		if completionTime, completed := task.GetTimeCompletion(); completed {
			completedAt = formatTime(completionTime)
		}
		status := "не выполнена"
		if task.IsDone() {
			status = "выполнена"
		}
		_, _ = pp.Fprintln(output, taskOutput{
			Title:       task.GetTitle(),
			Text:        task.GetText(),
			CreatedAt:   formatTime(task.GetTimeCreation()),
			Status:      status,
			CompletedAt: completedAt,
		})
	}
}

func printEvents(output io.Writer, events []*entity.Event) {
	fmt.Fprintln(output, "Список событий:")
	for _, event := range events {
		_, _ = pp.Fprintln(output, eventOutput{
			Text:        event.GetText(),
			Description: event.GetDescription(),
			CreatedAt:   formatTime(event.GetCreationTime()),
		})
	}
}

func formatTime(value time.Time) string {
	return value.Format(time.RFC3339Nano)
}
