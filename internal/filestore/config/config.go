package config

type Config struct {
	AppName       string `env:"FS_APP_NAME" envDefault:"filestore"`
	ID            uint   `env:"FS_ID"`
	AppEnv        string `env:"FS_APP_ENV"  envDefault:"local"`
	GRPCPort      uint   `env:"FS_GRPC_PORT" envDefault:"9000"`
	ReflectionAPI bool   `env:"FS_REFLECTION_API" envDefault:"true"`
}
