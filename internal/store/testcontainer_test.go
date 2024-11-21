package beat

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"testing"
	"time"

	libpostgres "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/postgres"
	"github.com/docker/go-connections/nat"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	postgresContainer testcontainers.Container
	tdb               *libpostgres.Postgres
	postgresHost      string
	postgresPort      nat.Port
	ctx               = context.Background()
	dbName            = "drop-audiostreaming"
	dbUser            = "postgres"
	dbPassword        = "postgres"
)

func setupContainer() error {
	var err error
	postgresContainer, err = postgres.Run(ctx,
		"postgres:16.4-alpine",
		postgres.WithInitScripts(filepath.Join("testdata", "setup.sql")),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		return err
	}

	postgresHost, err = postgresContainer.Host(ctx)
	if err != nil {
		return err
	}

	postgresPort, err = postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		return err
	}

	return nil
}

func setupDBConnection() error {
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPassword, postgresHost, postgresPort.Port(), dbName)

	var err error
	tdb, err = libpostgres.New(ctx, connectionString)
	if err != nil {
		return err
	}

	return nil
}

func TestMain(m *testing.M) {
	if err := setupContainer(); err != nil {
		log.Fatal(err.Error())
	}

	if err := setupDBConnection(); err != nil {
		log.Fatal(err.Error())
	}

	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			log.Fatal(err.Error())
		}
	}()

	m.Run()
}
