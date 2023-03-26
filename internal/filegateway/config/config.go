package config

import "time"

type Config struct {
	AppName              string        `env:"FG_APP_NAME" envDefault:"filegateway"`
	AppEnv               string        `env:"FG_APP_ENV"  envDefault:"local"`
	HTTPPort             uint          `env:"FG_HTTP_PORT" envDefault:"8080"`
	MaxFileSize          int64         `env:"FG_MAX_FILE_SIZE" envDefault:"10485760"` // 10Mb
	NumberOfChunks       int64         `env:"FG_NUMBER_OF_CHUNKS" envDefault:"3"`
	StorageServers       []string      `env:"FG_STORAGE_SERVERS" envSeparator:";" envDefault:"localhost:9000;localhost:9001;localhost:9002"`
	StorageServerTimeout time.Duration `env:"FG_STORAGE_SERVER_TIMEOUT" envDefault:"10s"`
}
