package user

type User struct {
	Id           string `json:"id"`
	CompanyId    string `json:"company_id"`
	RoleId       string `json:"role_id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Phone        string `json:"phone"`
	PasswordHash string
}
