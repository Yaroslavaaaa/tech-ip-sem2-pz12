package graph

import (
	"graphql-service/internal/service"
)

// Resolver - структура для внедрения зависимостей
type Resolver struct {
	TaskService *service.TaskService
}

// NewResolver создаёт новый резолвер
func NewResolver(taskService *service.TaskService) *Resolver {
	return &Resolver{
		TaskService: taskService,
	}
}

// Mutation возвращает MutationResolver
func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

// Query возвращает QueryResolver
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct {
	*Resolver
}

type queryResolver struct {
	*Resolver
}
