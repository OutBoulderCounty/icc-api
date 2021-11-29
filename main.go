package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"api/database"
	"api/forms"
	"api/users"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stytchauth/stytch-go/v3/stytch"
	"github.com/stytchauth/stytch-go/v3/stytch/stytchapi"
)

func setup() (*gin.Engine, *database.DB) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowWildcard = true
	config.AllowOrigins = []string{"http://localhost:8000", "https://*.inclusivecareco.org", "http://localhost:3000", "http://localhost:3002"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	// TODO: switch to prod when ready
	db, err := database.Connect("dev")
	if err != nil {
		log.Fatal(err)
	}

	// TODO: switch to stytch.EnvLive when ready
	stytchClient := stytchapi.NewAPIClient(stytch.EnvTest, os.Getenv("STYTCH_PROJECT_ID"), os.Getenv("STYTCH_SECRET"))

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World!")
	})

	r.POST("/login", func(c *gin.Context) {
		users.Login(c, stytchClient, db)
	})

	r.POST("/authenticate", func(c *gin.Context) {
		users.AuthenticateUser(c, stytchClient)
	})

	// for testing locally without a UI
	r.GET("/localauth", func(c *gin.Context) {
		var login struct {
			Token string `form:"token"`
		}
		err := c.ShouldBindQuery(&login)
		if err != nil {
			fmt.Println("Failed to bind query: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		sessionToken, err := users.Authenticate(login.Token, stytchClient)
		if err != nil {
			fmt.Println("Failed to authenticate: " + err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}
		fmt.Println("Authenticated!")
		c.JSON(http.StatusOK, gin.H{
			"session_token": sessionToken,
		})
	})

	authorized := r.Group("/forms", authRequired(stytchClient, db))

	authorized.GET("", func(c *gin.Context) {
		foundForms, err := forms.GetForms(db.DB)
		if err != nil {
			panic(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"forms": foundForms,
		})
	})

	return r, db
}

func main() {
	r, db := setup()
	defer db.Close()
	r.Run()
}

func authRequired(stytchClient *stytchapi.API, db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}
		body := stytch.SessionsAuthenticateParams{
			SessionToken:           token,
			SessionDurationMinutes: 10080,
		}
		resp, err := stytchClient.Sessions.Authenticate(&body)
		if err != nil {
			fmt.Println("Failed to authorize: " + err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}
		fmt.Println("Authorized!", resp.Session.UserID)
		c.Set("stytch_user_id", resp.Session.UserID)
		c.Next()
	}
}
