package pg

import (
	"context"
	"time"

	"github.com/g0p43r/tui_psql/internal/domain"
	"github.com/g0p43r/tui_psql/internal/errs"
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
		return nil, errs.E(errs.CodeQuery, "pg.ListTables.Query", "Failed to load tables.", err)
	}
	defer rows.Close()

	var tables []domain.DBObject
	for rows.Next() {
		var schema string
		var name string
		var tableType string

		if err := rows.Scan(&schema, &name, &tableType); err != nil {
			return nil, errs.E(errs.CodeQuery, "pg.ListTables.Scan", "Failed to read table metadata.", err)
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
		return nil, errs.E(errs.CodeQuery, "pg.ListTables.Iterate", "Failed to iterate tables.", err)
	}

	return tables, nil
}
