package main

import (
	"github.com/caarlos0/env"
	"github.com/denismitr/shardstore/internal/common/logger"
	"github.com/denismitr/shardstore/internal/filestore/config"
	"github.com/denismitr/shardstore/internal/filestore/uploadserver"
	"log"
	"os"
)

func main() {
	cfg := &config.Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("failed to retrieve env variables, %v", err)
	}

	lg := logger.NewStdoutLogger(logger.Env(cfg.AppEnv), cfg.AppName)

	uploadSrv := uploadserver.NewServer(cfg, lg)
	if err := uploadserver.StartGRPCServer(cfg, lg, uploadSrv); err != nil {
		lg.Error(err)
		os.Exit(1)
	}
}
