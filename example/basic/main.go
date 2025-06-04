package main

import (
	"context"
	"time"

	fastapp "github.com/katalabut/fast-app"
	"github.com/katalabut/fast-app/configloader"
	"github.com/katalabut/fast-app/logger"
	"github.com/katalabut/fast-app/service"
)

type Config struct {
	App         fastapp.Config
	DebugServer service.DebugServer
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
	cfg, err := configloader.New[Config]()
	if err != nil {
		logger.Fatal(context.Background(), "failed to load config:", err)
	}

	apiService := NewApiService()

	fastapp.New(cfg.App).
		Add(service.NewDefaultDebugService(cfg.DebugServer)).
		Add(apiService).
		Start()
}
