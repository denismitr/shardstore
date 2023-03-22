package config

import "time"

type Config struct {
	AppName              string        `env:"FG_APP_NAME" envDefault:"filegateway"`
	AppEnv               string        `env:"FG_APP_ENV"  envDefault:"local"`
	LogLevel             string        `env:"FG_LOG_LEVEL" envDefault:"info"`
	HTTPPort             uint          `env:"FG_HTTP_PORT" envDefault:"8080"`
	MaxFileSize          int64         `env:"FG_MAX_FILE_SIZE" envDefault:"10485760"` // 10Mb
	NumberOfChunks       int64         `env:"FG_NUMBER_OF_CHUNKS" envDefault:"5"`
	StorageServers       []string      `env:"STORAGE_SERVERS" envSeparator:":"`
	StorageServerTimeout time.Duration `env:"STORAGE_SERVER_TIMEOUT" envDefault:"3s"`
}
