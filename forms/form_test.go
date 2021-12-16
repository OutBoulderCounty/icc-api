package forms_test

import (
	"api/env"
	"api/forms"
	"strings"
	"testing"
)

const pathToDotEnv = "../.env"

func TestGetForm(t *testing.T) {
	e := env.TestSetup(t, true, pathToDotEnv)
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

func TestNewForm(t *testing.T) {
	e := env.TestSetup(t, true, pathToDotEnv)
	element := forms.Element{
		Label: "Test Element",
		Type:  "text",
	}
	newForm := forms.Form{
		Name:     "Test Form",
		Required: false,
		Live:     false,
		Elements: []*forms.Element{&element},
	}

	form, err := forms.NewForm(&newForm, e.DB)
	if err != nil {
		t.Error("error creating new form. " + err.Error())
	}
	if form.ID == 0 {
		t.Error("form id is 0")
	}
	if form.Name == "" {
		t.Error("form name is empty")
	}
}

func TestUpdateForm(t *testing.T) {
	e := env.TestSetup(t, true, pathToDotEnv)
	newForm := forms.Form{
		Name:     "Form to update",
		Required: false,
		Live:     false,
	}
	form, err := forms.NewForm(&newForm, e.DB)
	if err != nil {
		t.Error("error creating new form. " + err.Error())
		return
	}
	form.Name = "Updated Form"
	form.Elements = []*forms.Element{
		{
			Label:    "New element for updated form",
			Type:     "text",
			Position: 0,
			Required: false,
		},
	}
	err = forms.UpdateForm(form, e.DB)
	if err != nil {
		t.Error("error updating form. " + err.Error())
		return
	}
	updatedForm, err := forms.GetForm(form.ID, e.DB)
	if err != nil {
		t.Error("error getting updated form. " + err.Error())
		return
	}
	if updatedForm.Name != form.Name {
		t.Error("form name does not match")
	}
	if updatedForm.ID != form.ID {
		t.Error("form ID does not match")
	}
	if updatedForm.Required != form.Required {
		t.Error("form required value does not match")
	}
	if updatedForm.Live != form.Live {
		t.Error("form live value does not match")
	}
	if updatedForm.Elements[0].Label != form.Elements[0].Label {
		t.Error("form element label does not match")
	}
	if updatedForm.Elements[0].Type != form.Elements[0].Type {
		t.Error("form element type does not match")
	}

	// delete the form
	err = forms.DeleteForm(form.ID, e.DB)
	if err != nil {
		t.Error("error deleting form. " + err.Error())
		return
	}
	// validate the form is gone
	deletedForm, err := forms.GetForm(form.ID, e.DB)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			t.Log("form deleted successfully")
		} else {
			t.Error("error getting deleted form. " + err.Error())
			return
		}
	}
	if deletedForm != nil {
		t.Error("form still exists")
	}
}
