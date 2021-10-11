package main

import (
	"errors"
	"fmt"

	"github.com/ThreeDotsLabs/go-web-app-antipatterns/02-enums/02-typed-iota/role"
)

func CreateUser(r role.Role) error {
	if r == role.Unknown {
		return errors.New("no role provided")
	}

	fmt.Println("Creating user with role", r)

	return nil
}

func main() {
	err := CreateUser(0)
	if err != nil {
		fmt.Println(err)
	}

	err = CreateUser(role.Guest)
	if err != nil {
		fmt.Println(err)
	}

	err = CreateUser(role.Admin)
	if err != nil {
		fmt.Println(err)
	}

	err = CreateUser(role.Role(42))
	if err != nil {
		fmt.Println(err)
	}
}
