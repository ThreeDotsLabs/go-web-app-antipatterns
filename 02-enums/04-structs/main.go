package main

import (
	"errors"
	"fmt"

	"github.com/ThreeDotsLabs/go-web-app-antipatterns/02-enums/04-structs/role"
)

func CreateUser(r role.Role) error {
	if r == role.Unknown {
		return errors.New("no role provided")
	}

	fmt.Println("Creating user with role", r)

	return nil
}

func main() {
	err := CreateUser(role.Role{})
	if err != nil {
		fmt.Println(err)
	}

	err = CreateUser(role.Guest)
	if err != nil {
		fmt.Println(err)
	}

	admin, err := role.FromString("admin")
	if err != nil {
		fmt.Println(err)
	}

	err = CreateUser(admin)
	if err != nil {
		fmt.Println(err)
	}

	_, err = role.FromString("super-admin")
	if err != nil {
		fmt.Println(err)
	}
}
