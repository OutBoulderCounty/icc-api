package responses_test

import (
	"api/env"
	"api/forms/responses"
	"api/users"
	"testing"
	"time"
)

const pathToDotEnv = "../../.env"

func TestNewResponse(t *testing.T) {
	e := env.TestSetup(t, true, pathToDotEnv)
	elementID := int64(1)
	userID := int64(1)
	value := "test"
	response, err := responses.NewResponse(elementID, userID, value, e.DB)
	if err != nil {
		t.Error(err)
		return
	}
	if response.ID == 0 {
		t.Error("Expected ID to be set")
	}
	if response.ElementID != elementID {
		t.Error("Expected ElementID to be", elementID, "; got", response.ElementID)
	}
	if response.UserID != userID {
		t.Error("Expected UserID to be", userID, "; got", response.UserID)
	}
	if response.Value != value {
		t.Error("Expected Value to be", value, "; got", response.Value)
	}
}

// make sure we can't create a response with a non-existent element
func TestNewResponseWithInvalidElement(t *testing.T) {
	e := env.TestSetup(t, true, pathToDotEnv)
	// generate an element ID that doesn't exist
	selectMaxElementID := "select max(id) from elements;"
	var maxElementID int64
	err := e.DB.QueryRow(selectMaxElementID).Scan(&maxElementID)
	if err != nil {
		t.Error("failed to get max element ID: " + err.Error())
		return
	}
	elementID := maxElementID + 1000000
	userID := int64(1)
	value := "test"
	_, err = responses.NewResponse(elementID, userID, value, e.DB)
	if err == nil {
		t.Error("Expected error when creating response with invalid element")
	}
}

// make sure we can't create a response with a non-existent user
func TestNewResponseWithInvalidUser(t *testing.T) {
	e := env.TestSetup(t, true, pathToDotEnv)
	elementID := int64(1)
	// generate a user ID that doesn't exist
	selectMaxUserID := "select max(id) from users;"
	var maxUserID int64
	err := e.DB.QueryRow(selectMaxUserID).Scan(&maxUserID)
	if err != nil {
		t.Error("failed to get max user ID: " + err.Error())
		return
	}
	userID := maxUserID + 1000000
	value := "test"
	_, err = responses.NewResponse(elementID, userID, value, e.DB)
	if err == nil {
		t.Error("Expected error when creating response with invalid user")
	}
}

func TestNewResponseWithOptions(t *testing.T) {
	e := env.TestSetup(t, true, pathToDotEnv)
	selectElement := "select distinct(elementID) from options;"
	var elementID int64
	err := e.DB.QueryRow(selectElement).Scan(&elementID)
	if err != nil {
		t.Error("failed to get element ID: " + err.Error())
		return
	}
	selectOptions := "select id from options where elementID = ?;"
	var optionIDs []int64
	rows, err := e.DB.Query(selectOptions, elementID)
	if err != nil {
		t.Error("failed to get options: " + err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		var optionID int64
		err := rows.Scan(&optionID)
		if err != nil {
			t.Error("failed to get option ID: " + err.Error())
			return
		}
		optionIDs = append(optionIDs, optionID)
	}
	userID := int64(1)

	currentTime := time.Now()
	response, err := responses.NewResponseWithOptions(elementID, userID, optionIDs, e.DB)
	if err != nil {
		t.Error("failed to create response with options: " + err.Error())
		return
	}
	if response.ID == 0 {
		t.Error("expected response ID to be set")
	}
	if response.ElementID != elementID {
		t.Error("expected ElementID to be", elementID, "; got", response.ElementID)
	}
	if response.UserID != userID {
		t.Error("expected UserID to be", userID, "; got", response.UserID)
	}
	currentMinute := currentTime.Round(time.Minute)
	createdMinute := response.CreatedAt.Round(time.Minute)
	if createdMinute != currentMinute {
		t.Error("expected CreatedAt to be", currentMinute, "; got", createdMinute)
	}
	// TODO: test that the response has the correct options
	if len(response.OptionIDs) != len(optionIDs) {
		t.Error("expected OptionIDs to be", optionIDs, "; got", response.OptionIDs)
	}
}

func TestNewResponseWithInvalidOptions(t *testing.T) {
	e := env.TestSetup(t, true, pathToDotEnv)
	// generate an invalid option ID
	selectMaxOptionID := "select max(id) from options;"
	var maxOptionID int64
	err := e.DB.QueryRow(selectMaxOptionID).Scan(&maxOptionID)
	if err != nil {
		t.Error("failed to get max option ID: " + err.Error())
		return
	}
	optionID := maxOptionID + 1000000
	elementID := int64(2)
	userID := int64(1)
	_, err = responses.NewResponseWithOptions(elementID, userID, []int64{optionID}, e.DB)
	if err == nil {
		t.Error("expected error when creating response with invalid option")
	}
}

func TestGetResponse(t *testing.T) {
	e := env.TestSetup(t, true, pathToDotEnv)
	selectResponse := "select id from responses;"
	var responseID int64
	err := e.DB.QueryRow(selectResponse).Scan(&responseID)
	if err != nil {
		t.Error("failed to get response ID: " + err.Error())
		return
	}
	response, err := responses.GetResponse(responseID, e.DB)
	if err != nil {
		t.Error("failed to get response: " + err.Error())
		return
	}
	if response.ID != responseID {
		t.Error("expected response ID to be", responseID, "; got", response.ID)
	}
	if response.ElementID == 0 {
		t.Error("expected ElementID to be set")
	}
	if response.UserID == 0 {
		t.Error("expected UserID to be set")
	}
	if response.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestGetResponseWithEmptyValue(t *testing.T) {
	e := env.TestSetup(t, true, pathToDotEnv)
	selectResponse := "select id from responses where value is null;"
	var responseID int64
	err := e.DB.QueryRow(selectResponse).Scan(&responseID)
	if err != nil {
		t.Error("failed to get response ID: " + err.Error())
		return
	}
	response, err := responses.GetResponse(responseID, e.DB)
	if err != nil {
		t.Error("failed to get response: " + err.Error())
		return
	}
	if response.Value != "" {
		t.Error("expected Value to be empty; got", response.Value)
	}
}

func TestGetResponseWithOptions(t *testing.T) {
	e := env.TestSetup(t, true, pathToDotEnv)
	selectResponse := "select id from responses where id in (select distinct responseID from response_options);"
	var responseID int64
	err := e.DB.QueryRow(selectResponse).Scan(&responseID)
	if err != nil {
		t.Error("failed to get response ID: " + err.Error())
		return
	}
	response, err := responses.GetResponse(responseID, e.DB)
	if err != nil {
		t.Error("failed to get response: " + err.Error())
		return
	}
	if response.OptionIDs == nil {
		t.Error("expected OptionIDs to be set")
	}
}

func TestGetFormResponsesByToken(t *testing.T) {
	e := env.TestSetup(t, true, pathToDotEnv)
	token := users.TestSessionToken
	responses, err := responses.GetFormResponsesByToken(token, e)
	if err != nil {
		t.Error("failed to get form responses: " + err.Error())
		return
	}
	if len(responses) == 0 {
		t.Error("expected form responses to be returned")
	}
	// each response should have a form name
	for _, response := range responses {
		if response.FormName == "" {
			t.Error("expected form name to be set")
		}
	}
}
