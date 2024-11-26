package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/a-berahman/todo-list/config"
	"github.com/a-berahman/todo-list/internal/application"
	"github.com/a-berahman/todo-list/internal/handlers"
	"github.com/a-berahman/todo-list/internal/infra/db"
	"github.com/a-berahman/todo-list/internal/infra/queue"
	"github.com/a-berahman/todo-list/internal/infra/storage"
	"github.com/go-playground/validator"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
)

func main() {
	logger := slog.Default()

	conf, err := loadConfig()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	dbConn := initDB(conf.DBURL)
	defer dbConn.Close(context.Background())

	e := setupEcho(logger)

	todoService := application.NewTodoService(
		db.New(dbConn),
		storage.NewS3FileStorage(conf.AWSConf.S3Conf.Region, conf.AWSConf.S3Conf.Bucket, conf.AWSConf.Endpoint, conf.AWSConf.S3Conf.DisableSSL, conf.AWSConf.S3Conf.ForcePathStyle),
		queue.NewSQSPublisher(conf.AWSConf.SQSConf.Region, conf.AWSConf.SQSConf.QueueURL, conf.AWSConf.Endpoint, conf.AWSConf.SQSConf.DisableSSL),
		logger,
	)

	h := handlers.NewHandler(todoService, logger)
	e.POST("api/v1/upload", h.TodoHandler.CreateTodo)

	go func() {
		if err := e.Start(conf.Port); err != nil {
			logger.Info("shutting down the server", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logger.Error("failed to shutdown server gracefully", "error", err)
		os.Exit(1)
	}

	logger.Info("server shutdown successfully")
}

func setupEcho(logger *slog.Logger) *echo.Echo {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(requestTimer(logger))

	v := validator.New()
	v.RegisterValidation("datetime", func(fl validator.FieldLevel) bool {
		_, err := time.Parse(fl.Param(), fl.Field().String())
		return err == nil
	})
	e.Validator = &CustomValidator{Validator: v}

	return e
}

func initDB(dbURL string) *pgx.Conn {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := conn.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	return conn
}

func loadConfig() (*config.Config, error) {
	viper.SetConfigFile(filepath.Join(getProjectRoot(), ".env"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("cannot load config file: %v", err)
	}

	return config.NewConfig()
}
func getProjectRoot() string {
	_, b, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(b), "..")
	return projectRoot
}

func requestTimer(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			duration := time.Since(start)

			logger.Info("request completed",
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
				"status", c.Response().Status,
				"duration", duration.String(),
				"duration_ms", duration.Milliseconds(),
			)

			return err
		}
	}

}

type CustomValidator struct {
	Validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.Validator.Struct(i)
}
