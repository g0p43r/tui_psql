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
		select
			n.nspname as table_schema,
			c.relname as table_name,
			case
				when c.relkind in ('v', 'm') then 'VIEW'
				else 'BASE TABLE'
			end as table_type
		from pg_catalog.pg_class c
		join pg_catalog.pg_namespace n on n.oid = c.relnamespace
		where c.relkind in ('r', 'p', 'v', 'm', 'f')
		  and n.nspname not in ('pg_catalog', 'information_schema')
		  and n.nspname not like 'pg_toast%'
		order by n.nspname, c.relname
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
