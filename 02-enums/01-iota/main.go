package main

import (
	"fmt"

	"github.com/ThreeDotsLabs/go-web-app-antipatterns/02-enums/01-iota/role"
)

func CreateUser(r int) error {
	fmt.Println("Creating user with role", r)
	return nil
}

func main() {
	err := CreateUser(-1)
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

	err = CreateUser(42)
	if err != nil {
		fmt.Println(err)
	}
}
