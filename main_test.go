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

	newResponseBody, err := json.Marshal(map[string]string{
		"element_id": "1",
		"value":      "test",
	})
	if err != nil {
		t.Error("Failed to marshal response body: " + err.Error())
	}
	newResponseReq, err := http.NewRequest("POST", "/response", bytes.NewBuffer(newResponseBody))
	if err != nil {
		t.Error("Failed to create new response request: " + err.Error())
	}
	newResponseReq.Header.Set("Authorization", users.TestSessionToken)
	newResponseTest := func(t *testing.T, bdy []byte) bool {
		type newResponseResp struct {
			Response responses.Response `json:"response"`
		}
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

	testCases := []struct {
		name     string
		request  *http.Request
		wantCode int
		wantBody *string
		testBody func(*testing.T, []byte) bool
	}{
		{name: "ProviderLogin", request: providerLoginReq, wantCode: http.StatusOK},
		{name: "AuthenticateToken", request: authTokenReq, wantCode: http.StatusOK, testBody: authTokenBodyTest},
		{name: "GetForms", request: getFormsReq, wantCode: http.StatusOK, testBody: getFormsBodyTest},
		{name: "GetForm", request: getFormReq, wantCode: http.StatusOK, testBody: getFormBodyTest},
		{name: "UpdateUser", request: updateUserReq, wantCode: http.StatusOK, testBody: updateUserTest},
		{name: "GetUser", request: getUserReq, wantCode: http.StatusOK},
		{name: "NewResponse", request: newResponseReq, wantCode: http.StatusOK, testBody: newResponseTest},
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
			if tc.wantBody == nil {
				return
			}
			if w.Body.String() != *tc.wantBody {
				t.Errorf("got body %s; want %s", w.Body.String(), *tc.wantBody)
			}
		})
	}
}
