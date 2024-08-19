package main

import "errors"

type User struct {
	id        int
	email     string
	points    int
	discounts *Discounts
}

func (u *User) UsePointsAsDiscount(points int) error {
	if points <= 0 {
		return errors.New("points must be greater than 0")
	}

	if u.points < points {
		return errors.New("not enough points")
	}

	u.points -= points
	u.discounts.AddNextOrderDiscount(points)

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

func (u *User) Discounts() *Discounts {
	return u.discounts
}

type Discounts struct {
	nextOrderDiscount int
}

func (c *Discounts) NextOrderDiscount() int {
	return c.nextOrderDiscount
}

func (c *Discounts) AddNextOrderDiscount(discount int) {
	c.nextOrderDiscount += discount
}

func UnmarshalUser(id int, email string, points int, discounts *Discounts) *User {
	return &User{
		id:        id,
		email:     email,
		points:    points,
		discounts: discounts,
	}
}

func UnmarshalDiscounts(discount int) *Discounts {
	return &Discounts{
		nextOrderDiscount: discount,
	}
}
