package postgres

import (
	"context"
	"log/slog"
	"time"

	sl "github.com/MAXXXIMUS-tropical-milkshake/drop-audio-streaming/internal/lib/logger"
	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_defaultMaxPoolSize  = 10
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)

type Postgres struct {
	maxPoolSize  int32
	connAttempts int32
	connTimeout  time.Duration

	DB *pgxpool.Pool
}

func New(ctx context.Context, dbURL string, log *slog.Logger, opts ...Option) (*Postgres, error) {
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
		if err != nil {
			continue
		}

		if err = db.Ping(ctx); err == nil {
			db.Config().MaxConns = pg.maxPoolSize
			db.Config().MaxConnLifetime = time.Hour

			pg.DB = db
			break
		}

		log.Debug("postgres is trying to connect", slog.Any("attempts left", pg.connAttempts))

		time.Sleep(pg.connTimeout)

		pg.connAttempts--
	}

	if err != nil {
		log.Error("failed to connect to database", sl.Err(err))
		return nil, err
	}

	return pg, nil
}

func (p *Postgres) Close(ctx context.Context) {
	p.DB.Close()
}
