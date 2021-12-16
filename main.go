// Inclusive Care CO REST API
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"api/env"
	"api/forms"
	"api/forms/responses"
	"api/users"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func setup() *env.Env {
	godotenv.Load()
	appEnv := os.Getenv("APP_ENV")
	var environment *env.Env
	var err error
	switch appEnv {
	case "prod":
		environment, err = env.Connect(env.EnvProd)
	case "test":
		environment, err = env.Connect(env.EnvTest)
	case "dev":
		environment, err = env.Connect(env.EnvDev)
	default:
		log.Fatal("Invalid APP_ENV")
	}
	if err != nil {
		log.Fatal("Failed to connect services: " + err.Error())
	}

	config := cors.DefaultConfig()
	config.AllowWildcard = true
	config.AllowOrigins = []string{"http://localhost:8000", "https://*.inclusivecareco.org", "http://localhost:3000", "http://localhost:3002", "https://icc-provider-ui.vercel.app"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	environment.Router.Use(cors.New(config))

	environment.Router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World!")
	})

	environment.Router.POST("/login", func(c *gin.Context) {
		users.LoginHandler(c, environment)
	})

	environment.Router.POST("/authenticate", func(c *gin.Context) {
		users.AuthenticateUser(c, environment)
	})

	// for testing locally without a UI
	environment.Router.GET("/localauth", func(c *gin.Context) {
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
		sessionToken, err := users.Authenticate(login.Token, environment)
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

	authorizedUser := environment.Router.Group("/user", authRequired(environment))

	authorizedUser.PUT("", func(c *gin.Context) {
		users.UpdateUserHandler(c, environment)
	})

	authorizedUser.GET("", func(c *gin.Context) {
		users.GetUserHandler(c, environment)
	})

	authorizedForms := environment.Router.Group("/forms", authRequired(environment))
	authorizedForms.GET("", func(c *gin.Context) {
		foundForms, err := forms.GetForms(environment.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"forms": foundForms,
		})
	})
	authorizedForms.GET("/responses", func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		formResps, err := responses.GetFormResponsesByToken(token, environment)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"form_responses": formResps,
		})
	})

	authorizedForm := environment.Router.Group("/form", authRequired(environment))
	authorizedForm.GET("/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		form, err := forms.GetForm(id, environment.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"form": form})
	})
	authorizedForm.GET("/:id/responses", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		token := c.Request.Header.Get("Authorization")
		resps, err := responses.GetResponsesByFormAndToken(id, token, environment)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"responses": resps,
		})
	})
	authorizedForm.POST("", adminAuthRequired(environment), func(c *gin.Context) {
		var form forms.Form
		err := c.ShouldBindJSON(&form)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		newForm, err := forms.NewForm(&form, environment.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"form": newForm,
		})
	})
	authorizedForm.PUT("", adminAuthRequired(environment), func(c *gin.Context) {
		var form forms.Form
		err := c.ShouldBindJSON(&form)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		err = forms.UpdateForm(&form, environment.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.Status(http.StatusOK)
	})
	authorizedForm.DELETE("/:id", adminAuthRequired(environment), func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		err = forms.DeleteForm(id, environment.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	})

	authorizedResponse := environment.Router.Group("/response", authRequired(environment))
	authorizedResponse.POST("", func(c *gin.Context) {
		var response responses.Response
		err := c.ShouldBindJSON(&response)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		// get user ID from session token
		user, err := users.GetUserBySession(c.Request.Header.Get("Authorization"), environment)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		var resp *responses.Response
		// NOTE: potential problem here because someone could pass both option IDs and a value.
		// If Option IDs are passed, any value passed will not be stored.
		if response.OptionIDs != nil {
			resp, err = responses.NewResponseWithOptions(response.ElementID, user.ID, response.OptionIDs, environment.DB)
		} else {
			resp, err = responses.NewResponse(response.ElementID, user.ID, response.Value, environment.DB)
		}
		if err != nil {
			if err.Error() == "user must accept the user agreement" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"response": resp})
	})
	authorizedResponse.GET("/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		response, err := responses.GetResponse(id, environment.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		// check user owns the response
		userID, userIDExists := c.Get("user_id")
		if userIDExists {
			if userID != response.UserID {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "User does not own response",
				})
				return
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unable to get user ID from context",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"response": response})
	})

	authorizedResponses := environment.Router.Group("/responses", authRequired(environment))
	authorizedResponses.GET("", func(c *gin.Context) {
		resps, err := responses.GetResponses(environment.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		userID, userIDExists := c.Get("user_id")
		var userResps []*responses.Response
		if userIDExists {
			for _, resp := range resps {
				if resp.UserID == userID {
					userResps = append(userResps, resp)
				}
			}
		}
		c.JSON(http.StatusOK, gin.H{"responses": userResps})
	})

	return environment
}

func main() {
	env := setup()
	defer env.DB.Close()
	env.Router.Run()
}

func authRequired(environment *env.Env) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}
		user, err := users.GetUserBySession(token, environment)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}
		c.Set("user_id", user.ID)
		c.Set("stytch_user_id", user.StytchUserID)
		c.Next()
	}
}

func adminAuthRequired(environment *env.Env) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: no token provided",
			})
			c.Abort()
			return
		}
		user, err := users.GetUserBySession(token, environment)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}
		for _, role := range user.ActiveRoles {
			if role == "admin" {
				c.Set("user_id", user.ID)
				c.Set("stytch_user_id", user.StytchUserID)
				c.Next()
				return
			}
		}
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User is not an admin",
		})
		c.Abort()
	}
}
