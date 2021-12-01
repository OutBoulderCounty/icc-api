package users_test

import (
	"api/env"
	"api/users"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestLogin(t *testing.T) {
	t.Parallel()

	err := godotenv.Load("../.env")
	if err != nil {
		t.Error("Failed to load .env. " + err.Error())
	}

	// login a user
	email := os.Getenv("TEST_USER")
	redirectURL := "http://localhost:8080/localauth"
	e, err := env.Connect(env.EnvTest)
	if err != nil {
		t.Error("Failed to connect services. " + err.Error())
	}
	user := users.UserReq{
		Email:       email,
		RedirectURL: redirectURL,
	}
	err = users.Login(user, e)
	if err != nil {
		t.Error("Login failed. " + err.Error())
	}
}

// take in a user and update it
func TestUpdateUser(t *testing.T) {
	t.Parallel()

	// login a user

	// update the user with additional data

	// check that the user has been updated
}
