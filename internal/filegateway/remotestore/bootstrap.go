package remotestore

import (
	"context"
	"fmt"
	"github.com/denismitr/shardstore/internal/filegateway/config"
	"github.com/denismitr/shardstore/internal/filegateway/multishard"
	storeserverv1 "github.com/denismitr/shardstore/pkg/storeserver/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func bootstrapClients(cfg *config.Config) (map[multishard.ServerIDX]storeserverv1.UploadServiceClient, error) {
	result := make(map[multishard.ServerIDX]storeserverv1.UploadServiceClient, len(cfg.StorageServers))
	for idx, remoteServer := range cfg.StorageServers {
		conn, err := connect(cfg, remoteServer)
		if err != nil {
			return nil, err
		}
		result[multishard.ServerIDX(idx)] = storeserverv1.NewUploadServiceClient(conn)
	}
	return result, nil
}

func connect(cfg *config.Config, remoteServer string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.StorageServerTimeout)
	defer cancel()

	credentials := grpc.WithTransportCredentials(insecure.NewCredentials())

	conn, err := grpc.DialContext(ctx, remoteServer, credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to remote storag server %s: %w", remoteServer, err)
	}

	return conn, nil
}
