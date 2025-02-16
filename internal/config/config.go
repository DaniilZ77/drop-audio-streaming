package config

import (
	"flag"
	"time"

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
		GRPCPort           string
		GRPCClientRetries  uint
		GRPCClientTimeout  time.Duration
		GRPCUserClientAddr string
		HTTPPort           string
		ReadTimeout        int
		VerificationSecret string
		URLTTL             int
		JWTSecret          string
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
		FileSizeLimit    int64
		ArchiveSizeLimit int64
		ImageSizeLimit   int64
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
	verificationSecret := flag.String("verification_secret", "secret", "secret to verify url data")
	urlTTL := flag.Int("url_ttl", 10, "url ttl in minutes")
	jwtSecret := flag.String("jwt_secret", "secret", "jwt secret")

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

	// grpc client
	grpcClientRetries := flag.Uint("grpc_client_retries", 1, "grpc client retries")
	grpcClientTimeout := flag.Duration("grpc_client_timeout", 2*time.Second, "grpc client timeout")
	grpcUserClientAddr := flag.String("grpc_user_client_addr", "", "grpc user client addr")

	// media limits
	fileSizeLimit := flag.Int64("file_size_limit", 2<<24, "file size limit in bytes")
	archiveSizeLimit := flag.Int64("archive_size_limit", 2<<31, "archive size limit in bytes")
	imageSizeLimit := flag.Int64("image_size_limit", 2<<24, "image size limit in bytes")

	flag.Parse()

	cfg := &Config{
		HTTP: HTTP{
			GRPCPort:           *gRPCPort,
			GRPCClientRetries:  *grpcClientRetries,
			GRPCClientTimeout:  *grpcClientTimeout,
			GRPCUserClientAddr: *grpcUserClientAddr,
			HTTPPort:           *httpPort,
			ReadTimeout:        *readTimeout,
			VerificationSecret: *verificationSecret,
			URLTTL:             *urlTTL,
			JWTSecret:          *jwtSecret,
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
			FileSizeLimit:    *fileSizeLimit,
			ArchiveSizeLimit: *archiveSizeLimit,
			ImageSizeLimit:   *imageSizeLimit,
		},
	}

	return cfg, nil
}
