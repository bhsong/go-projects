package task

type Storage interface {
	Load() ([]Task, error)
	Save(tasks []Task) error
}

type MemoryStorage struct {
	tasks []Task
}

func NewMemoryStorage(initial []Task) *MemoryStorage {
	cp := make([]Task, len(initial))
	copy(cp, initial)
	return &MemoryStorage{tasks: cp}
}

func (m *MemoryStorage) Load() ([]Task, error) {
	cp := make([]Task, len(m.tasks))
	copy(cp, m.tasks)
	return cp, nil
}

func (m *MemoryStorage) Save(tasks []Task) error {
	cp := make([]Task, len(tasks))
	copy(cp, tasks)
	m.tasks = cp
	return nil
}

var _ Storage = (*MemoryStorage)(nil)
