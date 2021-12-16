package forms

import (
	"database/sql"
	"errors"
	"fmt"
)

type Form struct {
	ID       int64      `json:"id"`
	Name     string     `json:"name"`
	Required bool       `json:"required"`
	Live     bool       `json:"live"`
	Elements []*Element `json:"elements"`
}

type Element struct {
	ID       int64     `json:"id"`
	FormID   int64     `json:"form_id"`
	Label    string    `json:"label"`
	Type     string    `json:"type"`
	Position int       `json:"position"` // index
	Required bool      `json:"required"`
	Priority int       `json:"priority"`
	Search   bool      `json:"search"`
	Options  []*Option `json:"options"`
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
			element.Options = append(element.Options, &option)
		}
		form.Elements = append(form.Elements, &element)
	}

	return &form, nil
}

func NewForm(form *Form, db *sql.DB) (*Form, error) {
	resp, err := db.Exec("INSERT INTO forms (name, required, live) VALUES (?, ?, ?)", form.Name, form.Required, form.Live)
	if err != nil {
		return nil, errors.New("failed to insert form: " + err.Error())
	}
	id, err := resp.LastInsertId()
	if err != nil {
		return nil, errors.New("failed to get inserted form id: " + err.Error())
	}
	form.ID = id
	var elems []*Element
	for _, element := range form.Elements {
		element.FormID = id
		elem, err := NewElement(element, db)
		if err != nil {
			return nil, errors.New("failed to insert element: " + err.Error())
		}
		elems = append(elems, elem)
	}
	form.Elements = elems
	return form, nil
}

func NewElement(element *Element, db *sql.DB) (*Element, error) {
	resp, err := db.Exec("INSERT INTO elements (formID, label, type, position, required, priority, search) VALUES (?, ?, ?, ?, ?, ?, ?)", element.FormID, element.Label, element.Type, element.Position, element.Required, element.Priority, element.Search)
	if err != nil {
		return nil, errors.New("failed to insert element: " + err.Error())
	}
	id, err := resp.LastInsertId()
	if err != nil {
		return nil, errors.New("failed to get inserted element id: " + err.Error())
	}
	element.ID = id
	for i, option := range element.Options {
		option.ElementID = id
		option, err := NewOption(option, db)
		if err != nil {
			return nil, errors.New("failed to insert option: " + err.Error())
		}
		element.Options[i] = option
	}
	return element, nil
}

func NewOption(option *Option, db *sql.DB) (*Option, error) {
	resp, err := db.Exec("INSERT INTO options (elementID, name, position) VALUES (?, ?, ?)", option.ElementID, option.Name, option.Position)
	if err != nil {
		return nil, errors.New("failed to insert option: " + err.Error())
	}
	id, err := resp.LastInsertId()
	if err != nil {
		return nil, errors.New("failed to get inserted option id: " + err.Error())
	}
	option.ID = id
	return option, nil
}

func UpdateForm(form *Form, db *sql.DB) error {
	_, err := db.Exec("UPDATE forms SET name = ?, required = ?, live = ? WHERE id = ?", form.Name, form.Required, form.Live, form.ID)
	if err != nil {
		return errors.New("failed to update form: " + err.Error())
	}
	for _, element := range form.Elements {
		if element.ID > 0 {
			err := UpdateElement(element, db)
			if err != nil {
				return errors.New("failed to update element: " + err.Error())
			}
		} else {
			element.FormID = form.ID
			_, err := NewElement(element, db)
			if err != nil {
				return errors.New("failed to create element: " + err.Error())
			}
		}
	}
	return nil
}

func UpdateElement(element *Element, db *sql.DB) error {
	_, err := db.Exec("UPDATE elements SET label = ?, type = ?, position = ?, required = ?, priority = ?, search = ? WHERE id = ?", element.Label, element.Type, element.Position, element.Required, element.Priority, element.Search, element.ID)
	if err != nil {
		return errors.New("failed to update element: " + err.Error())
	}
	for _, option := range element.Options {
		if option.ID > 0 {
			err := UpdateOption(option, db)
			if err != nil {
				return errors.New("failed to update option: " + err.Error())
			}
		} else {
			option.ElementID = element.ID
			_, err := NewOption(option, db)
			if err != nil {
				return errors.New("failed to create option: " + err.Error())
			}
		}
	}
	return nil
}

func UpdateOption(option *Option, db *sql.DB) error {
	_, err := db.Exec("UPDATE options SET name = ?, position = ? WHERE id = ?", option.Name, option.Position, option.ID)
	if err != nil {
		return errors.New("failed to update option: " + err.Error())
	}
	return nil
}

func DeleteForm(id int64, db *sql.DB) error {
	_, err := db.Exec("DELETE FROM forms WHERE id = ?", id)
	if err != nil {
		return errors.New("failed to delete form: " + err.Error())
	}
	_, err = db.Exec("DELETE FROM elements WHERE formID = ?", id)
	if err != nil {
		return errors.New("failed to delete elements: " + err.Error())
	}
	_, err = db.Exec("DELETE FROM options WHERE elementID IN (SELECT id FROM elements WHERE formID = ?)", id)
	if err != nil {
		return errors.New("failed to delete options: " + err.Error())
	}
	return nil
}
