package user

// RegisterRequest is the JSON body for the sign-up endpoint.
type RegisterRequest struct {
	// Email address of the new user
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
	// Password must be at least 8 characters
	Password string `json:"password" binding:"required,min=8" example:"securepass123"`
	// Phone number in international format
	PhoneNumber string `json:"phone_number" binding:"required" example:"+79991234567"`
}

// RegisterResponse is returned after successful registration.
type RegisterResponse struct {
	// Unique identifier of the created user
	UserID int64 `json:"user_id" example:"1"`
}

// LoginRequest is the JSON body for the sign-in endpoint.
type LoginRequest struct {
	// Email address of the user
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
	// Account password
	Password string `json:"password" binding:"required" example:"securepass123"`
}

// LoginResponse is returned after successful authentication.
type LoginResponse struct {
	// Unique identifier of the authenticated user
	UserID int64 `json:"user_id" example:"1"`
	// JWT access token
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
}

// UserResponse represents user profile data returned by GetUser.
type UserResponse struct {
	// Unique user identifier
	ID int64 `json:"id" example:"1"`
	// Email address
	Email string `json:"email" example:"user@example.com"`
	// Phone number
	PhoneNumber string `json:"phone_number" example:"+79991234567"`
	// User role: user or admin
	Role string `json:"role" example:"user" enums:"user, admin"`
	// Account creation timestamp in RFC 3339 format
	CreatedAt string `json:"created_at" example:"2025-01-15T10:30:00Z"`
}

// PassengerInput is the JSON body for adding a new passenger.
type PassengerInput struct {
	// Legal first name as in travel document
	FirstName string `json:"first_name" binding:"required" example:"Ivan"`
	// Legal last name as in travel document
	LastName string `json:"last_name" binding:"required" example:"Ivanov"`
	// Middle name (patronymic), optional
	MiddleName string `json:"middle_name" example:"Petrovich"`
	// Date of birth in YYYY-MM-DD format
	BirthDate string `json:"birth_date" binding:"required" example:"1990-05-20"`
	// Gender: male or female
	Gender string `json:"gender" binding:"required,oneof=male female" example:"male" enums:"male,female"`
	// Travel document number
	DocumentNumber string `json:"document_number" binding:"required" example:"7712345678"`
	// Travel document type
	DocumentType string `json:"document_type" binding:"required" example:"passport"`
	// Country of citizenship (ISO 3166-1 alpha-3)
	Citizenship string `json:"citizenship" binding:"required" example:"RUS"`
}

// AddPassengerResponse is returned after a passenger is successfully created.
type AddPassengerResponse struct {
	// Unique identifier of the created passenger
	PassengerID int64 `json:"passenger_id" example:"10"`
}

// PassengerResponse represents a single passenger in API responses.
type PassengerResponse struct {
	// Unique passenger identifier
	ID int64 `json:"id" example:"10"`
	// Owner user identifier
	UserID int64 `json:"user_id" example:"1"`
	// Legal first name
	FirstName string `json:"first_name" example:"Ivan"`
	// Legal last name
	LastName string `json:"last_name" example:"Ivanov"`
	// Middle name (patronymic)
	MiddleName string `json:"middle_name" example:"Petrovich"`
	// Date of birth in YYYY-MM-DD format
	BirthDate string `json:"birth_date" example:"1990-05-20"`
	// Gender: male, female
	Gender string `json:"gender" example:"male" enums:"male,female"`
	// Travel document number
	DocumentNumber string `json:"document_number" example:"7712345678"`
	// Travel document type
	DocumentType string `json:"document_type" example:"passport"`
	// Country of citizenship (ISO 3166-1 alpha-3)
	Citizenship string `json:"citizenship" example:"RUS"`
}
