package role

type Role uint

const (
	Unknown Role = iota
	Guest
	Member
	Moderator
	Admin
)
