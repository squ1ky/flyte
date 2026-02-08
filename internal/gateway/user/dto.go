package user

type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	PhoneNumber string `json:"phone_number"`
}

type RegisterResponse struct {
	UserID int64 `json:"user_id"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type PassengerInput struct {
	FirstName      string `json:"first_name" binding:"required"`
	LastName       string `json:"last_name" binding:"required"`
	MiddleName     string `json:"middle_name"`
	BirthDate      string `json:"birth_date" binding:"required"`
	Gender         string `json:"gender" binding:"required,oneof=male female"`
	DocumentNumber string `json:"document_number" binding:"required"`
	DocumentType   string `json:"document_type" binding:"required"`
	Citizenship    string `json:"citizenship" binding:"required,len=3"`
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
