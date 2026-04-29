package pg

import (
	"context"
	"net"
	"net/url"
	"time"

	"github.com/g0p43r/tui_psql/internal/domain"
	"github.com/g0p43r/tui_psql/internal/errs"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(profile domain.ConnectionProfile) (*pgxpool.Pool, error) {
	connURL := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(profile.User, profile.Password),
		Host:   net.JoinHostPort(profile.Host, profile.Port),
		Path:   "/" + profile.Database,
	}
	query := connURL.Query()
	query.Set("sslmode", profile.SSLMode)
	connURL.RawQuery = query.Encode()

	cfg, err := pgxpool.ParseConfig(connURL.String())
	if err != nil {
		return nil, errs.E(errs.CodeConnect, "pg.Connect.ParseConfig", "Invalid connection configuration.", err)
	}
	cfg.MaxConns = 4
	cfg.MinConns = 0

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, errs.E(errs.CodeConnect, "pg.Connect.ConnectConfig", "Connection failed.", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, errs.E(errs.CodeConnect, "pg.Connect.Ping", "Connected but ping failed.", err)
	}

	return pool, nil
}
