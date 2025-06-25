package user

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	GroupID  string `json:"group_id"`
}

type UserGroup struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	RoleID string `json:"group_id"`
}

type UserRole struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
