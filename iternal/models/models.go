package models

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Salt     string `json:"-"`
}

func (rs *User) ScanRow(rows pgx.Rows) error {
	values, err := rows.Values()
	if err != nil {
		return err
	}

	for i := range values {
		switch strings.ToLower(rows.FieldDescriptions()[i].Name) {
		case "login":
			rs.Login = values[i].(string)
		case "password":
			rs.Password = values[i].(string)
		case "salt":
			rs.Salt = values[i].(string)
		}
	}

	return nil
}

func NewUserByJSON(j []byte) (*User, error) {
	var u User
	err := json.Unmarshal(j, &u)

	if err != nil {
		return nil, fmt.Errorf("unmarshall user json %s: %w", string(j), err)
	}

	return &u, nil
}
