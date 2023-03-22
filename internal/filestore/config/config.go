package config

type Config struct {
	AppName       string `env:"STORE_APP_NAME" envDefault:"filestore"`
	ID            uint   `env:"STORE_ID"`
	AppEnv        string `env:"STORE_APP_ENV"  envDefault:"local"`
	LogLevel      string `env:"STORE_LOG_LEVEL" envDefault:"info"`
	GRPCPort      uint   `env:"STORE_GRPC_PORT" envDefault:"9000"`
	ReflectionAPI bool   `env:"REFLECTION_API"                 envDefault:"true"`
}
