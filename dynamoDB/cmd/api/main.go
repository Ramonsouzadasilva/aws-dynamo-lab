package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	appConfig "github.com/ramon/goals-tasks-api/internal/config"
	"github.com/ramon/goals-tasks-api/internal/infra"
	"github.com/ramon/goals-tasks-api/internal/middleware"
	
	// Modules Handlers & Repositories
	dashboardHandler "github.com/ramon/goals-tasks-api/internal/modules/dashboard/handler"
	dashboardUseCase "github.com/ramon/goals-tasks-api/internal/modules/dashboard/usecase"
	
	goalHandler "github.com/ramon/goals-tasks-api/internal/modules/goal/handler"
	goalRepo "github.com/ramon/goals-tasks-api/internal/modules/goal/repository"
	goalUseCase "github.com/ramon/goals-tasks-api/internal/modules/goal/usecase"
	
	taskHandler "github.com/ramon/goals-tasks-api/internal/modules/task/handler"
	taskRepo "github.com/ramon/goals-tasks-api/internal/modules/task/repository"
	taskUseCase "github.com/ramon/goals-tasks-api/internal/modules/task/usecase"
	
	userHandler "github.com/ramon/goals-tasks-api/internal/modules/user/handler"
	userRepo "github.com/ramon/goals-tasks-api/internal/modules/user/repository"
	userUseCase "github.com/ramon/goals-tasks-api/internal/modules/user/usecase"
)

func main() {
	// Initialize JSON Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.InfoContext(ctx, "Iniciando Goals and Tasks API...")

	// Load Configuration
	cfg := appConfig.Load()

	// Initialize OpenTelemetry
	telemetry, err := infra.InitTelemetry(ctx, "goals-tasks-api")
	if err != nil {
		slog.ErrorContext(ctx, "Falha ao inicializar OpenTelemetry", "error", err)
	} else {
		defer telemetry.Shutdown(context.Background())
	}

	// Initialize AWS Client
	awsClient, err := infra.NewAWSClient(ctx, cfg)
	if err != nil {
		slog.ErrorContext(ctx, "Falha ao conectar com AWS (DynamoDB)", "error", err)
		os.Exit(1)
	}

	// Initialize Repositories
	userRepository := userRepo.NewDynamoUserRepository(awsClient.DynamoDB, cfg.DynamoDBTable)
	taskRepository := taskRepo.NewDynamoTaskRepository(awsClient.DynamoDB, cfg.DynamoDBTable)
	goalRepository := goalRepo.NewDynamoGoalRepository(awsClient.DynamoDB, cfg.DynamoDBTable)

	// Initialize UseCases
	authUC := userUseCase.NewAuthUseCase(userRepository, cfg.JWTSecret)
	taskUC := taskUseCase.NewTaskUseCase(taskRepository, goalRepository)
	// goalUseCase needs taskRepository to count tasks dynamically for goal progress
	goalUC := goalUseCase.NewGoalUseCase(goalRepository, taskRepository)
	dashboardUC := dashboardUseCase.NewDashboardUseCase(goalRepository, taskRepository)

	// Initialize Handlers
	authH := userHandler.NewAuthHandler(authUC)
	goalH := goalHandler.NewGoalHandler(goalUC)
	taskH := taskHandler.NewTaskHandler(taskUC)
	dashboardH := dashboardHandler.NewDashboardHandler(dashboardUC)
	healthChecker := infra.NewHealthChecker(awsClient.DynamoDB, cfg.DynamoDBTable)

	// Router setup
	r := chi.NewRouter()

	// Global Middlewares
	r.Use(middleware.CorrelationID)
	r.Use(middleware.Logging)
	r.Use(middleware.Recovery)
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.RateLimit(10.0, 20)) // 10 RPS, Burst 20

	// Health check routes
	r.Get("/health", healthChecker.HealthHandler)
	r.Get("/ready", healthChecker.ReadyHandler)

	// Authentication routes (Public)
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authH.Register)
		r.Post("/login", authH.Login)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(cfg.JWTSecret))

		// Goals CRUD
		r.Route("/goals", func(r chi.Router) {
			r.Post("/", goalH.Create)
			r.Get("/", goalH.List)
			r.Get("/{id}", goalH.Get)
			r.Put("/{id}", goalH.Update)
			r.Delete("/{id}", goalH.Delete)
		})

		// Tasks CRUD
		r.Route("/tasks", func(r chi.Router) {
			r.Post("/", taskH.Create)
			r.Get("/", taskH.List)
			r.Get("/{id}", taskH.Get)
			r.Put("/{id}", taskH.Update)
			r.Delete("/{id}", taskH.Delete)
		})

		// Special Analytical Endpoints
		r.Get("/dashboard", dashboardH.Get)
		r.Get("/weekly-tasks", taskH.ListWeekly)
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	serverError := make(chan error, 1)
	go func() {
		slog.InfoContext(ctx, "Servidor HTTP escutando", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverError <- err
		}
	}()

	// Graceful Shutdown
	select {
	case err := <-serverError:
		slog.ErrorContext(ctx, "Erro crítico no servidor HTTP", "error", err)
	case <-ctx.Done():
		slog.InfoContext(ctx, "Sinal de término recebido. Iniciando desligamento gracioso...")
		
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.ErrorContext(ctx, "Falha ao desligar servidor HTTP", "error", err)
		} else {
			slog.InfoContext(ctx, "Servidor HTTP finalizado com sucesso")
		}
	}
}
