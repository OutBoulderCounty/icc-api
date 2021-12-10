package responses

import (
	"database/sql"
	"errors"
	"fmt"
)

type Response struct {
	ID        int64
	ElementID int64 `json:"element_id"`
	UserID    int64
	Value     string  `json:"value"`
	OptionIDs []int64 `json:"option_ids"`
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
	}
	result, err := db.Exec("INSERT INTO responses (elementID, userID, value) VALUES (?, ?, ?)", elementID, userID, value)
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
	}
	result, err := db.Exec("INSERT INTO responses (elementID, userID) VALUES (?, ?)", elementID, userID)
	if err != nil {
		return nil, errors.New("error inserting response: " + err.Error())
	}
	resp.ID, err = result.LastInsertId()
	if err != nil {
		return nil, errors.New("error getting last insert id: " + err.Error())
	}

	// insert response options
	for _, optionID := range optionIDs {
		_, err = db.Exec("INSERT INTO response_options (responseID, optionID) VALUES (?, ?)", resp.ID, optionID)
		if err != nil {
			return nil, errors.New("error inserting response options: " + err.Error())
		}
	}

	return resp, nil
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
