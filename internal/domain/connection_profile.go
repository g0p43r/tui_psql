package domain

type ConnectionProfile struct {
	Name     string
	Host     string
	Port     string
	Database string
	User     string
	Password string
	SSLMode  string
}
