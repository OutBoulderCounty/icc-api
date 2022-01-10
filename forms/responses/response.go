package responses

import (
	"api/env"
	"api/users"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Response struct {
	ID        int64     `json:"id"`
	FormID    int64     `json:"form_id"`
	ElementID int64     `json:"element_id"`
	UserID    int64     `json:"user_id"`
	Value     string    `json:"value"`
	OptionIDs []int64   `json:"option_ids"`
	CreatedAt time.Time `json:"created_at"`
	Approved  bool      `json:"approved"`
}

type sqlResponse struct {
	Response
	Value sql.NullString
}

func (r *sqlResponse) ToResponse() *Response {
	resp := &Response{
		ID:        r.ID,
		FormID:    r.FormID,
		ElementID: r.ElementID,
		UserID:    r.UserID,
		Value:     r.Value.String,
		OptionIDs: r.OptionIDs,
		CreatedAt: r.CreatedAt,
		Approved:  r.Approved,
	}
	return resp
}

type FormResponse struct {
	FormID         int64     `json:"form_id"`
	FormName       string    `json:"form_name"`
	LastResponseAt time.Time `json:"last_responded_time"`
}

func NewResponse(elementID int64, userID int64, value string, db *sql.DB) (*Response, error) {
	// validate element exists
	selectElement := "SELECT id FROM elements WHERE id = ?"
	var selectedElement int64
	err := db.QueryRow(selectElement, elementID).Scan(&selectedElement)
	if err != nil {
		return nil, fmt.Errorf("error selecting element %v: %s", elementID, err.Error())
	}
	if selectedElement == 0 {
		return nil, fmt.Errorf("element %v not found", elementID)
	}

	// validate user
	err = validateUser(userID, db)
	if err != nil {
		return nil, err
	}

	resp := &Response{
		ElementID: elementID,
		UserID:    userID,
		Value:     value,
		CreatedAt: time.Now(),
	}
	result, err := db.Exec("INSERT INTO responses (elementID, userID, value, createdAt) VALUES (?, ?, ?, ?)", elementID, userID, value, resp.CreatedAt)
	if err != nil {
		return nil, errors.New("error inserting response: " + err.Error())
	}
	resp.ID, err = result.LastInsertId()
	if err != nil {
		return nil, errors.New("error getting last insert id: " + err.Error())
	}
	resp.FormID, err = getElementFormID(elementID, db)
	if err != nil {
		return nil, errors.New("error getting response form id: " + err.Error())
	}
	return resp, nil
}

func NewResponseWithOptions(elementID int64, userID int64, optionIDs []int64, db *sql.DB) (*Response, error) {
	err := validateElement(elementID, db)
	if err != nil {
		return nil, err
	}
	err = validateUser(userID, db)
	if err != nil {
		return nil, err
	}
	resp := &Response{
		ElementID: elementID,
		UserID:    userID,
		OptionIDs: optionIDs,
		CreatedAt: time.Now(),
	}
	result, err := db.Exec("INSERT INTO responses (elementID, userID, createdAt) VALUES (?, ?, ?)", elementID, userID, resp.CreatedAt)
	if err != nil {
		return nil, errors.New("error inserting response: " + err.Error())
	}
	resp.ID, err = result.LastInsertId()
	if err != nil {
		return nil, errors.New("error getting last insert id: " + err.Error())
	}

	// insert response options
	for _, optionID := range optionIDs {
		// validate the option exists for the element
		selectOption := "SELECT id FROM options WHERE id = ? AND elementID = ?"
		var selectedOption int64
		err := db.QueryRow(selectOption, optionID, elementID).Scan(&selectedOption)
		if err != nil {
			return nil, fmt.Errorf("error selecting option %v for element %v: %s", optionID, elementID, err.Error())
		}
		if selectedOption == 0 {
			return nil, fmt.Errorf("option %v not found for element %v", optionID, elementID)
		}

		_, err = db.Exec("INSERT INTO response_options (responseID, optionID) VALUES (?, ?)", resp.ID, optionID)
		if err != nil {
			return nil, errors.New("error inserting response options: " + err.Error())
		}
	}

	resp.FormID, err = getElementFormID(resp.ElementID, db)
	if err != nil {
		return nil, errors.New("error getting response form id: " + err.Error())
	}

	return resp, nil
}

func GetResponse(id int64, db *sql.DB) (*Response, error) {
	selectResponse := "SELECT id, elementID, userID, value, createdAt, approved FROM responses WHERE id = ?"
	var resp sqlResponse
	err := db.QueryRow(selectResponse, id).Scan(&resp.ID, &resp.ElementID, &resp.UserID, &resp.Value, &resp.CreatedAt, &resp.Approved)
	if err != nil {
		return nil, errors.New("error selecting response: " + err.Error())
	}
	resp.OptionIDs, err = getOptionsForResponse(id, db)
	if err != nil {
		return nil, errors.New("error getting response options: " + err.Error())
	}
	resp.FormID, err = getElementFormID(resp.ElementID, db)
	if err != nil {
		return nil, errors.New("error getting response form id: " + err.Error())
	}
	return resp.ToResponse(), nil
}

func GetResponses(db *sql.DB) ([]*Response, error) {
	selectResponses := "SELECT id, elementID, userID, value, createdAt, approved FROM responses"
	rows, err := db.Query(selectResponses)
	if err != nil {
		return nil, errors.New("error selecting responses: " + err.Error())
	}
	defer rows.Close()
	var responses []*Response
	for rows.Next() {
		var resp sqlResponse
		err := rows.Scan(&resp.ID, &resp.ElementID, &resp.UserID, &resp.Value, &resp.CreatedAt, &resp.Approved)
		if err != nil {
			return nil, errors.New("error scanning responses: " + err.Error())
		}
		resp.OptionIDs, err = getOptionsForResponse(resp.ID, db)
		if err != nil {
			return nil, errors.New("error getting response options: " + err.Error())
		}
		resp.FormID, err = getElementFormID(resp.ElementID, db)
		if err != nil {
			return nil, errors.New("error getting response form id: " + err.Error())
		}
		fmt.Println("Form ID:", resp.FormID)
		responses = append(responses, resp.ToResponse())
	}
	return responses, nil
}

func GetApprovedResponses(db *sql.DB) ([]*Response, error) {
	selectResponses := "SELECT id, elementID, userID, value, createdAt, approved FROM responses WHERE approved = true"
	rows, err := db.Query(selectResponses)
	if err != nil {
		return nil, errors.New("error selecting responses: " + err.Error())
	}
	defer rows.Close()
	var responses []*Response
	for rows.Next() {
		var resp sqlResponse
		err := rows.Scan(&resp.ID, &resp.ElementID, &resp.UserID, &resp.Value, &resp.CreatedAt, &resp.Approved)
		if err != nil {
			return nil, errors.New("error scanning responses: " + err.Error())
		}
		resp.OptionIDs, err = getOptionsForResponse(resp.ID, db)
		if err != nil {
			return nil, errors.New("error getting response options: " + err.Error())
		}
		resp.FormID, err = getElementFormID(resp.ElementID, db)
		if err != nil {
			return nil, errors.New("error getting response form id: " + err.Error())
		}
		responses = append(responses, resp.ToResponse())
	}
	return responses, nil
}

func GetApprovedResponsesByProvider(providerID int64, db *sql.DB) ([]*Response, error) {
	selectResponses := "SELECT id, elementID, userID, value, createdAt, approved FROM responses WHERE approved = true AND userID = ?"
	rows, err := db.Query(selectResponses, providerID)
	if err != nil {
		return nil, errors.New("error selecting responses: " + err.Error())
	}
	defer rows.Close()
	var responses []*Response
	for rows.Next() {
		var resp sqlResponse
		err := rows.Scan(&resp.ID, &resp.ElementID, &resp.UserID, &resp.Value, &resp.CreatedAt, &resp.Approved)
		if err != nil {
			return nil, errors.New("error scanning responses: " + err.Error())
		}
		resp.OptionIDs, err = getOptionsForResponse(resp.ID, db)
		if err != nil {
			return nil, errors.New("error getting response options: " + err.Error())
		}
		resp.FormID, err = getElementFormID(resp.ElementID, db)
		if err != nil {
			return nil, errors.New("error getting response form id: " + err.Error())
		}
		responses = append(responses, resp.ToResponse())
	}
	return responses, nil
}

func GetResponsesByProvider(providerID int64, db *sql.DB) ([]*Response, error) {
	selectResponses := "SELECT id, elementID, userID, value, createdAt, approved FROM responses WHERE userID = ?"
	rows, err := db.Query(selectResponses, providerID)
	if err != nil {
		return nil, errors.New("error selecting responses: " + err.Error())
	}
	defer rows.Close()
	var responses []*Response
	for rows.Next() {
		var resp sqlResponse
		err := rows.Scan(&resp.ID, &resp.ElementID, &resp.UserID, &resp.Value, &resp.CreatedAt, &resp.Approved)
		if err != nil {
			return nil, errors.New("error scanning responses: " + err.Error())
		}
		resp.OptionIDs, err = getOptionsForResponse(resp.ID, db)
		if err != nil {
			return nil, errors.New("error getting response options: " + err.Error())
		}
		resp.FormID, err = getElementFormID(resp.ElementID, db)
		if err != nil {
			return nil, errors.New("error getting response form id: " + err.Error())
		}
		fmt.Println("Form ID:", resp.FormID)
		responses = append(responses, resp.ToResponse())
	}
	return responses, nil
}

func getOptionsForResponse(responseID int64, db *sql.DB) ([]int64, error) {
	selectOptions := "SELECT optionID FROM response_options WHERE responseID = ?"
	rows, err := db.Query(selectOptions, responseID)
	if err != nil {
		return nil, errors.New("error selecting response options: " + err.Error())
	}
	defer rows.Close()
	var optionIDs []int64
	for rows.Next() {
		var optionID int64
		err := rows.Scan(&optionID)
		if err != nil {
			return nil, errors.New("error scanning response option: " + err.Error())
		}
		optionIDs = append(optionIDs, optionID)
	}
	return optionIDs, nil
}

func getElementFormID(elementID int64, db *sql.DB) (int64, error) {
	selectForm := "SELECT formID FROM elements WHERE id = ?"
	var formID int64
	err := db.QueryRow(selectForm, elementID).Scan(&formID)
	if err != nil {
		return 0, errors.New("error selecting form id: " + err.Error())
	}
	return formID, nil
}

func validateElement(elementID int64, db *sql.DB) error {
	selectElement := "SELECT id FROM elements WHERE id = ?"
	var selectedElement int64
	err := db.QueryRow(selectElement, elementID).Scan(&selectedElement)
	if err != nil {
		return fmt.Errorf("error selecting element %v: %s", elementID, err.Error())
	}
	if selectedElement == 0 {
		return fmt.Errorf("element %v not found", elementID)
	}
	return nil
}

func validateUser(userID int64, db *sql.DB) error {
	user, err := users.Get(userID, db)
	if err != nil {
		return errors.New("error getting user: " + err.Error())
	}
	if user.ID == 0 {
		return fmt.Errorf("user %v not found", userID)
	}
	// validate user accepted the user agreement
	if !user.AgreementAccepted {
		return errors.New("user must accept the user agreement")
	}
	return nil
}

func GetFormResponsesByToken(token string, e *env.Env) ([]*FormResponse, error) {
	user, err := users.GetUserBySession(token, e)
	if err != nil {
		return nil, errors.New("error getting user: " + err.Error())
	}
	selectFormResps := "select f.id, f.name, max(r.createdAt) from forms f, responses r where f.id = (select distinct e.formID from elements e where r.elementID = e.id) and userID = ? group by f.id"
	rows, err := e.DB.Query(selectFormResps, user.ID)
	if err != nil {
		return nil, errors.New("error selecting responses: " + err.Error())
	}
	defer rows.Close()
	var responses []*FormResponse
	for rows.Next() {
		var resp FormResponse
		err := rows.Scan(&resp.FormID, &resp.FormName, &resp.LastResponseAt)
		if err != nil {
			return nil, errors.New("error scanning form response: " + err.Error())
		}
		responses = append(responses, &resp)
	}
	return responses, nil
}

func GetApprovedFormResponses(db *sql.DB) ([]*FormResponse, error) {
	selectFormResps := "select f.id, f.name, max(r.createdAt) from forms f, responses r where f.id = (select distinct e.formID from elements e where r.elementID = e.id) and r.approved = true group by f.id"
	rows, err := db.Query(selectFormResps)
	if err != nil {
		return nil, errors.New("error selecting responses: " + err.Error())
	}
	defer rows.Close()
	var responses []*FormResponse
	for rows.Next() {
		var resp FormResponse
		err := rows.Scan(&resp.FormID, &resp.FormName, &resp.LastResponseAt)
		if err != nil {
			return nil, errors.New("error scanning form response: " + err.Error())
		}
		responses = append(responses, &resp)
	}
	return responses, nil
}

func GetResponsesByForm(formID int64, db *sql.DB) ([]*Response, error) {
	selectResponses := "SELECT id, elementID, userID, value, createdAt, approved FROM responses WHERE elementID IN (SELECT id FROM elements WHERE formID = ?)"
	rows, err := db.Query(selectResponses, formID)
	if err != nil {
		return nil, errors.New("error selecting responses: " + err.Error())
	}
	defer rows.Close()
	var responses []*Response
	for rows.Next() {
		var resp sqlResponse
		err := rows.Scan(&resp.ID, &resp.ElementID, &resp.UserID, &resp.Value, &resp.CreatedAt, &resp.Approved)
		if err != nil {
			return nil, errors.New("error scanning responses: " + err.Error())
		}
		resp.OptionIDs, err = getOptionsForResponse(resp.ID, db)
		if err != nil {
			return nil, errors.New("error getting response options: " + err.Error())
		}
		resp.FormID = formID
		responses = append(responses, resp.ToResponse())
	}
	return responses, nil
}

func GetApprovedResponsesByForm(formID int64, db *sql.DB) ([]*Response, error) {
	selectResponses := "SELECT id, elementID, userID, value, createdAt, approved FROM responses WHERE approved = true AND elementID IN (SELECT id FROM elements WHERE formID = ?)"
	rows, err := db.Query(selectResponses, formID)
	if err != nil {
		return nil, errors.New("error selecting responses: " + err.Error())
	}
	defer rows.Close()
	var responses []*Response
	for rows.Next() {
		var resp sqlResponse
		err := rows.Scan(&resp.ID, &resp.ElementID, &resp.UserID, &resp.Value, &resp.CreatedAt, &resp.Approved)
		if err != nil {
			return nil, errors.New("error scanning responses: " + err.Error())
		}
		resp.OptionIDs, err = getOptionsForResponse(resp.ID, db)
		if err != nil {
			return nil, errors.New("error getting response options: " + err.Error())
		}
		resp.FormID = formID
		responses = append(responses, resp.ToResponse())
	}
	return responses, nil
}

func GetResponsesByFormAndToken(formID int64, token string, e *env.Env) ([]*Response, error) {
	user, err := users.GetUserBySession(token, e)
	if err != nil {
		return nil, errors.New("error getting user: " + err.Error())
	}
	selectResponses := "SELECT r.id, r.elementID, r.userID, r.value, r.createdAt, r.approved FROM responses r, elements e WHERE r.elementID = e.id AND e.formID = ? AND r.userID = ?"
	rows, err := e.DB.Query(selectResponses, formID, user.ID)
	if err != nil {
		return nil, errors.New("error selecting responses: " + err.Error())
	}
	defer rows.Close()
	var responses []*Response
	for rows.Next() {
		var resp sqlResponse
		err := rows.Scan(&resp.ID, &resp.ElementID, &resp.UserID, &resp.Value, &resp.CreatedAt, &resp.Approved)
		if err != nil {
			return nil, errors.New("error scanning response: " + err.Error())
		}
		resp.OptionIDs, err = getOptionsForResponse(resp.ID, e.DB)
		if err != nil {
			return nil, errors.New("error getting response options: " + err.Error())
		}
		resp.FormID = formID
		responses = append(responses, resp.ToResponse())
	}
	return responses, nil
}

func ApproveResponse(id int64, approved bool, db *sql.DB) error {
	updateResponse := "UPDATE responses SET approved = ? WHERE id = ?"
	_, err := db.Exec(updateResponse, approved, id)
	if err != nil {
		return errors.New("error updating response: " + err.Error())
	}
	return nil
}
