package task

import (
	"encoding/json"
	"errors"

	"github.com/syndtr/goleveldb/leveldb"
)

type TaskStore struct {
	db *leveldb.DB
}

func NewTaskStore(dbPath string) (*TaskStore, error) {
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, err
	}

	return &TaskStore{
		db: db,
	}, nil
}

func (ts *TaskStore) Close() error {
	return ts.db.Close()
}

func (ts *TaskStore) AddTask(t Task) error {
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}

	err = ts.db.Put([]byte(t.GetTaskID()), data, nil)
	return err
}

func (ts *TaskStore) GetTask(taskID string) (Task, error) {
	data, err := ts.db.Get([]byte(taskID), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, errors.New("no task found with this ID")
		}
		return nil, err
	}

	var t Task
	err = json.Unmarshal(data, &t)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (ts *TaskStore) RemoveTask(taskID string) error {
	err := ts.db.Delete([]byte(taskID), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return errors.New("no task found with this ID")
		}
		return err
	}
	return nil
}
