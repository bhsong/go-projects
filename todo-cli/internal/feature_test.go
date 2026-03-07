// feature_test.go: 여러 함수가 협력해서 하나의 기능을 완성하는지 검증.
// 실제 파일 I/O는 t.TempDir()로 격리 — 테스트 후 자동 정리.
package feature_test

import (
	"path/filepath"
	"testing"

	"github.com/bhsong/go-projects/todo-cli/internal/storage"
	"github.com/bhsong/go-projects/todo-cli/internal/task"
)

// F-01: 추가 후 목록 확인 — Add → Save → Load
func TestFeature_AddAndReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")

	tasks, err := storage.Load(path)
	if err != nil {
		t.Fatalf("초기 Load 실패: %v", err)
	}

	tasks = task.Add(tasks, "PR 리뷰하기")

	if err := storage.Save(path, tasks); err != nil {
		t.Fatalf("Save 실패: %v", err)
	}

	loaded, err := storage.Load(path)
	if err != nil {
		t.Fatalf("재로드 실패: %v", err)
	}

	if len(loaded) != 1 {
		t.Fatalf("길이 1 기대, 실제: %d", len(loaded))
	}
	if loaded[0].Title != "PR 리뷰하기" {
		t.Errorf("Title 불일치: %s", loaded[0].Title)
	}
	if loaded[0].Done != false {
		t.Error("새 항목은 Done=false여야 함")
	}
}

// F-02: 완료 후 목록 확인 — Complete → Save → Load
func TestFeature_CompleteAndReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")

	// 선행 조건: 항목 1개 저장
	initial := task.Add([]task.Task{}, "배포 스크립트 수정")
	if err := storage.Save(path, initial); err != nil {
		t.Fatalf("초기 Save 실패: %v", err)
	}

	tasks, err := storage.Load(path)
	if err != nil {
		t.Fatalf("Load 실패: %v", err)
	}

	tasks, err = task.Complete(tasks, 1)
	if err != nil {
		t.Fatalf("Complete 실패: %v", err)
	}

	if err := storage.Save(path, tasks); err != nil {
		t.Fatalf("Save 실패: %v", err)
	}

	loaded, err := storage.Load(path)
	if err != nil {
		t.Fatalf("재로드 실패: %v", err)
	}

	if len(loaded) != 1 {
		t.Fatalf("길이 1 기대, 실제: %d", len(loaded))
	}
	if !loaded[0].Done {
		t.Error("Done=true로 저장되어야 함")
	}
}

// F-03: 삭제 후 목록 확인 — Delete → Save → Load
func TestFeature_DeleteAndReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")

	// 선행 조건: 항목 2개 저장
	initial := []task.Task{}
	initial = task.Add(initial, "할 일 A")
	initial = task.Add(initial, "할 일 B")
	if err := storage.Save(path, initial); err != nil {
		t.Fatalf("초기 Save 실패: %v", err)
	}

	tasks, err := storage.Load(path)
	if err != nil {
		t.Fatalf("Load 실패: %v", err)
	}

	tasks, err = task.Delete(tasks, 1)
	if err != nil {
		t.Fatalf("Delete 실패: %v", err)
	}

	if err := storage.Save(path, tasks); err != nil {
		t.Fatalf("Save 실패: %v", err)
	}

	loaded, err := storage.Load(path)
	if err != nil {
		t.Fatalf("재로드 실패: %v", err)
	}

	if len(loaded) != 1 {
		t.Fatalf("길이 1 기대, 실제: %d", len(loaded))
	}
	// ID=1이 삭제되었으므로 ID=2만 남아야 함
	if loaded[0].ID != 2 {
		t.Errorf("ID 2가 남아야 함, 실제: %d", loaded[0].ID)
	}
	if loaded[0].Title != "할 일 B" {
		t.Errorf("Title '할 일 B' 기대, 실제: %s", loaded[0].Title)
	}
}

// F-04: 프로그램 재시작 시뮬레이션 — 저장된 3개 항목이 그대로 복원됨
func TestFeature_PersistenceRestart(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")

	// "첫 번째 세션": 3개 항목 저장
	session1 := []task.Task{}
	session1 = task.Add(session1, "항목 1")
	session1 = task.Add(session1, "항목 2")
	session1 = task.Add(session1, "항목 3")
	if err := storage.Save(path, session1); err != nil {
		t.Fatalf("첫 번째 세션 Save 실패: %v", err)
	}

	// "두 번째 세션": 새로 Load해서 항목 수 확인
	session2, err := storage.Load(path)
	if err != nil {
		t.Fatalf("두 번째 세션 Load 실패: %v", err)
	}

	if len(session2) != 3 {
		t.Fatalf("이전에 저장한 3개 항목 복원 기대, 실제: %d", len(session2))
	}
	titles := []string{"항목 1", "항목 2", "항목 3"}
	for i, tk := range session2 {
		if tk.Title != titles[i] {
			t.Errorf("[%d] Title 불일치: 기대 '%s', 실제 '%s'", i, titles[i], tk.Title)
		}
	}
}

// F-05: 없는 ID 완료 시도 — error 반환, 파일 내용 변경 없음
func TestFeature_CompleteNotFound(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")

	// 선행 조건: 항목 1개 저장
	initial := task.Add([]task.Task{}, "존재하는 항목")
	if err := storage.Save(path, initial); err != nil {
		t.Fatalf("초기 Save 실패: %v", err)
	}

	tasks, err := storage.Load(path)
	if err != nil {
		t.Fatalf("Load 실패: %v", err)
	}

	_, err = task.Complete(tasks, 999)
	if err == nil {
		t.Fatal("없는 ID 완료 시도 -> 에러 반환해야 함")
	}

	// 에러 발생 시 Save를 호출하지 않으므로 파일 내용 변경 없음
	reloaded, err := storage.Load(path)
	if err != nil {
		t.Fatalf("재로드 실패: %v", err)
	}
	if len(reloaded) != 1 {
		t.Errorf("파일 내용 변경 없어야 함: 길이 1 기대, 실제: %d", len(reloaded))
	}
	if reloaded[0].Done != false {
		t.Error("파일의 Done이 변경되면 안 됨")
	}
}

// F-06: 없는 ID 삭제 시도 — error 반환, 파일 내용 변경 없음
func TestFeature_DeleteNotFound(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")

	// 선행 조건: 항목 1개 저장
	initial := task.Add([]task.Task{}, "존재하는 항목")
	if err := storage.Save(path, initial); err != nil {
		t.Fatalf("초기 Save 실패: %v", err)
	}

	tasks, err := storage.Load(path)
	if err != nil {
		t.Fatalf("Load 실패: %v", err)
	}

	_, err = task.Delete(tasks, 999)
	if err == nil {
		t.Fatal("없는 ID 삭제 시도 -> 에러 반환해야 함")
	}

	// 에러 발생 시 Save를 호출하지 않으므로 파일 내용 변경 없음
	reloaded, err := storage.Load(path)
	if err != nil {
		t.Fatalf("재로드 실패: %v", err)
	}
	if len(reloaded) != 1 {
		t.Errorf("파일 내용 변경 없어야 함: 길이 1 기대, 실제: %d", len(reloaded))
	}
}

// F-07: 연속 추가 시 ID 중복 없음 — ID가 1, 2, 3으로 순차 할당
func TestFeature_SequentialAddNoDuplicateID(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")

	tasks, _ := storage.Load(path)

	for i, title := range []string{"할 일 1", "할 일 2", "할 일 3"} {
		tasks = task.Add(tasks, title)
		if err := storage.Save(path, tasks); err != nil {
			t.Fatalf("[%d] Save 실패: %v", i, err)
		}
		// 매번 Load해서 실제 저장 상태 기준으로 nextID 계산
		tasks, _ = storage.Load(path)
	}

	if len(tasks) != 3 {
		t.Fatalf("길이 3 기대, 실제: %d", len(tasks))
	}

	seen := make(map[int]bool)
	for _, tk := range tasks {
		if seen[tk.ID] {
			t.Errorf("ID 중복 발견: %d", tk.ID)
		}
		seen[tk.ID] = true
	}

	for i, tk := range tasks {
		if tk.ID != i+1 {
			t.Errorf("인덱스 %d: ID %d 기대, 실제: %d", i, i+1, tk.ID)
		}
	}
}

// F-09: 여러 항목 중 일부만 완료 후 나머지 항목 상태 유지 확인
// runComplete가 구현할 동작: 대상 ID만 Done=true, 나머지는 Done=false 유지
func TestFeature_일부완료후나머지항목유지(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")

	// 선행 조건: 항목 3개 저장
	initial := []task.Task{}
	initial = task.Add(initial, "할 일 A")
	initial = task.Add(initial, "할 일 B")
	initial = task.Add(initial, "할 일 C")
	if err := storage.Save(path, initial); err != nil {
		t.Fatalf("초기 Save 실패: %v", err)
	}

	// ID=2만 완료 처리
	tasks, err := storage.Load(path)
	if err != nil {
		t.Fatalf("Load 실패: %v", err)
	}
	tasks, err = task.Complete(tasks, 2)
	if err != nil {
		t.Fatalf("Complete 실패: %v", err)
	}
	if err := storage.Save(path, tasks); err != nil {
		t.Fatalf("Save 실패: %v", err)
	}

	loaded, err := storage.Load(path)
	if err != nil {
		t.Fatalf("재로드 실패: %v", err)
	}
	if len(loaded) != 3 {
		t.Fatalf("길이 3 기대, 실제: %d", len(loaded))
	}
	// ID=1, ID=3은 Done=false 유지
	for _, tk := range loaded {
		if tk.ID == 2 {
			if !tk.Done {
				t.Errorf("ID=2는 Done=true여야 함")
			}
		} else {
			if tk.Done {
				t.Errorf("ID=%d는 Done=false여야 함 (변경 없어야 함)", tk.ID)
			}
		}
	}
}

// F-10: 삭제 후 새 항목 추가 시 삭제된 ID 재사용 없음
// runDelete + runAdd가 구현할 동작: nextID는 현재 최대 ID 기준이므로 삭제된 ID를 재사용하지 않음
func TestFeature_삭제후추가ID재사용없음(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")

	// 선행 조건: 항목 3개 저장 (ID: 1, 2, 3)
	initial := []task.Task{}
	initial = task.Add(initial, "항목 A") // ID=1
	initial = task.Add(initial, "항목 B") // ID=2
	initial = task.Add(initial, "항목 C") // ID=3
	if err := storage.Save(path, initial); err != nil {
		t.Fatalf("초기 Save 실패: %v", err)
	}

	// ID=2 삭제
	tasks, err := storage.Load(path)
	if err != nil {
		t.Fatalf("Load 실패: %v", err)
	}
	tasks, err = task.Delete(tasks, 2)
	if err != nil {
		t.Fatalf("Delete 실패: %v", err)
	}
	if err := storage.Save(path, tasks); err != nil {
		t.Fatalf("Save 실패: %v", err)
	}

	// 새 항목 추가 — nextID는 현재 최대 ID(3)+1=4여야 함, 2를 재사용하면 안 됨
	tasks, err = storage.Load(path)
	if err != nil {
		t.Fatalf("재로드 실패: %v", err)
	}
	tasks = task.Add(tasks, "새 항목")
	if err := storage.Save(path, tasks); err != nil {
		t.Fatalf("Save 실패: %v", err)
	}

	loaded, err := storage.Load(path)
	if err != nil {
		t.Fatalf("최종 Load 실패: %v", err)
	}
	if len(loaded) != 3 {
		t.Fatalf("길이 3 기대, 실제: %d", len(loaded))
	}

	// 새로 추가된 항목의 ID 확인
	newTask := loaded[len(loaded)-1]
	if newTask.Title != "새 항목" {
		t.Fatalf("마지막 항목이 '새 항목'이어야 함, 실제: %s", newTask.Title)
	}
	if newTask.ID == 2 {
		t.Errorf("삭제된 ID(2)를 재사용하면 안 됨, 실제 ID: %d", newTask.ID)
	}
	if newTask.ID != 4 {
		t.Errorf("새 항목 ID=4 기대 (max 3 + 1), 실제: %d", newTask.ID)
	}

	// ID 중복 없는지 검증
	seen := make(map[int]bool)
	for _, tk := range loaded {
		if seen[tk.ID] {
			t.Errorf("ID 중복 발견: %d", tk.ID)
		}
		seen[tk.ID] = true
	}
}

// F-08: 완료는 멱등성을 가짐 — Done=true 항목에 Complete 재실행 시 error 없음, Done=true 유지
func TestFeature_CompleteIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")

	// 선행 조건: Done=true 항목 준비
	initial := task.Add([]task.Task{}, "이미 완료된 항목")
	initial, err := task.Complete(initial, 1)
	if err != nil {
		t.Fatalf("첫 번째 Complete 실패: %v", err)
	}
	if err := storage.Save(path, initial); err != nil {
		t.Fatalf("초기 Save 실패: %v", err)
	}

	tasks, err := storage.Load(path)
	if err != nil {
		t.Fatalf("Load 실패: %v", err)
	}

	// 두 번째 Complete 실행
	tasks, err = task.Complete(tasks, 1)
	if err != nil {
		t.Fatalf("이미 완료된 항목 재완료 -> 에러 없어야 함: %v", err)
	}
	if !tasks[0].Done {
		t.Error("Done=true 유지되어야 함")
	}
}
