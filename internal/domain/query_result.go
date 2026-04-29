package domain

type QueryResult struct {
	Columns      []string
	ColumnTypes  []string
	Rows         [][]string
	CommandTag   string
	RowsAffected int64
	DurationMs   int64
}
