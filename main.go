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
	"api/forms/tally"
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
	case "local":
		environment, err = env.Connect(env.EnvLocal)
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

	environment.Router.POST("/response/tally", func(c *gin.Context) {
		var event tally.Event
		err := c.ShouldBindJSON(&event)
		if err != nil {
			fmt.Println("Failed to bind JSON: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		_, err = event.SaveResponse(environment.DB)
		if err != nil {
			fmt.Println("Failed to save response: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.Status(http.StatusOK)
	})

	environment.Router.POST("/form/tally/register", func(c *gin.Context) {
		var event tally.Event
		err := c.ShouldBindJSON(&event)
		if err != nil {
			msg := "Failed to bind JSON: " + err.Error()
			fmt.Println(msg)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": msg,
			})
			return
		}
		_, err = event.RegisterForm(environment.DB)
		if err != nil {
			msg := "Failed to register form: " + err.Error()
			fmt.Println(msg)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": msg,
			})
			return
		}
		c.Status(http.StatusOK)
	})

	environment.Router.GET("/forms/tally", func(c *gin.Context) {
		forms, err := tally.GetForms(environment.DB)
		if err != nil {
			msg := "Failed to get forms: " + err.Error()
			fmt.Println(msg)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": msg,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"forms": forms,
		})
	})

	environment.Router.GET("/responses/tally/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			msg := "Failed to parse int64: " + err.Error()
			fmt.Println(msg)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": msg,
			})
			return
		}
		responses, err := tally.GetPrettyResponse(id, environment.DB)
		if err != nil {
			msg := "Failed to get responses: " + err.Error()
			fmt.Println(msg)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": msg,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"responses": responses,
		})
	})

	environment.Router.GET("/providers", func(c *gin.Context) {
		providers, err := users.GetApprovedProviders(environment.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"providers": providers,
		})
	})

	environment.Router.GET("/provider/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		provider, err := users.GetApprovedProvider(&id, environment.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"provider": provider,
		})
	})

	environment.Router.GET("/provider/:id/responses", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		resps, err := responses.GetApprovedResponsesByProvider(id, environment.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
		c.JSON(http.StatusOK, gin.H{
			"responses": resps,
		})
	})

	environment.Router.GET("/provider/:id/responses/all", adminAuthRequired(environment), func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		resps, err := responses.GetResponsesByProvider(id, environment.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
		c.JSON(http.StatusOK, gin.H{
			"responses": resps,
		})
	})

	authorizedUser := environment.Router.Group("/user", authRequired(environment))
	authorizedUser.PUT("", func(c *gin.Context) {
		users.UpdateUserHandler(c, environment)
	})
	authorizedUser.GET("", func(c *gin.Context) {
		users.GetUserHandler(c, environment)
	})
	authorizedUser.GET("/:id", adminAuthRequired(environment), func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		user, err := users.Get(id, environment.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"user": user,
		})
	})
	authorizedUser.PUT("/agreement/:bool", func(c *gin.Context) {
		id := c.GetInt64("user_id")
		agreement, err := strconv.ParseBool(c.Param("bool"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		err = users.UpdateAgreement(&id, &agreement, environment)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
		})
	})

	adminUsers := environment.Router.Group("/users", adminAuthRequired(environment))
	adminUsers.GET("", func(c *gin.Context) {
		foundUsers, err := users.GetUsers(environment.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"users": foundUsers,
		})
	})

	unauthorizedForms := environment.Router.Group("/forms")
	unauthorizedForms.GET("", func(c *gin.Context) {
		foundForms, err := forms.GetLiveForms(environment.DB)
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
	unauthorizedForms.GET("/all", adminAuthRequired(environment), func(c *gin.Context) {
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
	unauthorizedForms.GET("/responses", func(c *gin.Context) {
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

	form := environment.Router.Group("/form")
	form.GET("/:id", func(c *gin.Context) {
		forms.GetFormHandler(c, true, environment.DB)
	})
	form.GET("/any/:id", adminAuthRequired(environment), func(c *gin.Context) {
		forms.GetFormHandler(c, false, environment.DB)
	})
	form.GET("/:id/responses", authRequired(environment), func(c *gin.Context) {
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
	form.GET("/:id/responses/all", adminAuthRequired(environment), func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		resps, err := responses.GetResponsesByForm(id, environment.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"responses": resps})
	})
	form.POST("", adminAuthRequired(environment), func(c *gin.Context) {
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
	form.PUT("", adminAuthRequired(environment), func(c *gin.Context) {
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
	form.DELETE("/:id", adminAuthRequired(environment), func(c *gin.Context) {
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
	authorizedResponse.PUT("/:id/approve/:approval", adminAuthRequired(environment), func(c *gin.Context) {
		approval, err := strconv.ParseBool(c.Param("approval"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		err = responses.ApproveResponse(id, approval, environment.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.Status(http.StatusOK)
	})
	authorizedResponse.GET("/any/:id", adminAuthRequired(environment), func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}
		resp, err := responses.GetResponse(id, environment.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"response": resp})
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
	authorizedResponses.GET("/all", adminAuthRequired(environment), func(c *gin.Context) {
		resps, err := responses.GetResponses(environment.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"responses": resps})
	})

	provider := environment.Router.Group("/provider")
	provider.PUT("/:id/approve/:approval", adminAuthRequired(environment), func(c *gin.Context) {
		approval, err := strconv.ParseBool(c.Param("approval"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		err = users.ApproveProvider(id, approval, environment.DB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.Status(http.StatusOK)
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
