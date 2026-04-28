package pg

import (
	"context"
	"fmt"
	"time"

	"github.com/g0p43r/tui_psql/internal/domain"
	"github.com/jackc/pgx/v5"
)

func ListTables(conn *pgx.Conn) ([]domain.DBObject, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := conn.Query(ctx, `
		select table_schema, table_name, table_type
		from information_schema.tables
		where table_schema not in ('pg_catalog', 'information_schema')
		order by table_schema, table_name
	`)
	if err != nil {
		return nil, fmt.Errorf("load tables: %w", err)
	}
	defer rows.Close()

	var tables []domain.DBObject
	for rows.Next() {
		var schema string
		var name string
		var tableType string

		if err := rows.Scan(&schema, &name, &tableType); err != nil {
			return nil, fmt.Errorf("scan table row: %w", err)
		}

		objectType := domain.ObjectTable
		if tableType == "VIEW" {
			objectType = domain.ObjectView
		}

		tables = append(tables, domain.DBObject{
			Schema: schema,
			Name:   name,
			Type:   objectType,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tables: %w", err)
	}

	return tables, nil
}
