package dao

type User struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	GroupId  string `json:"group_id"`
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
