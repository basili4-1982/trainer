package task

import (
	"context"
	"io"
	"strings"
	"time"
	"unicode"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

const defaultRunner = "php:7-cli"

var containers = map[string]string{
	"php7": "php:7-cli",
	"php8": "php:8-cli",
}

// Task представляет задачу, которую нужно выполнить
type Task struct {
	Name   string
	Code   string
	Runner string
}

// Result хранит результат выполнения задачи
type Result struct {
	TaskName string
	Success  bool
	Error    error
	Output   string
}

// RunPHPInDocker выполняет PHP код в Docker контейнере
func RunPHPInDocker(ctx context.Context, code, containerName string) (string, error) {
	cli, err := client.NewClientWithOpts(client.WithVersion("1.48"))
	if err != nil {
		return "", err
	}

	// Создание нового контейнера с PHP
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: containerName,
		Cmd:   []string{"php", "-r", code},
		Tty:   false,
	}, nil, nil, nil, "")
	if err != nil {
		return "", err
	}

	// Запуск контейнера
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", err
	}

	time.Sleep(5 * time.Second)

	defer cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{
		Force: true,
	})

	// Получение вывода контейнера

	reader, err := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, Follow: false})
	if err != nil {
		return "", err
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return cleanString(string(content)), nil
}

// TaskRunner выполняет проверку отдельной задачи
func TaskRunner(ctx context.Context, task Task) Result {

	containerName, ok := containers[task.Runner]

	if !ok {
		containerName = defaultRunner
	}

	output, err := RunPHPInDocker(ctx, task.Code, containerName)
	if err != nil {
		return Result{TaskName: task.Name, Success: false, Error: err}
	}

	return Result{TaskName: task.Name, Success: true, Output: output}
}

// RunTask запускает все задачу
func RunTask(ctx context.Context, task Task) Result {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	return TaskRunner(ctx, task)
}

func cleanString(s string) string {
	s = strings.TrimSpace(s)

	return strings.Map(func(r rune) rune {
		if unicode.IsGraphic(r) || r == ' ' { // оставляем пробелы
			return r
		}
		return -1 // удаляем символ
	}, s)
}
