package forms

import (
	"database/sql"
	"errors"
	"fmt"
)

type Form struct {
	ID       int64     `json:"id"`
	Name     string    `json:"name"`
	Required bool      `json:"required"`
	Live     bool      `json:"live"`
	Elements []Element `json:"elements"`
}

type Element struct {
	ID       int64    `json:"id"`
	FormID   int64    `json:"form_id"`
	Label    string   `json:"label"`
	Type     string   `json:"type"`
	Position int      `json:"position"` // index
	Required bool     `json:"required"`
	Priority int      `json:"priority"`
	Search   bool     `json:"search"`
	Options  []Option `json:"options"`
}

type Option struct {
	ID        int64  `json:"id"`
	ElementID int64  `json:"element_id"`
	Name      string `json:"name"`
	Position  int    `json:"position"` // index
}

func GetForms(db *sql.DB) ([]Form, error) {
	forms := []Form{}

	selectForms := "SELECT id, name, required, live FROM forms"
	rows, err := db.Query(selectForms)
	if err != nil {
		fmt.Println("Failed SQL: " + selectForms)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var form Form
		err := rows.Scan(&form.ID, &form.Name, &form.Required, &form.Live)
		if err != nil {
			return nil, err
		}
		forms = append(forms, form)
	}

	return forms, nil
}

func GetForm(id int64, db *sql.DB) (*Form, error) {
	var form Form
	selectForm := "SELECT id, name, required, live FROM forms WHERE id = ?"
	err := db.QueryRow(selectForm, id).Scan(&form.ID, &form.Name, &form.Required, &form.Live)
	if err != nil {
		return nil, errors.New("failed to get form: " + err.Error())
	}
	selectElements := "SELECT id, formID, label, type, position, required, priority, search FROM elements WHERE formID = ?"
	rows, err := db.Query(selectElements, form.ID)
	if err != nil {
		return nil, errors.New("failed to get elements: " + err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var element Element
		err := rows.Scan(&element.ID, &element.FormID, &element.Label, &element.Type, &element.Position, &element.Required, &element.Priority, &element.Search)
		if err != nil {
			return nil, errors.New("failed to scan element: " + err.Error())
		}
		selectOptions := "SELECT id, elementID, name, position FROM options WHERE elementID = ?"
		optionRows, err := db.Query(selectOptions, element.ID)
		if err != nil {
			return nil, errors.New("failed to get options: " + err.Error())
		}
		defer optionRows.Close()
		for optionRows.Next() {
			var option Option
			err := optionRows.Scan(&option.ID, &option.ElementID, &option.Name, &option.Position)
			if err != nil {
				return nil, errors.New("failed to scan option: " + err.Error())
			}
			element.Options = append(element.Options, option)
		}
		form.Elements = append(form.Elements, element)
	}

	return &form, nil
}
