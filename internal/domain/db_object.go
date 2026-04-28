package domain

type DBObjectType string

const (
	ObjectSchema DBObjectType = "schema"
	ObjectTable  DBObjectType = "table"
	ObjectView   DBObjectType = "view"
)

type DBObject struct {
	Schema string
	Name   string
	Type   DBObjectType
}
