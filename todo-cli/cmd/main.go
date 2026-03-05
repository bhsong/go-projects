package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/bhsong/go-projects/todo-cli/internal/storage"
	"github.com/bhsong/go-projects/todo-cli/internal/task"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("사용법: todo [add|list|done|delete]")
		os.Exit(1)
	}

	tasks, err := storage.Load("todo.json")
	if err != nil {
		log.Fatalf("에러: %v", err)
	}

	switch os.Args[1] {
	case "Add":
		tasks = task.Add(tasks, os.Args[2])
		if err != nil {
			fmt.Printf("ID를 다시 확인해보세요.")
			os.Exit(1)
		}
	case "list":
		for _, t := range tasks {
			fmt.Println(t.Title)
		}
	case "done":
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("ID는 숫자만 입력해주세요.")
		}
		tasks, err = task.Complete(tasks, id)
		if err != nil {
			fmt.Printf("ID를 다시 확인해보세요.")
			os.Exit(1)
		}
	}
	err = storage.Save("todo.json", tasks)
	if err != nil {
		log.Fatalf("에러: %v", err)
	}
}
