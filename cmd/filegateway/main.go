package main

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/denismitr/shardstore/internal/filegateway/config"
	"github.com/denismitr/shardstore/internal/filegateway/httpserver"
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

	//todo: use zero log

	shardManager := shardmanager.NewShardManager(cfg)
	grpcStore, err := remotestore.NewGRPCStore(cfg)
	if err != nil {
		fmt.Println(err) // todo: logger
		os.Exit(1)
	}

	p := processor.NewProcessor(cfg, shardManager, grpcStore)

	server := httpserver.NewServer(cfg, p)
	if err := server.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
