package users

import (
	"database/sql"
	"errors"
)

type Provider struct {
	ID           int64  `json:"id"`
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Pronouns     string `json:"pronouns"`
	PracticeName string `json:"practice_name"`
	Address      string `json:"address"`
	Specialty    string `json:"specialty"`
	Phone        string `json:"phone"`
}

type sqlProvider struct {
	Provider
	Email        sql.NullString
	FirstName    sql.NullString
	LastName     sql.NullString
	Pronouns     sql.NullString
	PracticeName sql.NullString
	Address      sql.NullString
	Specialty    sql.NullString
	Phone        sql.NullString
}

func (p *sqlProvider) ToProvider() *Provider {
	var provider Provider
	provider.ID = p.ID
	provider.Email = p.Email.String
	provider.FirstName = p.FirstName.String
	provider.LastName = p.LastName.String
	provider.Pronouns = p.Pronouns.String
	provider.PracticeName = p.PracticeName.String
	provider.Address = p.Address.String
	provider.Specialty = p.Specialty.String
	provider.Phone = p.Phone.String
	return &provider
}

func ApproveProvider(userID int64, approved bool, db *sql.DB) error {
	_, err := db.Exec("update users set approvedProvider = ? where id = ?", approved, userID)
	if err != nil {
		return errors.New("error updating user. " + err.Error())
	}
	return nil
}

func GetApprovedProviders(db *sql.DB) ([]*Provider, error) {
	rows, err := db.Query("select id, email, firstName, lastName, pronouns, practiceName, address, specialty, phone from users where approvedProvider = true")
	if err != nil {
		return nil, errors.New("error getting approved providers. " + err.Error())
	}
	defer rows.Close()

	var providers []*Provider
	for rows.Next() {
		var dbProvider sqlProvider
		err := rows.Scan(
			&dbProvider.ID,
			&dbProvider.Email,
			&dbProvider.FirstName,
			&dbProvider.LastName,
			&dbProvider.Pronouns,
			&dbProvider.PracticeName,
			&dbProvider.Address,
			&dbProvider.Specialty,
			&dbProvider.Phone,
		)
		if err != nil {
			return nil, err
		}
		provider := dbProvider.ToProvider()
		providers = append(providers, provider)
	}
	return providers, nil
}

func GetApprovedProvider(id *int64, db *sql.DB) (*Provider, error) {
	row := db.QueryRow("select email, firstName, lastName, pronouns, practiceName, address, specialty, phone from users where id = ? and approvedProvider = true", id)
	var dbProvider sqlProvider
	err := row.Scan(
		&dbProvider.Email,
		&dbProvider.FirstName,
		&dbProvider.LastName,
		&dbProvider.Pronouns,
		&dbProvider.PracticeName,
		&dbProvider.Address,
		&dbProvider.Specialty,
		&dbProvider.Phone,
	)
	if err != nil {
		return nil, errors.New("error getting approved provider. " + err.Error())
	}
	dbProvider.ID = *id
	return dbProvider.ToProvider(), nil
}
