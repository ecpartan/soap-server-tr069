package entity

import "github.com/ecpartan/soap-server-tr069/utils"

type User struct {
	ID       utils.ID `json:"id"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Email    string   `json:"email"`
	GroupId  string   `json:"group_id"`
}

type UserView struct {
	Username string
	Password string
	Email    string
}

func NewUserView(username, password, email string) (*UserView, error) {
	return &UserView{Username: username, Password: password, Email: email}, nil
}

type UserRole struct {
	ID          utils.ID `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
}

type UserGroup struct {
	ID     utils.ID `json:"id"`
	Name   string   `json:"name"`
	Roleid utils.ID `json:"role_id"`
}

func NewUser(view UserView) (*User, error) {
	return &User{ID: utils.NewID(), Username: view.Username, Password: view.Password, Email: view.Email}, nil
}

func (u *User) Validate() error {
	return nil
}
