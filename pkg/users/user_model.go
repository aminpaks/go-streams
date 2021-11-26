package users

import (
	"encoding/json"

	"github.com/google/uuid"
)

func NewUser(name, email string) *User {
	return &User{
		Id:    uuid.New(),
		Name:  name,
		Email: email,
	}
}

type User struct {
	Id    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

func (u *User) WithId(id uuid.UUID) *User {
	u.Id = id
	return u
}

func (u *User) Bytes() []byte {
	b, _ := json.Marshal(u)
	return b
}

func (u *User) String() string {
	return string(u.Bytes())
}

func ParseUser(v string) (*User, error) {
	var user User
	err := json.Unmarshal([]byte(v), &user)
	return &user, err
}

func ParseUserWithOptionalId(v []byte) (*User, error) {
	type userCreation struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	var user *userCreation = &userCreation{}
	err := json.Unmarshal(v, user)
	if err != nil {
		return nil, err
	}

	return NewUser(user.Name, user.Email), nil
}
