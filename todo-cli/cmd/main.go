package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/bhsong/go-projects/todo-cli/internal/storage"
	"github.com/bhsong/go-projects/todo-cli/internal/task"
)

const jsonFile string = "tasks.json"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "사용법: todo [add|list|done|delete]")
		os.Exit(1)
	}

	tasks, err := storage.Load(jsonFile)
	if err != nil {
		log.Fatalf("에러: %v", err)
	}

	switch os.Args[1] {
	case "add":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "사용법: todo add [할 일]")
			os.Exit(1)
		}
		tasks = task.Add(tasks, os.Args[2])
		fmt.Println(os.Args[2] + "추가됨")
	case "list":
		if tasks == nil {
			fmt.Fprint(os.Stderr, "할 일이 없습니다.")
			os.Exit(0)
		}

		for _, t := range tasks {
			if t.Done {
				fmt.Print("✅ ")
			} else {
				fmt.Print("⬜ ")
			}
			fmt.Println("[" + strconv.Itoa(t.ID) + "]" + t.Title)
		}
		return
	case "done":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "사용법: todo done [ID]")
			os.Exit(1)
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "오류: %v\n", err)
			os.Exit(1)
		}
		tasks, err = task.Complete(tasks, id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "오류: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("ID %d 완료\n", id)
	case "delete":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "사용법: todo delete [ID]")
			os.Exit(1)
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "오류: %v\n", err)
			os.Exit(1)
		}
		tasks, err = task.Delete(tasks, id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "오류: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("ID %d 삭제\n", id)
	default:
		fmt.Fprint(os.Stderr, "알 수 없는 명령입니다.")
		os.Exit(1)
	}
	err = storage.Save(jsonFile, tasks)
	if err != nil {
		log.Fatalf("에러: %v", err)
	}
}
