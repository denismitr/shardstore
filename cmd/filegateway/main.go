package main

import (
	"github.com/caarlos0/env"
	"github.com/denismitr/shardstore/internal/common/closer"
	"github.com/denismitr/shardstore/internal/common/logger"
	"github.com/denismitr/shardstore/internal/filegateway/config"
	"github.com/denismitr/shardstore/internal/filegateway/downloader"
	"github.com/denismitr/shardstore/internal/filegateway/httpserver"
	"github.com/denismitr/shardstore/internal/filegateway/metastore"
	"github.com/denismitr/shardstore/internal/filegateway/remotestore"
	"github.com/denismitr/shardstore/internal/filegateway/shardmanager"
	"github.com/denismitr/shardstore/internal/filegateway/uploader"
	"log"
	"os"
)

func main() {
	cfg := &config.Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("failed to retrieve env variables, %v", err)
	}

	lg := logger.NewStdoutLogger(logger.Env(cfg.AppEnv), cfg.AppName)

	defer closer.CloseAll()

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

	fileUploader := uploader.NewUploader(cfg, shardManager, grpcRemoteStore, metaStore, lg)
	fileDownloader := downloader.NewDownloader(cfg, grpcRemoteStore, metaStore, lg)

	server := httpserver.NewServer(cfg, lg, fileUploader, fileDownloader)
	if err := server.Start(); err != nil {
		lg.Error(err)
		os.Exit(1)
	}
}
