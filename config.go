package app

import "github.com/katalabut/fast-app/logger"

type Config struct {
	Logger logger.Config

	AutoMaxProcs struct {
		Enabled bool
		Min     int
	}

	DebugServer struct {
		Enabled bool
		Port    int
	}
}
