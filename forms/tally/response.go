package tally

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type PrettyResponse struct {
	ID            int64     `json:"id"`
	FormName      string    `json:"form_name"`
	CreatedAt     time.Time `json:"created_at"`
	UserFirstName string    `json:"user_first_name"`
	UserLastName  string    `json:"user_last_name"`
	UserEmail     string    `json:"user_email"`
	Questions     []QnA     `json:"questions"`
}

type QnA struct {
	Key      string      `json:"key"`
	Question string      `json:"question"`
	Answer   interface{} `json:"answer"`
	Options  interface{} `json:"options,omitempty"`
}

func (q *QnA) UnmarshalJSON(data []byte) error {
	var field Field
	err := json.Unmarshal(data, &field)
	if err != nil {
		return errors.New("error unmarshalling field: " + err.Error())
	}
	q.Key = field.Key
	q.Question = field.Label
	q.Answer = field.Value
	q.Options = field.Options
	return nil
}

func GetPrettyResponse(id int64, db *sql.DB) (*PrettyResponse, error) {
	query := "select r.id, f.name, r.created_at, u.firstName, u.lastName, u.email, r.fields from tally_responses r, users u, tally_forms f where r.user_id = u.id and r.form_id = f.id and r.id = ?"
	row := db.QueryRow(query, id)
	var response PrettyResponse
	var fields []byte
	var firstName sql.NullString
	var lastName sql.NullString
	err := row.Scan(&response.ID, &response.FormName, &response.CreatedAt, &firstName, &lastName, &response.UserEmail, &fields)
	if err != nil {
		return nil, errors.New("error getting response from database: " + err.Error())
	}
	response.UserFirstName = firstName.String
	response.UserLastName = lastName.String
	err = json.Unmarshal(fields, &response.Questions)
	if err != nil {
		return nil, errors.New("error unmarshalling fields: " + err.Error())
	}
	return &response, nil
}
