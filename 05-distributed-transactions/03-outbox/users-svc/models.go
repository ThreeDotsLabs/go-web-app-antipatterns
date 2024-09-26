package main

import "errors"

type PointsUsedForDiscount struct {
	UserID int `json:"user_id"`
	Points int `json:"points"`
}

type User struct {
	id     int
	email  string
	points int
}

func (u *User) UsePoints(points int) error {
	if points <= 0 {
		return errors.New("points must be greater than 0")
	}

	if u.points < points {
		return errors.New("not enough points")
	}

	u.points -= points

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

func UnmarshalUser(id int, email string, points int) *User {
	return &User{
		id:     id,
		email:  email,
		points: points,
	}
}
