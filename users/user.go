package users

import (
	"database/sql"
)

type User struct {
	ID           int64
	StytchUserID string
	Email        string
	ActiveRoles  []string
	FirstName    string
	LastName     string
	Pronouns     string
	PracticeName string
	Address      string
	Specialty    string
	Phone        string
	ImageURL     string
}

// retrieves a single user from the database
func Get(id int64, db *sql.DB) (*User, error) {
	row := db.QueryRow("SELECT id, stytchUserID, email, firstName, lastName, pronouns, practiceName, address, specialty, phone, imageURL FROM users WHERE id = ?", id)
	var user User
	err := row.Scan(
		&user.ID,
		&user.StytchUserID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Pronouns,
		&user.PracticeName,
		&user.Address,
		&user.Specialty,
		&user.Phone,
		&user.ImageURL,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
