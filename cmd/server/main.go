package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/wthareutalking/subscription-service/internal/config"
	"github.com/wthareutalking/subscription-service/internal/database"
	"github.com/wthareutalking/subscription-service/internal/handler"
	"github.com/wthareutalking/subscription-service/internal/logger"
	"github.com/wthareutalking/subscription-service/internal/repository"
	"github.com/wthareutalking/subscription-service/internal/service"
	"go.uber.org/zap"

	_ "github.com/wthareutalking/subscription-service/docs"
)

// @title           Subscription Service API
// @version         1.0
// @description     API для управления онлайн-подписками.
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	// Инициализация логгера
	logger.Init()
	defer logger.Sync()

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		logger.Log.Fatal("Cannot load config", zap.Error(err))
	}

	// Миграции базы данных
	if err := database.RunMigrations(cfg.Database.URL()); err != nil {
		logger.Log.Fatal("Failed to run migrations", zap.Error(err))
	}

	// Подключение к PostgreSQL
	db, err := sql.Open("postgres", cfg.Database.URL())
	if err != nil {
		logger.Log.Fatal("Failed to open database", zap.Error(err))
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Log.Fatal("Failed to ping database", zap.Error(err))
	}
	logger.Log.Info("Connected to database successfully")

	// Инициализация слоёв
	repo := repository.NewSubscriptionRepo(db)
	svc := service.NewSubscriptionService(repo)
	h := handler.NewSubscriptionHandler(svc)

	// Настройка роутера с middleware логирования
	r := chi.NewRouter()

	// Middleware для логирования всех HTTP-запросов
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.Log.Info("HTTP request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Duration("duration", time.Since(start)),
			)
		})
	})

	// Health-check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	//Swagger
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/subscriptions", h.Create)
		r.Get("/subscriptions", h.List)
		r.Get("/subscriptions/summary", h.Summary) // важно: до {id}
		r.Get("/subscriptions/{id}", h.GetByID)
		r.Put("/subscriptions/{id}", h.Update)
		r.Delete("/subscriptions/{id}", h.Delete)
	})

	//Запуск HTTP-сервера
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		logger.Log.Info("Starting server", zap.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Server failed", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Log.Info("Server exited")
}
