package responses

import (
	"database/sql"
	"errors"
	"fmt"
)

type Response struct {
	ID        int64
	ElementID int64
	UserID    int64
	Value     string
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
