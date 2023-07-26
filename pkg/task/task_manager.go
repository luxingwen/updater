package task

import (
	"errors"
	"sync"
)

type TaskManager struct {
	tasks map[string]Task
	mu    sync.RWMutex
}

func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks: make(map[string]Task),
	}
}

func (tm *TaskManager) AddTask(t Task) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.tasks[t.GetTaskID()]; exists {
		return errors.New("task with this ID already exists")
	}

	tm.tasks[t.GetTaskID()] = t
	return nil
}

func (tm *TaskManager) GetTask(taskID string) (Task, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	task, exists := tm.tasks[taskID]
	if !exists {
		return nil, errors.New("no task found with this ID")
	}

	return task, nil
}

func (tm *TaskManager) RemoveTask(taskID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.tasks[taskID]; !exists {
		return errors.New("no task found with this ID")
	}

	delete(tm.tasks, taskID)
	return nil
}

func (tm *TaskManager) GetAllTasks() map[string]Task {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	// Make a copy of the map to avoid race conditions
	copiedTasks := make(map[string]Task, len(tm.tasks))
	for id, task := range tm.tasks {
		copiedTasks[id] = task
	}

	return copiedTasks
}
