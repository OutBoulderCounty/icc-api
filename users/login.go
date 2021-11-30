package users

import (
	"database/sql"
	"fmt"
	"net/http"

	"api/env"

	"github.com/gin-gonic/gin"
	"github.com/stytchauth/stytch-go/v3/stytch"
)

type UserReq struct {
	Email       string   `json:"email"`
	RedirectURL string   `json:"redirect_url"` // must be defined in Stytch as a redirect URL
	Roles       []string `json:"roles"`
}

type Role struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Protected bool   `json:"protected"`
}

type UserRole struct {
	ID     int64 `json:"id"`
	UserID int64 `json:"userID"`
	RoleID int64 `json:"roleID"`
	Active bool  `json:"active"`
}

func Login(c *gin.Context, e *env.Env) error {
	var user UserReq
	err := c.BindJSON(&user)
	if err != nil {
		fmt.Println("Failed to bind JSON: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return err
	}

	body := stytch.MagicLinksEmailLoginOrCreateParams{
		Email:              user.Email,
		LoginMagicLinkURL:  user.RedirectURL,
		SignupMagicLinkURL: user.RedirectURL,
	}
	resp, err := e.Stytch.MagicLinks.Email.LoginOrCreate(&body)
	if err != nil {
		fmt.Println("Failed to create magic link: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return err
	}

	row := e.DB.QueryRow("SELECT id FROM users WHERE email = ? AND stytchUserID = ?", user.Email, resp.UserID)
	var userID int64
	err = row.Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			result, err := e.SqlExecute(fmt.Sprintf("INSERT INTO users (stytchUserID, email) VALUES ('%s', '%s')", resp.UserID, user.Email))
			if err != nil {
				fmt.Println("Failed to create user: " + err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return err
			}
			userID, err = result.LastInsertId()
			if err != nil {
				fmt.Println("Failed to get user ID: " + err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return err
			}
		} else {
			fmt.Println("Failed to query user: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return err
		}
	}

	roles, err := user.validateRoles(e.DB)
	if err != nil {
		fmt.Println("Failed to validate roles: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return err
	}
	for i := 0; i < len(roles); i++ {
		created, err := roles[i].addUserToRole(e)
		if err != nil {
			fmt.Println("Failed to add user to role: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return err
		}
		if created {
			row := e.DB.QueryRow("SELECT name FROM roles WHERE id = ?", roles[i].RoleID)
			var name string
			err := row.Scan(&name)
			if err != nil {
				fmt.Println("Failed to query role: " + err.Error())
			}
			fmt.Printf("Added user %s to role %s\n", user.Email, name)
			// TODO: send notification to slack or email
		}
	}

	var status int
	if resp.UserCreated {
		fmt.Println("User created")
		status = http.StatusCreated
	} else {
		status = http.StatusOK
	}
	c.JSON(status, gin.H{
		"id":             userID,
		"stytch_user_id": resp.UserID,
	})
	return nil
}

func (u UserReq) validateRoles(db *sql.DB) (userRoles []UserRole, err error) {
	var validRoles []Role
	rows, err := db.Query("SELECT id, name, protected FROM roles")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var role Role
		err = rows.Scan(&role.ID, &role.Name, &role.Protected)
		if err != nil {
			return nil, err
		}
		validRoles = append(validRoles, role)
	}

	if err != nil {
		fmt.Println("Failed to query roles: " + err.Error())
		return nil, err
	}
	for _, role := range u.Roles {
		for _, validRole := range validRoles {
			if role == validRole.Name {
				userRole := UserRole{
					RoleID: validRole.ID,
					Active: !validRole.Protected,
				}
				userRoles = append(userRoles, userRole)
			}
		}
	}
	return userRoles, nil
}

func (ur UserRole) addUserToRole(e *env.Env) (created bool, err error) {
	created = false
	// check if user is already in role
	rows, err := e.DB.Query("SELECT id FROM user_roles WHERE userID = ? AND roleID = ?", ur.UserID, ur.RoleID)
	if err != nil {
		fmt.Println("Failed to query user role: " + err.Error())
		return created, err
	}
	defer rows.Close()
	if rows.Next() {
		return created, nil
	}
	_, err = e.SqlExecute(fmt.Sprintf("INSERT INTO user_roles (userID, roleID, active) VALUES (%v, %v, %v)", ur.UserID, ur.RoleID, ur.Active))
	if err != nil {
		fmt.Println("Failed to create user role: " + err.Error())
		return created, err
	}
	created = true
	return created, nil
}
