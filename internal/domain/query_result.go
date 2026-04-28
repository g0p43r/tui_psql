package domain

type QueryResult struct {
	Columns      []string
	Rows         [][]string
	CommandTag   string
	RowsAffected int64
	DurationMs   int64
}
