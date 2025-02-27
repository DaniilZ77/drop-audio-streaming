package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env                string     `yaml:"env" env-default:"local"`
	DatabaseURL        string     `yaml:"database_url" env:"DATABASE_URL" env-required:"true"`
	JwtSecret          string     `yaml:"jwt_secret" env-required:"true"`
	Tls                Tls        `yaml:"tls"`
	GrpcPort           string     `yaml:"grpc_port" env-required:"true"`
	HttpPort           string     `yaml:"http_port" env-required:"true"`
	VerificationSecret string     `yaml:"verification_secret" env-required:"true"`
	UrlTtl             int        `yaml:"url_ttl" env-required:"true"`
	FileSizeLimit      int64      `yaml:"file_size_limit" env-required:"true"`
	ArchiveSizeLimit   int64      `yaml:"archive_size_limit" env-required:"true"`
	ImageSizeLimit     int64      `yaml:"image_size_limit" env-required:"true"`
	Minio              Minio      `yaml:"minio" env-required:"true"`
	GrpcClient         GrpcClient `yaml:"grpc_client" env-required:"true"`
}

type Tls struct {
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

type Minio struct {
	Password string `yaml:"password" env:"MINIO_PASSWORD" env-required:"true"`
	User     string `yaml:"user" env:"MINIO_USER" env-required:"true"`
	Url      string `yaml:"url" env:"MINIO_URL" env-required:"true"`
	Bucket   string `yaml:"bucket" env:"MINIO_BUCKET" env-required:"true"`
	UseSsl   bool   `yaml:"use_ssl" env-default:"false"`
	Location string `yaml:"location" env-required:"true"`
}

type GrpcClient struct {
	Retries uint          `yaml:"retries" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env-required:"true"`
	Port    string        `yaml:"port" env-required:"true"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	return MustLoadPath(configPath)
}

func MustLoadPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
