package main

import (
	"api/forms"
	"api/forms/responses"
	"api/users"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRoutes(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	env := setup()
	defer env.DB.Close()

	body, err := json.Marshal(users.UserReq{
		Email:       users.TestUser,
		RedirectURL: users.TestRedirectURL,
		Roles:       []string{"provider"},
	})
	if err != nil {
		t.Error("Failed to marshal user req body: " + err.Error())
	}
	providerLoginReq, err := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	if err != nil {
		t.Error("Failed to create request: " + err.Error())
	}

	authTokenBody, err := json.Marshal(users.Auth{
		Token: users.TestToken,
	})
	if err != nil {
		t.Error("Failed to marshal auth token body: " + err.Error())
	}
	authTokenReq, err := http.NewRequest("POST", "/authenticate", bytes.NewBuffer(authTokenBody))
	if err != nil {
		t.Error("Failed to create request: " + err.Error())
	}

	authTokenBodyTest := func(t *testing.T, bdy []byte) bool {
		var sessAuth users.SessionAuth
		err := json.Unmarshal(bdy, &sessAuth)
		if err != nil {
			t.Error("Failed to unmarshal response body: " + err.Error())
			return false
		}
		// TODO: validate session auth
		if sessAuth.SessionToken == "" {
			t.Error("Session token is empty")
			return false
		}
		return true
	}

	getFormsReq, err := http.NewRequest("GET", "/forms", nil)
	if err != nil {
		t.Error("Failed to create request: " + err.Error())
	}
	getFormsReq.Header.Set("Authorization", users.TestSessionToken)

	getFormsBodyTest := func(t *testing.T, bdy []byte) bool {
		t.Log("Body: " + string(bdy))
		type formList struct {
			Forms []forms.Form `json:"forms"`
		}
		var forms formList
		err := json.Unmarshal(bdy, &forms)
		if err != nil {
			t.Error("Failed to unmarshal response body: " + err.Error())
			return false
		}
		if len(forms.Forms) >= 1 {
			return true
		}
		return false
	}

	getFormReq, err := http.NewRequest("GET", "/form/1", nil)
	if err != nil {
		t.Error("Failed to create get form request: " + err.Error())
	}
	getFormReq.Header.Set("Authorization", users.TestSessionToken)

	getFormBodyTest := func(t *testing.T, bdy []byte) bool {
		type formResp struct {
			Form forms.Form `json:"form"`
		}
		var form formResp
		err := json.Unmarshal(bdy, &form)
		if err != nil {
			t.Error("Failed to unmarshal get form response body: " + err.Error())
			return false
		}
		if form.Form.ID != 1 {
			t.Error("Form ID is incorrect")
			return false
		}
		if len(form.Form.Elements) == 0 {
			t.Error("Form has no elements")
			return false
		}
		return true
	}

	updateUserBody, err := json.Marshal(users.User{
		FirstName:         "Integration",
		LastName:          "Test",
		AgreementAccepted: true,
	})
	if err != nil {
		t.Error("Failed to marshal update user body: " + err.Error())
	}
	updateUserReq, err := http.NewRequest("PUT", "/user", bytes.NewBuffer(updateUserBody))
	if err != nil {
		t.Error("Failed to create update user request: " + err.Error())
	}
	updateUserReq.Header.Set("Authorization", users.TestSessionToken)

	updateUserTest := func(t *testing.T, bdy []byte) bool {
		type updateUserResp struct {
			User users.User `json:"user"`
		}
		var resp updateUserResp
		err := json.Unmarshal(bdy, &resp)
		if err != nil {
			t.Error("Failed to unmarshal user response: " + err.Error())
			return false
		}
		if !resp.User.AgreementAccepted {
			t.Error("User agreement not marked as accepted")
			return false
		}

		user, err := users.Get(resp.User.ID, env.DB)
		if err != nil {
			t.Error("Failed to get user: " + err.Error())
			return false
		}
		t.Log("User agreement accepted:", user.AgreementAccepted)
		if !user.AgreementAccepted {
			t.Error("User agreement not accepted")
			return false
		}
		return true
	}

	getUserReq, err := http.NewRequest("GET", "/user", nil)
	if err != nil {
		t.Error("Failed to create get user request: " + err.Error())
	}
	getUserReq.Header.Set("Authorization", users.TestSessionToken)

	newResponseBody, err := json.Marshal(responses.Response{
		ElementID: 1,
		Value:     "test",
	})
	if err != nil {
		t.Error("Failed to marshal response body: " + err.Error())
	}
	newResponseReq, err := http.NewRequest("POST", "/response", bytes.NewBuffer(newResponseBody))
	if err != nil {
		t.Error("Failed to create new response request: " + err.Error())
	}
	newResponseReq.Header.Set("Authorization", users.TestSessionToken)
	type newResponseResp struct {
		Response responses.Response `json:"response"`
	}
	newResponseTest := func(t *testing.T, bdy []byte) bool {
		var resp newResponseResp
		err := json.Unmarshal(bdy, &resp)
		if err != nil {
			t.Error("Failed to unmarshal response body: " + err.Error())
			return false
		}
		if resp.Response.ID == 0 {
			t.Error("Response ID is 0")
			return false
		}
		if resp.Response.ElementID != 1 {
			t.Error(fmt.Sprintf("Response element ID is %d, expected 1", resp.Response.ElementID))
			return false
		}
		if resp.Response.Value != "test" {
			t.Error(fmt.Sprintf("Response value is %s, expected test", resp.Response.Value))
			return false
		}
		return true
	}

	newResponseWithOptionsBody, err := json.Marshal(responses.Response{
		ElementID: 2,
		OptionIDs: []int64{2, 3, 4},
	})
	if err != nil {
		t.Error("Failed to marshal response with options body: " + err.Error())
		return
	}
	newResponseWithOptionsReq, err := http.NewRequest("POST", "/response", bytes.NewBuffer(newResponseWithOptionsBody))
	if err != nil {
		t.Error("Failed to create new response with options request: " + err.Error())
		return
	}
	newResponseWithOptionsReq.Header.Set("Authorization", users.TestSessionToken)
	newResponseWithOptionsTest := func(t *testing.T, bdy []byte) bool {
		var resp newResponseResp
		err := json.Unmarshal(bdy, &resp)
		if err != nil {
			t.Error("Failed to unmarshal response body: " + err.Error())
			return false
		}
		var found []int64
		for _, opt := range resp.Response.OptionIDs {
			if opt != 2 && opt != 3 && opt != 4 {
				t.Error(fmt.Sprintf("Response option ID is %d, expected 2, 3, or 4", opt))
				return false
			}
			found = append(found, opt)
		}
		if len(found) != 3 {
			t.Error("Response option IDs are incorrect. Got:", found)
			return false
		}
		return true
	}

	testUser, err := users.GetUserBySession(users.TestSessionToken, env)
	if err != nil {
		t.Error("Failed to get user ID: " + err.Error())
		return
	}

	// get a response ID with a non-zero createdAt timestamp
	var responseID int64
	err = env.DB.QueryRow("SELECT id FROM responses WHERE createdAt > 0 AND userID = ?", testUser.ID).Scan(&responseID)
	if err != nil {
		t.Error("Failed to get response ID: " + err.Error())
		return
	}
	getResponseReq, err := http.NewRequest("GET", fmt.Sprintf("/response/%d", responseID), nil)
	if err != nil {
		t.Error("Failed to create get response request: " + err.Error())
	}
	getResponseReq.Header.Set("Authorization", users.TestSessionToken)
	getResponseTest := func(t *testing.T, bdy []byte) bool {
		type getResponseResp struct {
			Response responses.Response `json:"response"`
		}
		var resp getResponseResp
		err := json.Unmarshal(bdy, &resp)
		if err != nil {
			t.Error("Failed to unmarshal response body: " + err.Error())
			return false
		}
		if resp.Response.ID != responseID {
			t.Error("Response ID is incorrect")
			return false
		}
		if resp.Response.ElementID == 0 {
			t.Error("Response element ID is not set")
			return false
		}
		if resp.Response.CreatedAt.IsZero() {
			t.Error("Response created at timestamp is not set")
			return false
		}
		return true
	}

	// test a response with a user that doesn't own the response
	var userResponseID int64
	selectResponse := "SELECT id FROM responses WHERE userID != ? AND createdAt > 0"
	err = env.DB.QueryRow(selectResponse, testUser.ID).Scan(&userResponseID)
	if err != nil {
		t.Error("Failed to get response ID: " + err.Error())
		return
	}
	// test for unauthorized response when user does not own the form response
	getResponseWithIncorrectUserReq, err := http.NewRequest("GET", fmt.Sprintf("/response/%d", userResponseID), nil)
	if err != nil {
		t.Error("Failed to create get response request: " + err.Error())
	}
	getResponseWithIncorrectUserReq.Header.Set("Authorization", users.TestSessionToken)

	getFormResponsesReq, err := http.NewRequest("GET", "/forms/responses", nil)
	if err != nil {
		t.Error("Failed to create get form responses request: " + err.Error())
	}
	getFormResponsesReq.Header.Set("Authorization", users.TestSessionToken)
	getFormResponsesTest := func(t *testing.T, bdy []byte) bool {
		type getFormResponsesResp struct {
			Responses []responses.FormResponse `json:"form_responses"`
		}
		var resp getFormResponsesResp
		err := json.Unmarshal(bdy, &resp)
		if err != nil {
			t.Error("Failed to unmarshal response body: " + err.Error())
			return false
		}
		if len(resp.Responses) == 0 {
			t.Error("No responses found")
			return false
		}
		return true
	}

	testCases := []struct {
		name     string
		request  *http.Request
		wantCode int
		testBody func(*testing.T, []byte) bool
	}{
		{name: "ProviderLogin", request: providerLoginReq, wantCode: http.StatusOK},
		{name: "AuthenticateToken", request: authTokenReq, wantCode: http.StatusOK, testBody: authTokenBodyTest},
		{name: "GetForms", request: getFormsReq, wantCode: http.StatusOK, testBody: getFormsBodyTest},
		{name: "GetForm", request: getFormReq, wantCode: http.StatusOK, testBody: getFormBodyTest},
		{name: "UpdateUser", request: updateUserReq, wantCode: http.StatusOK, testBody: updateUserTest},
		{name: "GetUser", request: getUserReq, wantCode: http.StatusOK},
		{name: "NewResponse", request: newResponseReq, wantCode: http.StatusOK, testBody: newResponseTest},
		{name: "NewResponseWithOptions", request: newResponseWithOptionsReq, wantCode: http.StatusOK, testBody: newResponseWithOptionsTest},
		{name: "GetResponse", request: getResponseReq, wantCode: http.StatusOK, testBody: getResponseTest},
		{name: "GetResponseWithIncorrectUser", request: getResponseWithIncorrectUserReq, wantCode: http.StatusUnauthorized},
		{name: "GetFormResponses", request: getFormResponsesReq, wantCode: http.StatusOK, testBody: getFormResponsesTest},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			env.Router.ServeHTTP(w, tc.request)
			if w.Code != tc.wantCode {
				t.Errorf("got status code %d; want %d", w.Code, tc.wantCode)
			}
			if tc.testBody != nil {
				test := tc.testBody(t, w.Body.Bytes())
				if test {
					t.Log("body test passed")
					return
				}
				t.Error("body test failed")
			}
		})
	}
}
