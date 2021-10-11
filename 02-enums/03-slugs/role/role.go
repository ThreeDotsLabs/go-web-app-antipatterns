package role

type Role string

const (
	Unknown   Role = ""
	Guest     Role = "guest"
	Member    Role = "member"
	Moderator Role = "moderator"
	Admin     Role = "admin"
)
