package main

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/denismitr/shardstore/internal/common/logger"
	"github.com/denismitr/shardstore/internal/filegateway/config"
	"github.com/denismitr/shardstore/internal/filegateway/httpserver"
	"github.com/denismitr/shardstore/internal/filegateway/metastore"
	"github.com/denismitr/shardstore/internal/filegateway/processor"
	"github.com/denismitr/shardstore/internal/filegateway/remotestore"
	"github.com/denismitr/shardstore/internal/filegateway/shardmanager"
	"log"
	"os"
)

func main() {
	cfg := &config.Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("failed to retrieve env variables, %v", err)
	}

	lg := logger.NewStdoutLogger(logger.Env(cfg.AppEnv), cfg.AppName)

	shardManager, err := shardmanager.NewShardManager(cfg, lg)
	if err != nil {
		lg.Error(err)
		os.Exit(1)
	}

	grpcRemoteStore, err := remotestore.NewGRPCStore(cfg, lg)
	if err != nil {
		lg.Error(err)
		os.Exit(1)
	}

	metaStore, err := metastore.NewTmpMetaStore(cfg.AppName, lg)
	if err != nil {
		lg.Error(err)
		os.Exit(1)
	}

	p := processor.NewProcessor(cfg, shardManager, grpcRemoteStore, metaStore)

	server := httpserver.NewServer(cfg, p)
	if err := server.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
