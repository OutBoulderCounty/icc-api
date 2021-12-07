package forms_test

import (
	"api/env"
	"api/forms"
	"testing"
)

func TestGetForm(t *testing.T) {
	e := env.TestSetup(t, true)
	// get a form that has elements and options
	selectForms := "select id from forms where id in (select distinct formID from elements where id in (select distinct elementID from options))"
	row := e.DB.QueryRow(selectForms)
	var formID int64
	err := row.Scan(&formID)
	if err != nil {
		t.Error("error getting form ID. " + err.Error())
	}

	form, err := forms.GetForm(formID, e.DB)
	if err != nil {
		t.Error("error getting form. " + err.Error())
	}
	if form.ID != formID {
		t.Error("form id does not match")
	}
	if form.Name == "" {
		t.Error("form name is empty")
	}
	if len(form.Elements) == 0 {
		t.Error("form elements is empty")
		return
	}
	if len(form.Elements[0].Options) == 0 {
		t.Error("form element options is empty")
	}
}
