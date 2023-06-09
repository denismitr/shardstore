package remotestore

import (
	"context"
	"fmt"
	"github.com/denismitr/shardstore/internal/common/closer"
	"github.com/denismitr/shardstore/internal/filegateway/config"
	"github.com/denismitr/shardstore/internal/filegateway/multishard"
	storeserverv1 "github.com/denismitr/shardstore/pkg/storeserver/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func bootstrapClients(cfg *config.Config) (map[multishard.ServerIdx]storeserverv1.FileServiceClient, error) {
	result := make(map[multishard.ServerIdx]storeserverv1.FileServiceClient, len(cfg.StorageServers))
	for idx, remoteServer := range cfg.StorageServers {
		conn, err := connect(cfg, remoteServer)
		if err != nil {
			return nil, err
		}
		result[multishard.ServerIdx(idx)] = storeserverv1.NewFileServiceClient(conn)
		closer.Add(func() error {
			return conn.Close()
		})
	}
	return result, nil
}

func connect(cfg *config.Config, remoteServer string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.StorageServerTimeout)
	defer cancel()

	credentials := grpc.WithTransportCredentials(insecure.NewCredentials())

	conn, err := grpc.DialContext(ctx, remoteServer, credentials, grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to remote storag server %s: %w", remoteServer, err)
	}

	return conn, nil
}
