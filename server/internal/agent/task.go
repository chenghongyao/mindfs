package agent

import (
	"sync"
	"time"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// Task represents an agent task
type Task struct {
	ID          string                 `json:"id"`
	SessionKey  string                 `json:"session_key"`
	Type        string                 `json:"type"` // chat, view, skill
	Status      TaskStatus             `json:"status"`
	Progress    int                    `json:"progress"` // 0-100
	Message     string                 `json:"message,omitempty"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TaskUpdate represents a task status update
type TaskUpdate struct {
	TaskID   string     `json:"task_id"`
	Status   TaskStatus `json:"status"`
	Progress int        `json:"progress"`
	Message  string     `json:"message,omitempty"`
	Error    string     `json:"error,omitempty"`
}

// TaskListener is called when a task is updated
type TaskListener func(update TaskUpdate)

// TaskQueue manages a queue of tasks
type TaskQueue struct {
	mu        sync.RWMutex
	tasks     map[string]*Task
	queue     []string // task IDs in order
	listeners []TaskListener
	maxTasks  int
}

// NewTaskQueue creates a new task queue
func NewTaskQueue(maxTasks int) *TaskQueue {
	if maxTasks <= 0 {
		maxTasks = 100
	}
	return &TaskQueue{
		tasks:    make(map[string]*Task),
		queue:    make([]string, 0),
		maxTasks: maxTasks,
	}
}

// AddListener adds a task update listener
func (q *TaskQueue) AddListener(listener TaskListener) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.listeners = append(q.listeners, listener)
}

// notifyListeners notifies all listeners of a task update
func (q *TaskQueue) notifyListeners(update TaskUpdate) {
	q.mu.RLock()
	listeners := make([]TaskListener, len(q.listeners))
	copy(listeners, q.listeners)
	q.mu.RUnlock()

	for _, listener := range listeners {
		go listener(update)
	}
}

// Add adds a new task to the queue
func (q *TaskQueue) Add(task *Task) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Remove oldest tasks if queue is full
	for len(q.queue) >= q.maxTasks {
		oldestID := q.queue[0]
		q.queue = q.queue[1:]
		delete(q.tasks, oldestID)
	}

	task.CreatedAt = time.Now()
	task.Status = TaskStatusPending
	q.tasks[task.ID] = task
	q.queue = append(q.queue, task.ID)

	q.notifyListeners(TaskUpdate{
		TaskID:   task.ID,
		Status:   task.Status,
		Progress: task.Progress,
	})
}

// Get returns a task by ID
func (q *TaskQueue) Get(taskID string) *Task {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if task, ok := q.tasks[taskID]; ok {
		// Return a copy
		taskCopy := *task
		return &taskCopy
	}
	return nil
}

// Start marks a task as running
func (q *TaskQueue) Start(taskID string) {
	q.mu.Lock()
	task, ok := q.tasks[taskID]
	if !ok {
		q.mu.Unlock()
		return
	}

	now := time.Now()
	task.Status = TaskStatusRunning
	task.StartedAt = &now
	q.mu.Unlock()

	q.notifyListeners(TaskUpdate{
		TaskID:   taskID,
		Status:   TaskStatusRunning,
		Progress: task.Progress,
	})
}

// UpdateProgress updates task progress
func (q *TaskQueue) UpdateProgress(taskID string, progress int, message string) {
	q.mu.Lock()
	task, ok := q.tasks[taskID]
	if !ok {
		q.mu.Unlock()
		return
	}

	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}

	task.Progress = progress
	task.Message = message
	q.mu.Unlock()

	q.notifyListeners(TaskUpdate{
		TaskID:   taskID,
		Status:   task.Status,
		Progress: progress,
		Message:  message,
	})
}

// Complete marks a task as completed
func (q *TaskQueue) Complete(taskID string) {
	q.mu.Lock()
	task, ok := q.tasks[taskID]
	if !ok {
		q.mu.Unlock()
		return
	}

	now := time.Now()
	task.Status = TaskStatusCompleted
	task.CompletedAt = &now
	task.Progress = 100
	q.mu.Unlock()

	q.notifyListeners(TaskUpdate{
		TaskID:   taskID,
		Status:   TaskStatusCompleted,
		Progress: 100,
	})
}

// Fail marks a task as failed
func (q *TaskQueue) Fail(taskID string, err string) {
	q.mu.Lock()
	task, ok := q.tasks[taskID]
	if !ok {
		q.mu.Unlock()
		return
	}

	now := time.Now()
	task.Status = TaskStatusFailed
	task.CompletedAt = &now
	task.Error = err
	q.mu.Unlock()

	q.notifyListeners(TaskUpdate{
		TaskID:   taskID,
		Status:   TaskStatusFailed,
		Progress: task.Progress,
		Error:    err,
	})
}

// Cancel marks a task as cancelled
func (q *TaskQueue) Cancel(taskID string) {
	q.mu.Lock()
	task, ok := q.tasks[taskID]
	if !ok {
		q.mu.Unlock()
		return
	}

	now := time.Now()
	task.Status = TaskStatusCancelled
	task.CompletedAt = &now
	q.mu.Unlock()

	q.notifyListeners(TaskUpdate{
		TaskID:   taskID,
		Status:   TaskStatusCancelled,
		Progress: task.Progress,
	})
}

// List returns all tasks
func (q *TaskQueue) List() []*Task {
	q.mu.RLock()
	defer q.mu.RUnlock()

	result := make([]*Task, 0, len(q.queue))
	for _, id := range q.queue {
		if task, ok := q.tasks[id]; ok {
			taskCopy := *task
			result = append(result, &taskCopy)
		}
	}
	return result
}

// ListBySession returns tasks for a session
func (q *TaskQueue) ListBySession(sessionKey string) []*Task {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var result []*Task
	for _, id := range q.queue {
		if task, ok := q.tasks[id]; ok && task.SessionKey == sessionKey {
			taskCopy := *task
			result = append(result, &taskCopy)
		}
	}
	return result
}

// GetPending returns pending tasks
func (q *TaskQueue) GetPending() []*Task {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var result []*Task
	for _, id := range q.queue {
		if task, ok := q.tasks[id]; ok && task.Status == TaskStatusPending {
			taskCopy := *task
			result = append(result, &taskCopy)
		}
	}
	return result
}

// GetRunning returns running tasks
func (q *TaskQueue) GetRunning() []*Task {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var result []*Task
	for _, id := range q.queue {
		if task, ok := q.tasks[id]; ok && task.Status == TaskStatusRunning {
			taskCopy := *task
			result = append(result, &taskCopy)
		}
	}
	return result
}
