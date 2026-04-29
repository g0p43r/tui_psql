package domain

type SQLQueryType string

const (
	QueryTypeAuto   SQLQueryType = "auto"
	QueryTypeSelect SQLQueryType = "select"
	QueryTypeInsert SQLQueryType = "insert"
	QueryTypeUpdate SQLQueryType = "update"
	QueryTypeDelete SQLQueryType = "delete"
	QueryTypeExec   SQLQueryType = "exec"
)

func (t SQLQueryType) IsRead() bool {
	return t == QueryTypeSelect
}
