package dto

import (
	"github.com/ecpartan/soap-server-tr069/pkg/errors"
)

type CreateUserInput struct {
	Username string
	Password string
	Email    string
	GroupId  string
}

func NewUserInput(
	username string,
	password string,
	email string,
	groupId string,
) (CreateUserInput, error) {

	//validate
	if username == "" {
		return CreateUserInput{}, errors.Wrap(nil, "Username is empty")
	}

	input := CreateUserInput{
		Username: username,
		Password: password,
		Email:    email,
		GroupId:  groupId,
	}

	return input, nil
}

type CreateUserOutput struct {
	Id       string
	Username string
	Email    string
	GroupId  string
}