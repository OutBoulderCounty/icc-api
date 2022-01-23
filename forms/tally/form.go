package tally

import (
	"database/sql"
	"errors"
)

type Form struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Required bool   `json:"required"`
}

func (e *Event) RegisterForm(db *sql.DB) (*Form, error) {
	var form Form
	fields := e.Data.Fields
	for i := 0; i < len(fields); i++ {
		field := fields[i]
		switch field.Label {
		case "Name":
			form.Name = field.Value.(string)
		case "URL":
			form.URL = field.Value.(string)
		}
		for j := 0; j < len(field.Options); j++ {
			option := field.Options[j]
			if option.Text == "Required" {
				if field.Value != nil {
					list := field.Value.([]string)
					for k := 0; k < len(list); k++ {
						if list[k] == option.ID {
							form.Required = true
							break
						}
					}
				}
			}
		}
	}

	query := "insert into tally_forms (name, url, required) values (?, ?, ?)"
	result, err := db.Exec(query, form.Name, form.URL, form.Required)
	if err != nil {
		return nil, errors.New("error inserting form into database: " + err.Error())
	}
	form.ID, err = result.LastInsertId()
	if err != nil {
		return nil, errors.New("error getting form ID: " + err.Error())
	}

	return &form, nil
}

func GetForms(db *sql.DB) ([]*Form, error) {
	query := "select id, name, url, required from tally_forms"
	rows, err := db.Query(query)
	if err != nil {
		return nil, errors.New("error getting forms: " + err.Error())
	}
	defer rows.Close()

	var forms []*Form
	for rows.Next() {
		var form Form
		err = rows.Scan(&form.ID, &form.Name, &form.URL, &form.Required)
		if err != nil {
			return nil, errors.New("error scanning form: " + err.Error())
		}
		forms = append(forms, &form)
	}
	err = rows.Err()
	if err != nil {
		return nil, errors.New("error getting forms: " + err.Error())
	}

	return forms, nil
}
