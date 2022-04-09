package main

import (
	"context"
	"errors"
)

type User struct {
	ID     string
	Active bool
}

func UserFromContext(ctx context.Context) (User, error) {
	u, ok := ctx.Value("user").(User)
	if !ok {
		return User{}, errors.New("could not get user from context")
	}
	return u, nil
}

func ContextWithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, "user", user)
}
