package entity

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        string    `json:"id" dynamodbav:"id"`
	Email     string    `json:"email" dynamodbav:"email"`
	Password  string    `json:"-" dynamodbav:"password"`
	CreatedAt time.Time `json:"created_at" dynamodbav:"created_at"`
}

func NewUser(id, email, password string) (*User, error) {
	if email == "" {
		return nil, errors.New("o e-mail não pode ser vazio")
	}
	if len(password) < 6 {
		return nil, errors.New("a senha deve ter pelo menos 6 caracteres")
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:        id,
		Email:     email,
		Password:  hashedPassword,
		CreatedAt: time.Now().UTC(),
	}, nil
}

// RecreateExistingUser is a factory method to reconstitute a user from the repository without rehashing the password
func RecreateExistingUser(id, email, hashedPassword string, createdAt time.Time) *User {
	return &User{
		ID:        id,
		Email:     email,
		Password:  hashedPassword,
		CreatedAt: createdAt,
	}
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
