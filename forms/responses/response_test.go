package responses_test

import (
	"api/env"
	"api/forms/responses"
	"api/users"
	"errors"
	"fmt"
	"testing"
	"time"
)

const pathToDotEnv = "../../.env"

func TestNewResponse(t *testing.T) {
	e := env.TestSetup(t, false, pathToDotEnv)
	elementID := int64(1)
	userID, err := getTestUserID(e)
	if err != nil {
		t.Error(err.Error())
		return
	}
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

// make sure a user can't create a response if they have not accepted the user agreement
func TestNewResponseWithInvalidUserAgreement(t *testing.T) {
	e := env.TestSetup(t, false, pathToDotEnv)
	// get the current user agreement status for the test user
	selectUser := fmt.Sprintf("select id, agreementAccepted from users where email = '%s';", users.TestUser)
	var userID int64
	var userAgreementAccepted bool
	err := e.DB.QueryRow(selectUser).Scan(&userID, &userAgreementAccepted)
	if err != nil {
		t.Error("failed to get user agreement status: " + err.Error())
		return
	}
	if userAgreementAccepted {
		// temporarily set the user agreement status to false
		updateUser := fmt.Sprintf("update users set agreementAccepted = false where email = '%s';", users.TestUser)
		_, err = e.DB.Exec(updateUser)
		if err != nil {
			t.Error("failed to update user agreement status: " + err.Error())
			return
		}
		defer func() {
			// reset the user agreement status
			updateUserAccepted := fmt.Sprintf("update users set agreementAccepted = true where email = '%s';", users.TestUser)
			_, err = e.DB.Exec(updateUserAccepted)
			if err != nil {
				t.Error("failed to reset user agreement status: " + err.Error())
				return
			}
		}()
	}

	elementID := int64(1)
	_, err = responses.NewResponse(elementID, userID, "test", e.DB)
	if err == nil {
		t.Error("expected error when creating response with invalid user agreement")
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
	userID, err := getTestUserID(e)
	if err != nil {
		t.Error(err.Error())
		return
	}
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
	e := env.TestSetup(t, false, pathToDotEnv)
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
	userID, err := getTestUserID(e)
	if err != nil {
		t.Error(err.Error())
		return
	}

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

func getTestUserID(e *env.Env) (int64, error) {
	selectUser := fmt.Sprintf("select id from users where email = '%s';", users.TestUser)
	var userID int64
	err := e.DB.QueryRow(selectUser).Scan(&userID)
	if err != nil {
		return 0, errors.New("failed to get user ID: " + err.Error())
	}
	return userID, nil
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
	userID, err := getTestUserID(e)
	if err != nil {
		t.Error(err.Error())
		return
	}
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

func TestGetResponses(t *testing.T) {
	e := env.TestSetup(t, true, pathToDotEnv)
	resps, err := responses.GetResponses(e.DB)
	if err != nil {
		t.Error("failed to get responses: " + err.Error())
		return
	}
	if len(resps) == 0 {
		t.Error("expected at least one response")
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

func TestGetResponsesByForm(t *testing.T) {
	e := env.TestSetup(t, true, pathToDotEnv)
	formID := int64(1)
	responses, err := responses.GetResponsesByFormAndToken(formID, users.TestSessionToken, e)
	if err != nil {
		t.Error("failed to get responses: " + err.Error())
		return
	}
	if len(responses) == 0 {
		t.Error("expected at least one response")
	}
	user, err := users.GetUserBySession(users.TestSessionToken, e)
	if err != nil {
		t.Error("failed to get user: " + err.Error())
		return
	}
	for _, response := range responses {
		// check if any returned element IDs are not part of the form
		selectFormID := "select formID from elements where id = ?"
		var elementFormID int64
		err := e.DB.QueryRow(selectFormID, response.ElementID).Scan(&elementFormID)
		if err != nil {
			t.Error("failed to get form ID: " + err.Error())
			return
		}
		if elementFormID != formID {
			t.Error("expected response element to have form ID", formID, "; got", elementFormID)
		}
		if response.UserID != user.ID {
			t.Error("expected response to have user ID", user.ID, "; got", response.UserID)
		}
	}
}

func TestResponseApproval(t *testing.T) {
	e := env.TestSetup(t, true, pathToDotEnv)
	selectResponse := "select id from responses where approved = false"
	var responseID int64
	err := e.DB.QueryRow(selectResponse).Scan(&responseID)
	if err != nil {
		t.Error("failed to get response ID: " + err.Error())
		return
	}
	err = responses.ApproveResponse(responseID, true, e.DB)
	if err != nil {
		t.Error("failed to approve response: " + err.Error())
		return
	}
	// validate that the response was approved
	resp, err := responses.GetResponse(responseID, e.DB)
	if err != nil {
		t.Error("failed to get response: " + err.Error())
	}
	if !resp.Approved {
		t.Error("expected response to be approved")
	}
	// disapprove response
	err = responses.ApproveResponse(responseID, false, e.DB)
	if err != nil {
		t.Error("failed to disapprove response: " + err.Error())
	}
}
