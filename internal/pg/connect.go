package pg

import (
	"context"
	"fmt"
	"time"

	"github.com/g0p43r/tui_psql/internal/domain"
	"github.com/g0p43r/tui_psql/internal/errs"
	"github.com/jackc/pgx/v5"
)

func Connect(profile domain.ConnectionProfile) (*pgx.Conn, error) {
	connString := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		profile.Host,
		profile.Port,
		profile.Database,
		profile.User,
		profile.Password,
		profile.SSLMode,
	)

	cfg, err := pgx.ParseConfig(connString)
	if err != nil {
		return nil, errs.E(errs.CodeConnect, "pg.Connect.ParseConfig", "Invalid connection configuration.", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return nil, errs.E(errs.CodeConnect, "pg.Connect.ConnectConfig", "Connection failed.", err)
	}

	if err := conn.Ping(ctx); err != nil {
		_ = conn.Close(context.Background())
		return nil, errs.E(errs.CodeConnect, "pg.Connect.Ping", "Connected but ping failed.", err)
	}

	return conn, nil
}
