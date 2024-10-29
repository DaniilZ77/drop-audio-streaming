package config

import (
	"flag"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
)

type (
	Config struct {
		HTTP
		Log
		DB
		TLS
		Audio
	}

	HTTP struct {
		GRPCPort    string
		HTTPPort    string
		ReadTimeout int
	}

	Log struct {
		Level string
	}

	DB struct {
		URL           string
		MinioPassword string
		MinioUser     string
		MinioEndpoint string
		MinioBucket   string
		MinioUseSSL   bool
		MinioLocation string
	}

	Audio struct {
		ChunkSize int
	}

	TLS struct {
		Cert string
		Key  string
	}
)

func NewConfig() (*Config, error) {
	gRPCPort := flag.String("grpc_port", "localhost:50010", "GRPC Port")
	logLevel := flag.String("log_level", string(logger.InfoLevel), "logger level")
	dbURL := flag.String("db_url", "", "url for connection to database")
	httpPort := flag.String("http_port", "localhost:8080", "HTTP Port")
	readTimeout := flag.Int("read_timeout", 5, "read timeout")

	// TLS
	cert := flag.String("cert", "", "path to cert file")
	key := flag.String("key", "", "path to key file")

	// Minio S3 storage
	minioPassword := flag.String("minio_password", "minioadmin", "minio password")
	minioUser := flag.String("minio_user", "minioadmin", "minio user")
	minioEndpoint := flag.String("minio_endpoint", "192.168.0.170:9000", "minio endpoint")
	minioBucket := flag.String("minio_bucket", "drop-audio", "minio bucket")
	minioUseSSL := flag.Bool("minio_use_ssl", false, "minio use ssl")
	minioLocation := flag.String("minio_location", "us-east-1", "minio location")

	// audio
	chunkSize := flag.Int("chunk_size", 1024, "chunk size")

	flag.Parse()

	cfg := &Config{
		HTTP: HTTP{
			GRPCPort:    *gRPCPort,
			HTTPPort:    *httpPort,
			ReadTimeout: *readTimeout,
		},
		Log: Log{
			Level: *logLevel,
		},
		DB: DB{
			URL:           *dbURL,
			MinioPassword: *minioPassword,
			MinioUser:     *minioUser,
			MinioEndpoint: *minioEndpoint,
			MinioBucket:   *minioBucket,
			MinioUseSSL:   *minioUseSSL,
			MinioLocation: *minioLocation,
		},
		TLS: TLS{
			Cert: *cert,
			Key:  *key,
		},
		Audio: Audio{
			ChunkSize: *chunkSize,
		},
	}

	return cfg, nil
}
