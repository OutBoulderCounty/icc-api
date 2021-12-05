package users

import (
	"api/env"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/stytchauth/stytch-go/v3/stytch"
)

const TestUser = "sandbox@stytch.com"
const TestRedirectURL = "http://localhost:8080/localauth"
const TestToken = "DOYoip3rvIMMW5lgItikFK-Ak1CfMsgjuiCyI7uuU94="
const TestSessionToken = "WJtR5BCy38Szd5AfoDpf0iqFKEt4EE5JhjlWUY7l3FtY"

type User struct {
	ID                int64
	StytchUserID      string
	Email             string
	ActiveRoles       []string
	FirstName         string
	LastName          string
	Pronouns          string
	PracticeName      string
	Address           string
	Specialty         string
	Phone             string
	AgreementAccepted bool
}

type sqlUser struct {
	User
	FirstName    sql.NullString
	LastName     sql.NullString
	Pronouns     sql.NullString
	PracticeName sql.NullString
	Address      sql.NullString
	Specialty    sql.NullString
	Phone        sql.NullString
}

func (u *sqlUser) ToUser() *User {
	var user User
	user.ID = u.ID
	user.StytchUserID = u.StytchUserID
	user.Email = u.Email
	user.FirstName = u.FirstName.String
	user.LastName = u.LastName.String
	user.Pronouns = u.Pronouns.String
	user.PracticeName = u.PracticeName.String
	user.Address = u.Address.String
	user.Specialty = u.Specialty.String
	user.Phone = u.Phone.String
	user.AgreementAccepted = u.AgreementAccepted
	return &user
}

// retrieves a single user from the database
func Get(id int64, db *sql.DB) (*User, error) {
	row := db.QueryRow("SELECT id, stytchUserID, email, firstName, lastName, pronouns, practiceName, address, specialty, phone, agreementAccepted FROM users WHERE id = ?", id)
	var dbUser sqlUser
	err := row.Scan(
		&dbUser.ID,
		&dbUser.StytchUserID,
		&dbUser.Email,
		&dbUser.FirstName,
		&dbUser.LastName,
		&dbUser.Pronouns,
		&dbUser.PracticeName,
		&dbUser.Address,
		&dbUser.Specialty,
		&dbUser.Phone,
		&dbUser.AgreementAccepted,
	)
	if err != nil {
		return nil, err
	}
	user := dbUser.ToUser()
	return user, nil
}

func GetUserByStytchID(stytchUserID *string, e *env.Env) (*User, error) {
	if stytchUserID == nil {
		return nil, errors.New("stytchUserID is required")
	}
	row := e.DB.QueryRow("SELECT id, stytchUserID, email, firstName, lastName, pronouns, practiceName, address, specialty, phone, agreementAccepted FROM users WHERE stytchUserID = ?", *stytchUserID)
	var dbUser sqlUser
	err := row.Scan(
		&dbUser.ID,
		&dbUser.StytchUserID,
		&dbUser.Email,
		&dbUser.FirstName,
		&dbUser.LastName,
		&dbUser.Pronouns,
		&dbUser.PracticeName,
		&dbUser.Address,
		&dbUser.Specialty,
		&dbUser.Phone,
		&dbUser.AgreementAccepted,
	)
	if err != nil {
		return nil, err
	}
	user := dbUser.ToUser()
	return user, nil
}

func GetUserBySession(sessionToken string, e *env.Env) (*User, error) {
	if sessionToken == "" {
		return nil, errors.New("session token is required")
	}
	// get user id from session token
	params := &stytch.SessionsAuthenticateParams{
		SessionToken: sessionToken,
	}
	resp, err := e.Stytch.Sessions.Authenticate(params)
	if err != nil {
		return nil, errors.New("failed to authenticate session: " + err.Error())
	}
	user, err := GetUserByStytchID(&resp.Session.UserID, e)
	if err != nil {
		return nil, errors.New("failed to get user from DB: " + err.Error())
	}
	return user, nil
}

func GetUserHandler(c *gin.Context, e *env.Env) {
	user, err := GetUserBySession(c.GetHeader("Authorization"), e)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func UpdateUser(sessionToken string, user *User, e *env.Env) (int64, error) {
	if sessionToken == "" {
		return 0, errors.New("session token is required")
	}
	// get user id from session token
	if user.StytchUserID == "" {
		params := &stytch.SessionsAuthenticateParams{
			SessionToken: sessionToken,
		}
		resp, err := e.Stytch.Sessions.Authenticate(params)
		if err != nil {
			return 0, errors.New("failed to authenticate session: " + err.Error())
		}
		user.StytchUserID = resp.Session.UserID
	}
	// get existing user from db
	existingUser, err := GetUserByStytchID(&user.StytchUserID, e)
	if err != nil {
		return 0, errors.New("failed to get existing user from DB: " + err.Error())
	}
	if user.Email == "" {
		user.Email = existingUser.Email
	}

	_, err = e.SqlExecute(fmt.Sprintf(
		"UPDATE users SET email = '%s', firstName = '%s', lastName = '%s', pronouns = '%s', practiceName = '%s', address = '%s', specialty = '%s', phone = '%s', agreementAccepted = %v WHERE stytchUserID = '%s'",
		user.Email,
		user.FirstName,
		user.LastName,
		user.Pronouns,
		user.PracticeName,
		user.Address,
		user.Specialty,
		user.Phone,
		user.AgreementAccepted,
		user.StytchUserID,
	))
	if err != nil {
		return 0, errors.New("failed to update user: " + err.Error())
	}

	return existingUser.ID, nil
}

func UpdateUserHandler(c *gin.Context, e *env.Env) {
	var user User
	err := c.BindJSON(&user)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	user.ID, err = UpdateUser(c.GetHeader("Authorization"), &user, e)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"user": user})
}

func DeleteUser(stytchUserID *string, e *env.Env) error {
	if stytchUserID == nil {
		return errors.New("stytchUserID is required")
	}
	_, err := e.Stytch.Users.Delete(*stytchUserID)
	if err != nil {
		if strings.Contains(err.Error(), "status code: 404") {
			fmt.Println("Stytch user not found")
		} else {
			return errors.New("failed to delete user from Stytch: " + err.Error())
		}
	}
	_, err = e.SqlExecute(fmt.Sprintf("DELETE FROM users WHERE stytchUserID = '%s'", *stytchUserID))
	if err != nil {
		return errors.New("failed to delete user from DB: " + err.Error())
	}
	return nil
}
