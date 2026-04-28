package pg

import (
	"context"
	"fmt"
	"time"

	"github.com/g0p43r/tui_psql/internal/domain"
	"github.com/g0p43r/tui_psql/internal/pg/formatter"
	"github.com/jackc/pgx/v5"
)

func PreviewTable(conn *pgx.Conn, table domain.DBObject, limit int) (domain.QueryResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := fmt.Sprintf(
		"select * from %s limit %d",
		pgx.Identifier{table.Schema, table.Name}.Sanitize(),
		limit,
	)

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return domain.QueryResult{}, fmt.Errorf("load table preview: %w", err)
	}
	defer rows.Close()

	fields := rows.FieldDescriptions()
	columns := make([]string, 0, len(fields))
	for _, field := range fields {
		columns = append(columns, field.Name)
	}

	result := domain.QueryResult{
		Columns: columns,
	}

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return domain.QueryResult{}, fmt.Errorf("read row values: %w", err)
		}

		line := make([]string, 0, len(values))
		for i, value := range values {
			line = append(line, formatter.Value(fields[i].DataTypeOID, value))
		}

		result.Rows = append(result.Rows, line)
	}

	if err := rows.Err(); err != nil {
		return domain.QueryResult{}, fmt.Errorf("iterate preview rows: %w", err)
	}

	result.RowsAffected = int64(len(result.Rows))
	return result, nil
}
