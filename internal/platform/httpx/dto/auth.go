package dto

type RegisterOwnerRequest struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	CompanyName string `json:"company_name"`
	Phone       string `json:"phone"`
	Password    string `json:"password"`
}

type LoginRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}
