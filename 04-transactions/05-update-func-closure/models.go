package main

import "errors"

type User struct {
	id     int
	email  string
	points int
	cart   *Cart
}

func (u *User) UsePointsAsDiscount(points int) error {
	if points <= 0 {
		return errors.New("points must be greater than 0")
	}

	if u.points < points {
		return errors.New("not enough points")
	}

	u.points -= points
	u.cart.AddDiscount(points)

	return nil
}

func (u *User) ID() int {
	return u.id
}

func (u *User) Email() string {
	return u.email
}

func (u *User) Points() int {
	return u.points
}

func (u *User) Cart() *Cart {
	return u.cart
}

type Cart struct {
	discount int
}

func (c *Cart) Discount() int {
	return c.discount
}

func (c *Cart) AddDiscount(discount int) {
	c.discount += discount
}

func UnmarshalUser(id int, email string, points int, cart *Cart) *User {
	return &User{
		id:     id,
		email:  email,
		points: points,
		cart:   cart,
	}
}

func UnmarshalCart(discount int) *Cart {
	return &Cart{
		discount: discount,
	}
}
