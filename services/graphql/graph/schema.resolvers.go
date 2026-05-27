package graph

import (
	"context"
	"graphql-service/graph/model"
	"time"
)

// CreateTask реализует Mutation.createTask
func (r *mutationResolver) CreateTask(ctx context.Context, input model.CreateTaskInput) (*model.Task, error) {
	dueDate := ""
	if input.DueDate != nil {
		dueDate = *input.DueDate
	}
	description := ""
	if input.Description != nil {
		description = *input.Description
	}

	task, err := r.TaskService.CreateTask(ctx, input.Title, description, dueDate)
	if err != nil {
		return nil, err
	}

	desc := task.Description
	due := task.DueDate

	return &model.Task{
		ID:          task.ID,
		Title:       task.Title,
		Description: &desc,
		DueDate:     &due,
		Done:        task.Done,
		CreatedAt:   task.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   task.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// UpdateTask реализует Mutation.updateTask
func (r *mutationResolver) UpdateTask(ctx context.Context, id string, input model.UpdateTaskInput) (*model.Task, error) {
	task, err := r.TaskService.UpdateTask(ctx, id, input.Title, input.Description, input.DueDate, input.Done)
	if err != nil {
		return nil, err
	}

	desc := task.Description
	due := task.DueDate

	return &model.Task{
		ID:          task.ID,
		Title:       task.Title,
		Description: &desc,
		DueDate:     &due,
		Done:        task.Done,
		CreatedAt:   task.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   task.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// DeleteTask реализует Mutation.deleteTask
func (r *mutationResolver) DeleteTask(ctx context.Context, id string) (bool, error) {
	return r.TaskService.DeleteTask(ctx, id)
}

// Tasks реализует Query.tasks
func (r *queryResolver) Tasks(ctx context.Context) ([]*model.Task, error) {
	tasks, err := r.TaskService.GetAllTasks(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Task, 0, len(tasks))
	for _, task := range tasks {
		desc := task.Description
		due := task.DueDate
		result = append(result, &model.Task{
			ID:          task.ID,
			Title:       task.Title,
			Description: &desc,
			DueDate:     &due,
			Done:        task.Done,
			CreatedAt:   task.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   task.UpdatedAt.Format(time.RFC3339),
		})
	}
	return result, nil
}

// Task реализует Query.task
func (r *queryResolver) Task(ctx context.Context, id string) (*model.Task, error) {
	task, err := r.TaskService.GetTaskByID(ctx, id)
	if err != nil {
		return nil, nil
	}
	if task == nil {
		return nil, nil
	}

	desc := task.Description
	due := task.DueDate

	return &model.Task{
		ID:          task.ID,
		Title:       task.Title,
		Description: &desc,
		DueDate:     &due,
		Done:        task.Done,
		CreatedAt:   task.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   task.UpdatedAt.Format(time.RFC3339),
	}, nil
}
