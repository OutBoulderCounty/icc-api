package responses

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Response struct {
	ID        int64     `json:"id"`
	ElementID int64     `json:"element_id"`
	UserID    int64     `json:"user_id"`
	Value     string    `json:"value"`
	OptionIDs []int64   `json:"option_ids"`
	CreatedAt time.Time `json:"created_at"`
}

type sqlResponse struct {
	Response
	Value sql.NullString
}

func (r *sqlResponse) ToResponse() *Response {
	resp := &Response{
		ID:        r.ID,
		ElementID: r.ElementID,
		UserID:    r.UserID,
		Value:     r.Value.String,
		CreatedAt: r.CreatedAt,
	}
	return resp
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

	// validate user exists
	selectUser := "SELECT id FROM users WHERE id = ?"
	var selectedUser int64
	err = db.QueryRow(selectUser, userID).Scan(&selectedUser)
	if err != nil {
		return nil, fmt.Errorf("error selecting user %v: %s", userID, err.Error())
	}
	if selectedUser == 0 {
		return nil, fmt.Errorf("user %v not found", userID)
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

	return resp, nil
}

func GetResponse(id int64, db *sql.DB) (*Response, error) {
	selectResponse := "SELECT id, elementID, userID, value, createdAt FROM responses WHERE id = ?"
	var resp sqlResponse
	err := db.QueryRow(selectResponse, id).Scan(&resp.ID, &resp.ElementID, &resp.UserID, &resp.Value, &resp.CreatedAt)
	if err != nil {
		return nil, errors.New("error selecting response: " + err.Error())
	}
	return resp.ToResponse(), nil
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
	selectUser := "SELECT id FROM users WHERE id = ?"
	var selectedUser int64
	err := db.QueryRow(selectUser, userID).Scan(&selectedUser)
	if err != nil {
		return fmt.Errorf("error selecting user %v: %s", userID, err.Error())
	}
	if selectedUser == 0 {
		return fmt.Errorf("user %v not found", userID)
	}
	return nil
}
