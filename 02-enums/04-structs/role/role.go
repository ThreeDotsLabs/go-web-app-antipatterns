package role

import (
	"errors"
	"fmt"
)

var (
	Guest     = Role{"guest"}
	Member    = Role{"member"}
	Moderator = Role{"moderator"}
	Admin     = Role{"admin"}

	ErrUnknownRole = errors.New("unknown role")
)

type Role struct {
	slug string
}

func (r Role) String() string {
	return r.slug
}

func FromString(s string) (*Role, error) {
	switch s {
	case Guest.slug:
		return &Guest, nil
	case Member.slug:
		return &Member, nil
	case Moderator.slug:
		return &Moderator, nil
	case Admin.slug:
		return &Admin, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownRole, s)
	}
}
