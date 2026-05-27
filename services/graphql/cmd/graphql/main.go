package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"graphql-service/internal/repository"
	"graphql-service/internal/service"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

func main() {
	port := os.Getenv("GRAPHQL_PORT")
	if port == "" {
		port = "8090"
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@postgres:5432/tasks_db?sslmode=disable"
	}

	// Подключение к БД
	repo, err := repository.NewTaskRepository(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer repo.Close()

	taskService := service.NewTaskService(repo)

	// Определение типа Task
	taskType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Task",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.ID),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					task := p.Source.(*repository.Task)
					return task.ID, nil
				},
			},
			"title": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					task := p.Source.(*repository.Task)
					return task.Title, nil
				},
			},
			"description": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					task := p.Source.(*repository.Task)
					return task.Description, nil
				},
			},
			"dueDate": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					task := p.Source.(*repository.Task)
					return task.DueDate, nil
				},
			},
			"done": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					task := p.Source.(*repository.Task)
					return task.Done, nil
				},
			},
			"createdAt": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					task := p.Source.(*repository.Task)
					return task.CreatedAt.Format(time.RFC3339), nil
				},
			},
			"updatedAt": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					task := p.Source.(*repository.Task)
					return task.UpdatedAt.Format(time.RFC3339), nil
				},
			},
		},
	})

	// Input тип для создания задачи
	createTaskInputType := graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "CreateTaskInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"title": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"description": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
			"dueDate": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
		},
	})

	// Input тип для обновления задачи
	updateTaskInputType := graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "UpdateTaskInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"title": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
			"description": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
			"dueDate": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
			"done": &graphql.InputObjectFieldConfig{
				Type: graphql.Boolean,
			},
		},
	})

	// Query: получение списка задач
	tasksQuery := &graphql.Field{
		Type: graphql.NewList(graphql.NewNonNull(taskType)),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			ctx := context.Background()
			return taskService.GetAllTasks(ctx)
		},
	}

	// Query: получение задачи по ID
	taskQuery := &graphql.Field{
		Type: taskType,
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.ID),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			id, ok := p.Args["id"].(string)
			if !ok {
				return nil, nil
			}
			ctx := context.Background()
			return taskService.GetTaskByID(ctx, id)
		},
	}

	// Mutation: создание задачи
	createTaskMutation := &graphql.Field{
		Type: taskType,
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(createTaskInputType),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input, ok := p.Args["input"].(map[string]interface{})
			if !ok {
				return nil, nil
			}
			title := input["title"].(string)
			description := ""
			if desc, ok := input["description"]; ok && desc != nil {
				description = desc.(string)
			}
			dueDate := ""
			if due, ok := input["dueDate"]; ok && due != nil {
				dueDate = due.(string)
			}
			ctx := context.Background()
			return taskService.CreateTask(ctx, title, description, dueDate)
		},
	}

	// Mutation: обновление задачи
	updateTaskMutation := &graphql.Field{
		Type: taskType,
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.ID),
			},
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(updateTaskInputType),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			id := p.Args["id"].(string)
			input := p.Args["input"].(map[string]interface{})

			var title, description, dueDate *string
			var done *bool

			if t, ok := input["title"]; ok && t != nil {
				val := t.(string)
				title = &val
			}
			if d, ok := input["description"]; ok && d != nil {
				val := d.(string)
				description = &val
			}
			if dd, ok := input["dueDate"]; ok && dd != nil {
				val := dd.(string)
				dueDate = &val
			}
			if d, ok := input["done"]; ok && d != nil {
				val := d.(bool)
				done = &val
			}

			ctx := context.Background()
			return taskService.UpdateTask(ctx, id, title, description, dueDate, done)
		},
	}

	// Mutation: удаление задачи
	deleteTaskMutation := &graphql.Field{
		Type: graphql.Boolean,
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.ID),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			id := p.Args["id"].(string)
			ctx := context.Background()
			return taskService.DeleteTask(ctx, id)
		},
	}

	// Корневая схема Query
	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"tasks": tasksQuery,
			"task":  taskQuery,
		},
	})

	// Корневая схема Mutation
	rootMutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createTask": createTaskMutation,
			"updateTask": updateTaskMutation,
			"deleteTask": deleteTaskMutation,
		},
	})

	// Схема GraphQL
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	})
	if err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	// Handler для GraphQL
	graphQLHandler := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	// HTTP сервер
	http.Handle("/graphql", graphQLHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
            <html>
                <head><title>GraphQL</title></head>
                <body>
                    <h1>GraphQL Server</h1>
                    <p>Use <a href="/graphql">/graphql</a> endpoint</p>
                </body>
            </html>
        `))
	})

	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		log.Printf("GraphQL server starting on port %s", port)
		log.Printf("GraphiQL available at http://localhost:%s/graphql", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down GraphQL server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("GraphQL server stopped")
}
