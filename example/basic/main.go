package main

import (
	"context"
	"time"

	app "github.com/katalabut/fast-app"
	"github.com/katalabut/fast-app/config"
	"github.com/katalabut/fast-app/logger"
)

type Config struct {
	App         app.Config
	DebugServer app.DebugServer
}

type ApiService struct {
}

func NewApiService() *ApiService {
	return &ApiService{}
}

func (s *ApiService) Run(ctx context.Context) error {
	logger.Info(ctx, "ApiService is running")
	time.Sleep(5 * time.Second)
	return nil
}

func (s *ApiService) Shutdown(ctx context.Context) error {
	logger.Info(ctx, "ApiService is shutting down")
	return nil
}

func main() {
	cfg, err := config.New[Config]()
	if err != nil {
		logger.Fatal(context.Background(), "failed to load config:", err)
	}

	apiService := NewApiService()

	app.New(cfg.App).
		Add(app.NewDefaultDebugService(cfg.DebugServer)).
		Add(apiService).
		Start()
}
