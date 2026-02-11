package user

type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	PhoneNumber string `json:"phone_number"`
}

type RegisterResponse struct {
	UserID int64 `json:"user_id"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	UserID int64  `json:"user_id"`
	Token  string `json:"token"`
}

type PassengerInput struct {
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	MiddleName     string `json:"middle_name"`
	BirthDate      string `json:"birth_date"`
	Gender         string `json:"gender"`
	DocumentNumber string `json:"document_number"`
	DocumentType   string `json:"document_type"`
	Citizenship    string `json:"citizenship"`
}

type AddPassengerResponse struct {
	PassengerID int64 `json:"passenger_id"`
}

type PassengerResponse struct {
	ID             int64  `json:"id"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	MiddleName     string `json:"middle_name"`
	BirthDate      string `json:"birth_date"`
	Gender         string `json:"gender"`
	DocumentNumber string `json:"document_number"`
	DocumentType   string `json:"document_type"`
	Citizenship    string `json:"citizenship"`
}
