package domain

import "time"

type User struct {
	ID           int64     `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	PhoneNumber  string    `db:"phone_number"`
	Role         string    `db:"role"`
	CreatedAt    time.Time `db:"created_at"`
}

type Passenger struct {
	ID             int64     `db:"id"`
	UserID         int64     `db:"user_id"`
	FirstName      string    `db:"first_name"`
	LastName       string    `db:"last_name"`
	MiddleName     string    `db:"middle_name"`
	BirthDate      time.Time `db:"birth_date"`
	Gender         string    `db:"gender"`
	DocumentNumber string    `db:"document_number"`
	DocumentType   string    `db:"document_type"`
	Citizenship    string    `db:"citizenship"`
	CreatedAt      time.Time `db:"created_at"`
}
