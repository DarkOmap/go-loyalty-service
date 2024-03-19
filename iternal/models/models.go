package models

import (
	"encoding/json"
	"fmt"
)

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Salt     string `json:"-"`
}

func NewUserByJSON(j []byte) (*User, error) {
	var u User
	err := json.Unmarshal(j, &u)

	if err != nil {
		return nil, fmt.Errorf("unmarshall user json %s: %w", string(j), err)
	}

	return &u, nil
}
