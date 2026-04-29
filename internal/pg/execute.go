package pg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/g0p43r/tui_psql/internal/domain"
	"github.com/g0p43r/tui_psql/internal/errs"
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
		return domain.QueryResult{}, errs.E(errs.CodeQuery, "pg.PreviewTable.Query", "Failed to load table preview.", err)
	}
	defer rows.Close()

	fields := rows.FieldDescriptions()
	columns := make([]string, 0, len(fields))
	columnTypes := make([]string, 0, len(fields))
	for _, field := range fields {
		columns = append(columns, field.Name)
		columnTypes = append(columnTypes, formatter.TypeName(field.DataTypeOID))
	}

	result := domain.QueryResult{
		Columns:     columns,
		ColumnTypes: columnTypes,
	}

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return domain.QueryResult{}, errs.E(errs.CodeQuery, "pg.PreviewTable.Values", "Failed to read row values.", err)
		}

		line := make([]string, 0, len(values))
		for i, value := range values {
			line = append(line, formatter.Value(fields[i].DataTypeOID, value))
		}

		result.Rows = append(result.Rows, line)
	}

	if err := rows.Err(); err != nil {
		return domain.QueryResult{}, errs.E(errs.CodeQuery, "pg.PreviewTable.Iterate", "Failed to iterate preview rows.", err)
	}

	result.RowsAffected = int64(len(result.Rows))
	return result, nil
}

func ExecuteSQL(conn *pgx.Conn, query string, queryType domain.SQLQueryType) (domain.QueryResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return domain.QueryResult{}, errs.Validation("pg.ExecuteSQL.Validate", "SQL is empty.")
	}

	resolved := resolveQueryType(query, queryType)
	start := time.Now()

	if resolved.IsRead() {
		return queryRows(conn, query, start)
	}

	return execStatement(conn, query, start)
}

func resolveQueryType(query string, requested domain.SQLQueryType) domain.SQLQueryType {
	if requested != domain.QueryTypeAuto {
		return requested
	}

	upper := strings.ToUpper(strings.TrimSpace(query))
	switch {
	case strings.HasPrefix(upper, "SELECT"), strings.HasPrefix(upper, "WITH"), strings.HasPrefix(upper, "SHOW"):
		return domain.QueryTypeSelect
	default:
		return domain.QueryTypeExec
	}
}

func queryRows(conn *pgx.Conn, query string, start time.Time) (domain.QueryResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return domain.QueryResult{}, errs.E(errs.CodeQuery, "pg.ExecuteSQL.Query", "SQL query failed.", err)
	}
	defer rows.Close()

	fields := rows.FieldDescriptions()
	columns := make([]string, 0, len(fields))
	columnTypes := make([]string, 0, len(fields))
	for _, field := range fields {
		columns = append(columns, field.Name)
		columnTypes = append(columnTypes, formatter.TypeName(field.DataTypeOID))
	}

	result := domain.QueryResult{
		Columns:     columns,
		ColumnTypes: columnTypes,
	}

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return domain.QueryResult{}, errs.E(errs.CodeQuery, "pg.ExecuteSQL.Values", "Failed to read query row.", err)
		}

		line := make([]string, 0, len(values))
		for i, value := range values {
			line = append(line, formatter.Value(fields[i].DataTypeOID, value))
		}

		result.Rows = append(result.Rows, line)
	}

	if err := rows.Err(); err != nil {
		return domain.QueryResult{}, errs.E(errs.CodeQuery, "pg.ExecuteSQL.Iterate", "Failed to iterate query rows.", err)
	}

	result.CommandTag = rows.CommandTag().String()
	result.RowsAffected = rows.CommandTag().RowsAffected()
	result.DurationMs = time.Since(start).Milliseconds()
	return result, nil
}

func execStatement(conn *pgx.Conn, query string, start time.Time) (domain.QueryResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	tag, err := conn.Exec(ctx, query)
	if err != nil {
		return domain.QueryResult{}, errs.E(errs.CodeQuery, "pg.ExecuteSQL.Exec", "SQL execution failed.", err)
	}

	return domain.QueryResult{
		CommandTag:   tag.String(),
		RowsAffected: tag.RowsAffected(),
		DurationMs:   time.Since(start).Milliseconds(),
	}, nil
}
