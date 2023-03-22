package main

import (
	"github.com/caarlos0/env"
	"github.com/denismitr/shardstore/internal/filestore/config"
	"github.com/denismitr/shardstore/internal/filestore/uploadserver"
	"log"
)

func main() {
	cfg := &config.Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("failed to retrieve env variables, %v", err)
	}

	//todo: use zero log
	uploadSrv := uploadserver.NewServer(cfg)
	if err := uploadserver.StartGRPCServer(cfg, uploadSrv); err != nil {
		log.Fatal(err)
	}
}
