package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/mj3smile/bank-statement-processor/internal/model/upload"
	"github.com/mj3smile/bank-statement-processor/internal/repository"
)

type uploadRepository struct {
	mu   sync.RWMutex
	task map[upload.ID]*upload.Task
}

func NewUploadRepository() repository.UploadRepository {
	return &uploadRepository{
		task: make(map[upload.ID]*upload.Task),
	}
}

func (u *uploadRepository) Save(ctx context.Context, uploadTask *upload.Task) error {
	if uploadTask == nil {
		return errors.New("upload task is nil")
	}

	if uploadTask.ID == "" {
		return errors.New("upload id is empty")
	}

	u.mu.Lock()
	defer u.mu.Unlock()
	if _, ok := u.task[uploadTask.ID]; ok {
		return errors.New("upload task already exists")
	}

	u.task[uploadTask.ID] = uploadTask
	return nil
}

func (u *uploadRepository) Update(ctx context.Context, updateValue *upload.Task) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	id := updateValue.ID
	_, ok := u.task[id]
	if !ok {
		return errors.New("upload task not found")
	}

	u.task[id].Message = updateValue.Message
	u.task[id].Status = updateValue.Status
	u.task[id].CompletedAt = updateValue.CompletedAt

	return nil
}

func (u *uploadRepository) GetByID(ctx context.Context, uploadID upload.ID) (*upload.Task, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()
	task, ok := u.task[uploadID]
	if !ok {
		return nil, errors.New("upload ID not found")
	}

	return task, nil
}
