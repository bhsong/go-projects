package main

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/bhsong/go-projects/todo-cli/internal/storage"
	"github.com/bhsong/go-projects/todo-cli/internal/task"
)

const dataFile string = "tasks.json"

func main() {
	s := storage.NewJSONStorage(dataFile)
	if len(os.Args) < 2 {
		printUsage(os.Stderr)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "add":
		err := runAdd(s, os.Stdout, os.Args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "오류: %v\n", err)
			os.Exit(1)
		}
	case "list":
		err := runList(s, os.Stdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "오류: %v\n", err)
			os.Exit(1)
		}
	case "done":
		err := runDone(s, os.Stdout, os.Args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "오류: %v\n", err)
			os.Exit(1)
		}
	case "delete":
		err := runDelete(s, os.Stdout, os.Args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "오류: %v\n", err)
			os.Exit(1)
		}
	default:
		printUsage(os.Stderr)
		os.Exit(1)
	}
}

func runAdd(s task.Storage, w io.Writer, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("사용법: todo add <할 일>")
	}

	title := args[2]

	tasks, err := s.Load()
	if err != nil {
		return fmt.Errorf("runAdd: %w", err)
	}

	tasks, err = task.Add(tasks, title)
	if err != nil {
		return fmt.Errorf("runAdd: %w", err)
	}

	err = s.Save(tasks)
	if err != nil {
		return fmt.Errorf("runAdd: %w", err)
	}

	task.PrintResult(w, "✅ 추가됨: "+title)

	return nil
}

func runList(s task.Storage, w io.Writer) error {
	tasks, err := s.Load()
	if err != nil {
		return fmt.Errorf("runList: %w", err)
	}

	task.PrintTasks(w, tasks)

	return nil
}

func runDone(s task.Storage, w io.Writer, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("사용법: todo done <ID>")
	}
	id, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("ID는 숫자여야 합니다: %q", args[2])
	}
	tasks, err := s.Load()
	if err != nil {
		return fmt.Errorf("runDone: %w", err)
	}
	tasks, err = task.Complete(tasks, id)
	if err != nil {
		return fmt.Errorf("runDone: %w", err)
	}
	err = s.Save(tasks)
	if err != nil {
		return fmt.Errorf("runDone: %w", err)
	}
	task.PrintResult(w, fmt.Sprintf("✅ 완료 처리: ID %d", id))
	return nil
}

func runDelete(s task.Storage, w io.Writer, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("사용법: todo delete <ID>")
	}
	id, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("ID는 숫자여야 합니다: %q", args[2])
	}
	tasks, err := s.Load()
	if err != nil {
		return fmt.Errorf("runDelete: %w", err)
	}
	tasks, err = task.Delete(tasks, id)
	if err != nil {
		return fmt.Errorf("runDelete: %w", err)
	}
	err = s.Save(tasks)
	if err != nil {
		return fmt.Errorf("runDelete: %w", err)
	}
	task.PrintResult(w, fmt.Sprintf("🗑️  삭제됨: ID %d", id))
	return nil
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "사용법: todo [add|list|done|delete]")
}
