package agent

import (
	"testing"
	"time"
)

func TestTaskQueueLifecycle(t *testing.T) {
	q := NewTaskQueue(10)
	updateCh := make(chan TaskUpdate, 8)
	q.AddListener(func(update TaskUpdate) {
		updateCh <- update
	})

	task := &Task{ID: "t1", SessionKey: "s1", Type: "chat"}
	q.Add(task)
	q.Start(task.ID)
	q.UpdateProgress(task.ID, 120, "running")
	q.Complete(task.ID)

	got := q.Get(task.ID)
	if got == nil {
		t.Fatalf("expected task, got nil")
	}
	if got.Status != TaskStatusCompleted {
		t.Fatalf("expected completed status, got %s", got.Status)
	}
	if got.Progress != 100 {
		t.Fatalf("expected progress 100, got %d", got.Progress)
	}
	if got.CompletedAt == nil {
		t.Fatalf("expected completed_at to be set")
	}

	// We should observe add/start/progress/complete updates.
	seenCompleted := false
	timeout := time.After(500 * time.Millisecond)
	for i := 0; i < 4; i++ {
		select {
		case u := <-updateCh:
			if u.Status == TaskStatusCompleted {
				seenCompleted = true
			}
		case <-timeout:
			t.Fatalf("timeout waiting for task updates")
		}
	}
	if !seenCompleted {
		t.Fatalf("expected completed update")
	}
}

func TestTaskQueueCapacityAndSessionFilter(t *testing.T) {
	q := NewTaskQueue(2)
	q.Add(&Task{ID: "t1", SessionKey: "s1", Type: "chat"})
	q.Add(&Task{ID: "t2", SessionKey: "s2", Type: "chat"})
	q.Add(&Task{ID: "t3", SessionKey: "s1", Type: "view"})

	if q.Get("t1") != nil {
		t.Fatalf("expected oldest task t1 evicted")
	}

	all := q.List()
	if len(all) != 2 {
		t.Fatalf("expected 2 tasks in queue, got %d", len(all))
	}

	bySession := q.ListBySession("s1")
	if len(bySession) != 1 || bySession[0].ID != "t3" {
		t.Fatalf("unexpected session filter result: %+v", bySession)
	}
}
