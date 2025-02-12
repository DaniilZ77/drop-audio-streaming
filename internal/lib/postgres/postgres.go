package postgres

import (
	"context"
	"time"

	"github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_defaultMaxPoolSize  = 10
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)

type Postgres struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration

	DB *pgxpool.Pool
}

func New(ctx context.Context, dbURL string, opts ...Option) (*Postgres, error) {
	pg := &Postgres{
		maxPoolSize:  _defaultMaxPoolSize,
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
	}

	// Custom options
	for _, opt := range opts {
		opt(pg)
	}

	var db *pgxpool.Pool
	var err error

	for pg.connAttempts > 0 {
		db, err = pgxpool.New(ctx, dbURL)
		if err == nil && db.Ping(ctx) == nil {
			db.Config().MaxConns = int32(pg.maxPoolSize)
			db.Config().MaxConnLifetime = time.Hour

			pg.DB = db
			break
		}

		logger.Log().Debug(ctx,
			"postgres is trying to connect, attempts left: %d", pg.connAttempts,
		)

		time.Sleep(pg.connTimeout)

		pg.connAttempts--
	}

	if err != nil {
		logger.Log().Fatal(ctx, "failed to connect to database: %s", err.Error())
		return nil, err
	}

	return pg, nil
}

func (p *Postgres) Close(ctx context.Context) {
	p.DB.Close()
}
