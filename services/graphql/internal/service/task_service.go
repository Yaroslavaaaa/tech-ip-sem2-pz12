package service

import (
	"context"
	"fmt"
	"time"

	"graphql-service/internal/repository"

	"github.com/google/uuid"
)

type TaskService struct {
	repo *repository.TaskRepository
}

func NewTaskService(repo *repository.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) GetAllTasks(ctx context.Context) ([]*repository.Task, error) {
	return s.repo.GetAll(ctx)
}

func (s *TaskService) GetTaskByID(ctx context.Context, id string) (*repository.Task, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	if task == nil {
		return nil, nil
	}
	return task, nil
}

func (s *TaskService) CreateTask(ctx context.Context, title, description, dueDate string) (*repository.Task, error) {
	now := time.Now()
	task := &repository.Task{
		ID:          "t_" + uuid.New().String()[:8],
		Title:       title,
		Description: description,
		DueDate:     dueDate,
		Done:        false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return task, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, id string, title, description, dueDate *string, done *bool) (*repository.Task, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	if task == nil {
		return nil, fmt.Errorf("task not found")
	}

	if title != nil {
		task.Title = *title
	}
	if description != nil {
		task.Description = *description
	}
	if dueDate != nil {
		task.DueDate = *dueDate
	}
	if done != nil {
		task.Done = *done
	}
	task.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return task, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, id string) (bool, error) {
	if err := s.repo.Delete(ctx, id); err != nil {
		if err.Error() == "task not found" {
			return false, nil
		}
		return false, fmt.Errorf("failed to delete task: %w", err)
	}
	return true, nil
}
