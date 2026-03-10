package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/bhsong/go-projects/todo-cli/internal/storage"
	"github.com/bhsong/go-projects/todo-cli/internal/task"
)

const dataFile = "tasks.json"

func main() {
	s := storage.NewJSONStorage(dataFile)
	if len(os.Args) < 2 {
		printUsage(os.Stderr)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "add":
		err := runAdd(s, os.Stdout, os.Args[2:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "오류: %v\n", err)
			os.Exit(1)
		}
	case "list":
		err := runList(s, os.Stdout, os.Args[2:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "오류: %v\n", err)
			os.Exit(1)
		}
	case "done":
		err := runDone(s, os.Stdout, os.Args[2:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "오류: %v\n", err)
			os.Exit(1)
		}
	case "delete":
		err := runDelete(s, os.Stdout, os.Args[2:])
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
	//if len(args) < 3 {
	//	return fmt.Errorf("사용법: todo add <할 일>")
	//}

	addCmd := flag.NewFlagSet("add", flag.ContinueOnError)
	priority := addCmd.String("priority", "normal", "우선순위 (high/normal/low)")

	if err := addCmd.Parse(args); err != nil {
		return fmt.Errorf("runAdd: 플래그 파싱 실패: %w", err)
	}

	if addCmd.Arg(0) == "" {
		return fmt.Errorf("사용법: todo add [--priority high|normal|low] <할 일>")
	}

	title := addCmd.Arg(0)

	tasks, err := s.Load()
	if err != nil {
		return fmt.Errorf("runAdd: %w", err)
	}

	tasks, err = task.Add(tasks, title, *priority)
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

func runList(s task.Storage, w io.Writer, args []string) error {

	listCmd := flag.NewFlagSet("list", flag.ContinueOnError)
	showAll := listCmd.Bool("all", false, "완료 항목 포함 전체 출력")
	showDoneOnly := listCmd.Bool("done", false, "완료 항목만 출력")
	filterPri := listCmd.String("priority", "", "우선순위 필터 (high/normal/low)")

	if err := listCmd.Parse(args); err != nil {
		return fmt.Errorf("runList: 플래그 파싱 실패: %w", err)
	}

	if *filterPri != "" && !task.IsValidPriority(*filterPri) {
		return fmt.Errorf("runList: 잘못된 우선순위입니다: %s (high/normal/low)", *filterPri)
	}

	tasks, err := s.Load()
	if err != nil {
		return fmt.Errorf("runList: %w", err)
	}

	opts := task.FilterOptions{
		ShowAll:      *showAll,
		ShowDoneOnly: *showDoneOnly,
		Priority:     *filterPri,
	}

	filtered := task.FilterTasks(tasks, opts)
	task.PrintTasks(w, filtered)

	return nil
}

func runDone(s task.Storage, w io.Writer, args []string) error {
	//if len(args) < 3 {
	//		return fmt.Errorf("사용법: todo done <ID>")
	//}

	doneCmd := flag.NewFlagSet("done", flag.ContinueOnError)

	if err := doneCmd.Parse(args); err != nil {
		return fmt.Errorf("runDone: 플래그 파싱 실패:%w", err)
	}

	if doneCmd.Arg(0) == "" {
		return fmt.Errorf("사용법: todo done <ID>")
	}

	id, err := strconv.Atoi(doneCmd.Arg(0))
	if err != nil {
		return fmt.Errorf("ID는 숫자여야 합니다: %s", doneCmd.Arg(0))
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
	task.PrintResult(w, fmt.Sprintf("✅ 완료됨: ID %d", id))
	return nil
}

func runDelete(s task.Storage, w io.Writer, args []string) error {
	//if len(args) < 3 {
	//	return fmt.Errorf("사용법: todo delete <ID>")
	//}

	deleteCmd := flag.NewFlagSet("delete", flag.ContinueOnError)

	if err := deleteCmd.Parse(args); err != nil {
		return fmt.Errorf("runDelete: 플래그 파싱 실패: %w", err)
	}

	if deleteCmd.Arg(0) == "" {
		return fmt.Errorf("사용법: todo delete <ID>")
	}

	id, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("ID는 숫자여야 합니다: %s", deleteCmd.Arg(0))
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
	task.PrintResult(w, fmt.Sprintf("🗑️  삭제됨: [%d]", id))
	return nil
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `사용법: 
											todo add [--priority high|normal|low] <할 일> 
											todo list [--all] [--done] [--priority high|normal|row] 
											todo done <ID> 
											todo delete <ID>`)
}
