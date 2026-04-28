package app

import (
	"github.com/g0p43r/tui_psql/internal/domain"
	"github.com/jackc/pgx/v5"
)

type connectSuccessMsg struct {
	profile domain.ConnectionProfile
	conn    *pgx.Conn
}

type connectErrorMsg struct {
	err error
}

type tablesLoadedMsg struct {
	tables []domain.DBObject
}

type tablesLoadErrorMsg struct {
	err error
}

type previewLoadedMsg struct {
	table  domain.DBObject
	result domain.QueryResult
}

type previewLoadErrorMsg struct {
	table domain.DBObject
	err   error
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
