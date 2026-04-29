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
	"github.com/jackc/pgx/v5/pgxpool"
)

const QueryRowLimit = 500

func PreviewTable(pool *pgxpool.Pool, table domain.DBObject, limit int) (domain.QueryResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := fmt.Sprintf(
		"select * from %s limit %d",
		pgx.Identifier{table.Schema, table.Name}.Sanitize(),
		limit,
	)

	rows, err := pool.Query(ctx, query)
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
	result.Limit = limit
	return result, nil
}

func ExecuteSQL(pool *pgxpool.Pool, query string, queryType domain.SQLQueryType) (domain.QueryResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return domain.QueryResult{}, errs.Validation("pg.ExecuteSQL.Validate", "SQL is empty.")
	}

	resolved := resolveQueryType(query, queryType)
	start := time.Now()

	if resolved.IsRead() {
		return queryRows(pool, query, QueryRowLimit, start)
	}

	return execStatement(pool, query, start)
}

func resolveQueryType(query string, requested domain.SQLQueryType) domain.SQLQueryType {
	if requested != domain.QueryTypeAuto {
		return requested
	}

	upper := strings.ToUpper(stripLeadingSQLComments(query))
	switch {
	case strings.HasPrefix(upper, "SELECT"), strings.HasPrefix(upper, "WITH"), strings.HasPrefix(upper, "SHOW"):
		return domain.QueryTypeSelect
	default:
		return domain.QueryTypeExec
	}
}

func stripLeadingSQLComments(query string) string {
	for {
		query = strings.TrimSpace(query)
		if strings.HasPrefix(query, "--") {
			if idx := strings.IndexByte(query, '\n'); idx >= 0 {
				query = query[idx+1:]
				continue
			}
			return ""
		}
		if strings.HasPrefix(query, "/*") {
			if idx := strings.Index(query, "*/"); idx >= 0 {
				query = query[idx+2:]
				continue
			}
			return ""
		}
		return query
	}
}

func queryRows(pool *pgxpool.Pool, query string, limit int, start time.Time) (domain.QueryResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	rows, err := pool.Query(ctx, query)
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
		Limit:       limit,
	}

	for rows.Next() {
		if limit > 0 && len(result.Rows) >= limit {
			result.Truncated = true
			break
		}

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
	if result.RowsAffected < 0 || result.Truncated {
		result.RowsAffected = int64(len(result.Rows))
	}
	result.DurationMs = time.Since(start).Milliseconds()
	return result, nil
}

func execStatement(pool *pgxpool.Pool, query string, start time.Time) (domain.QueryResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	tag, err := pool.Exec(ctx, query)
	if err != nil {
		return domain.QueryResult{}, errs.E(errs.CodeQuery, "pg.ExecuteSQL.Exec", "SQL execution failed.", err)
	}

	return domain.QueryResult{
		CommandTag:   tag.String(),
		RowsAffected: tag.RowsAffected(),
		DurationMs:   time.Since(start).Milliseconds(),
	}, nil
}

func CurrentSchema(pool *pgxpool.Pool) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var schema string
	if err := pool.QueryRow(ctx, "select current_schema()").Scan(&schema); err != nil {
		return "", errs.E(errs.CodeQuery, "pg.CurrentSchema.QueryRow", "Failed to resolve current schema.", err)
	}
	return strings.TrimSpace(schema), nil
}
