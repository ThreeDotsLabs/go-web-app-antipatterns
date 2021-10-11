package role

import "errors"

type Role struct {
	slug string
}

func (r Role) String() string {
	return r.slug
}

var (
	Unknown   = Role{""}
	Guest     = Role{"guest"}
	Member    = Role{"member"}
	Moderator = Role{"moderator"}
	Admin     = Role{"admin"}
)

func FromString(s string) (Role, error) {
	switch s {
	case Guest.slug:
		return Guest, nil
	case Member.slug:
		return Member, nil
	case Moderator.slug:
		return Moderator, nil
	case Admin.slug:
		return Admin, nil
	}

	return Unknown, errors.New("unknown role: " + s)
}
