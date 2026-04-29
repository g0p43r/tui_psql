package app

import (
	"github.com/g0p43r/tui_psql/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type connectSuccessMsg struct {
	profile domain.ConnectionProfile
	pool    *pgxpool.Pool
}

type connectErrorMsg struct {
	err error
}

type tablesLoadedMsg struct {
	requestID int
	tables    []domain.DBObject
}

type tablesLoadErrorMsg struct {
	requestID int
	err       error
}

type previewLoadedMsg struct {
	requestID int
	table     domain.DBObject
	result    domain.QueryResult
}

type previewLoadErrorMsg struct {
	requestID int
	table     domain.DBObject
	err       error
}

type sqlExecutedMsg struct {
	sessionID      int
	queryType      domain.SQLQueryType
	editorMode     string
	queryWorkbench bool
	newTableSchema string
	newTableName   string
	targetSchema   string
	targetTable    string
	sql            string
	result         domain.QueryResult
}

type sqlExecuteErrorMsg struct {
	sessionID      int
	queryType      domain.SQLQueryType
	editorMode     string
	queryWorkbench bool
	targetSchema   string
	targetTable    string
	sql            string
	err            error
}

type profilesLoadedMsg struct {
	profiles []domain.ConnectionProfile
	err      error
}

type profileSavedMsg struct {
	profiles []domain.ConnectionProfile
	err      error
}

type profileDeletedMsg struct {
	profiles []domain.ConnectionProfile
	err      error
}
