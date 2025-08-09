package entity

import "github.com/ecpartan/soap-server-tr069/utils"

type User struct {
	ID       utils.ID `json:"id"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Email    string   `json:"email"`
	GroupId  string   `json:"group_id"`
}

type CreateUser struct {
	Username string
	Password string
	Email    string
}

type UserRole struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UserGroup struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Roleid string `json:"role_id"`
}

func NewUser(view CreateUser) (*User, error) {
	return &User{ID: utils.NewID(), Username: view.Username, Password: view.Password, Email: view.Email}, nil
}

func (u *User) Validate() error {
	return nil
}
