package forms

import (
	"database/sql"
	"fmt"
)

type Form struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Live     bool   `json:"live"`
}

type Element struct {
	ID       int64  `json:"id"`
	FormID   int64  `json:"formID"`
	Label    string `json:"label"`
	Type     string `json:"type"`
	Position int    `json:"position"` // index
	Required bool   `json:"required"`
	Priority int    `json:"priority"`
	Search   bool   `json:"search"`
}

type Option struct {
	ID        int64  `json:"id"`
	ElementID int64  `json:"elementID"`
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
