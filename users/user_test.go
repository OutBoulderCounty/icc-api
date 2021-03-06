package users_test

import (
	"api/env"
	"api/users"
	"testing"

	"github.com/joho/godotenv"
)

func setup(t *testing.T, parallel bool) *env.Env {
	if parallel {
		t.Parallel()
	}
	err := godotenv.Load("../.env")
	if err != nil {
		t.Error("Failed to load ../.env. " + err.Error())
	}
	e, err := env.Connect(env.EnvTest)
	if err != nil {
		t.Error("Failed to connect services. " + err.Error())
	}
	return e
}

func TestLogin(t *testing.T) {
	e := setup(t, true)
	user := users.UserReq{
		Email:       users.TestUser,
		RedirectURL: users.TestRedirectURL,
	}
	_, err := users.Login(user, e)
	if err != nil {
		t.Error("Login failed. " + err.Error())
	}
}

// take in a user and update it
func TestUpdateUser(t *testing.T) {
	e := setup(t, true)

	userReq := users.UserReq{
		Email:       users.TestUser,
		RedirectURL: users.TestRedirectURL,
	}
	// login a user
	_, err := users.Login(userReq, e)
	if err != nil {
		t.Error("Login failed. " + err.Error())
	}
	sessToken, err := users.Authenticate(users.TestToken, e)
	if err != nil {
		t.Error("Failed to authenticate user. " + err.Error())
	}
	if sessToken == "" {
		t.Error("Failed to authenticate user. No session token returned.")
	}

	// update the user with additional data
	user := users.User{
		Email:             userReq.Email,
		FirstName:         "Test",
		LastName:          "Provider",
		Pronouns:          "they/them",
		PracticeName:      "Test's Practice",
		Address:           "123 Test St",
		Specialty:         "",
		Phone:             "123-456-7890",
		AgreementAccepted: true,
	}
	user.ID, err = users.UpdateUser(sessToken, &user, e)
	if err != nil {
		t.Error("Failed to update user. " + err.Error())
	}

	// check that the user has been updated
	u, err := users.Get(user.ID, e.DB)
	if err != nil {
		t.Error("Failed to get user. " + err.Error())
	}
	if u.Email != user.Email {
		t.Error("User email not updated.")
	}
	if u.FirstName != user.FirstName {
		t.Error("User first name not updated.")
	}
	if u.LastName != user.LastName {
		t.Error("User last name not updated.")
	}
	if u.Pronouns != user.Pronouns {
		t.Error("User pronouns not updated.")
	}
	if u.PracticeName != user.PracticeName {
		t.Error("User practice name not updated.")
	}
	if u.Address != user.Address {
		t.Error("User address not updated.")
	}
	if u.Specialty != user.Specialty {
		t.Error("User specialty not updated.")
	}
	if u.Phone != user.Phone {
		t.Error("User phone not updated.")
	}
}

func TestGetUserBySession(t *testing.T) {
	e := setup(t, true)

	userReq := users.UserReq{
		Email:       users.TestUser,
		RedirectURL: users.TestRedirectURL,
	}
	// login a user
	_, err := users.Login(userReq, e)
	if err != nil {
		t.Error("Login failed. " + err.Error())
	}
	sessToken, err := users.Authenticate(users.TestToken, e)
	if err != nil {
		t.Error("Failed to authenticate user. " + err.Error())
	}
	if sessToken == "" {
		t.Error("Failed to authenticate user. No session token returned.")
	}

	// get the user by session token
	u, err := users.GetUserBySession(sessToken, e)
	if err != nil {
		t.Error("Failed to get user by session. " + err.Error())
	}
	if u.Email != userReq.Email {
		t.Error("User email not returned.")
	}
}
