package models

type User struct {
	ID       uint
	Username string
	Password string // need to hash this
	Email    string
}
