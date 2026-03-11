package task

import (
	"fmt"
	"time"
)

type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	Priority  string    `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
}

type FilterOptions struct {
	ShowAll      bool   // true: 완료 포함 전체, false: 미완료만(기본값)
	ShowDoneOnly bool   // true: 완료만
	Priority     string // ""이면 필터 없음. "high","normal","low"이면 해당만 출력
}

var validPriorities = map[string]bool{
	"high":   true,
	"normal": true,
	"low":    true,
}

func Add(tasks []Task, title string, priority string) ([]Task, error) {
	if title == "" {
		return nil, fmt.Errorf("task.Add: 제목이 비어있습니다.")
	}

	if _, exist := validPriorities[priority]; !exist {
		return nil, fmt.Errorf("task.Add: 잘못된 우선수위입니다: %s (high/normal/low)", priority)
	}

	id := nextID(tasks)

	t := Task{
		ID:        id,
		Title:     title,
		Done:      false,
		Priority:  priority,
		CreatedAt: time.Now(),
	}

	tasks = append(tasks, t)

	return tasks, nil
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

func FilterTasks(tasks []Task, opts FilterOptions) []Task {
	result := []Task{} // 2차 필터 통과 후 태스크들
	for _, t := range tasks {
		if opts.ShowDoneOnly && !t.Done {
			continue
		} else if !opts.ShowAll && !opts.ShowDoneOnly && t.Done {
			continue
		}
		if opts.Priority == "" || t.Priority == opts.Priority {
			result = append(result, t)
		}
	}
	return result
}

func nextID(tasks []Task) int {
	maxID := 0
	for _, t := range tasks {
		if t.ID > maxID {
			maxID = t.ID
		}
	}
	return maxID + 1
}

func IsValidPriority(p string) bool {
	return validPriorities[p]
}
