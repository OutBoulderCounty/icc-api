package main

import (
	"api/forms"
	"api/users"
	"bytes"
	"encoding/json"
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

	updateUserBody, err := json.Marshal(users.User{
		FirstName: "Integration",
		LastName:  "Test",
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
		t.Log("Update user response body: " + string(bdy))
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
		{name: "UpdateUser", request: updateUserReq, wantCode: http.StatusOK, testBody: updateUserTest},
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
