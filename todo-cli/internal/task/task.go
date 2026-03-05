package task

import (
	"fmt"
	"time"
)

type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"createdat"`
}

func Add(tasks []Task, title string) []Task {
	id := nextID(tasks)

	t := Task{ID: id, Title: title}

	tasks = append(tasks, t)

	return tasks
}

func Complete(tasks []Task, id int) ([]Task, error) {
	for i := range tasks {
		if tasks[i].ID == id {
			tasks[i].Done = true
			return tasks, nil
		}
	}
	return tasks, fmt.Errorf("task.Complete: ID %d 없음", id)
}

func Delete(tasks []Task, id int) ([]Task, error) {
	for i := range tasks {
		if tasks[i].ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			return tasks, nil
		}
	}
	return tasks, fmt.Errorf("task.Delete: ID %d 없음", id)
}

func nextID(tasks []Task) int {
	max := 0
	for _, t := range tasks {
		if t.ID > max {
			max = t.ID
		}
	}
	return max + 1
}
