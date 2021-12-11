package env

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/stytchauth/stytch-go/v3/stytch"
	"github.com/stytchauth/stytch-go/v3/stytch/stytchapi"
)

type SqlConnection struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type Env struct {
	Name   envName
	DB     *sql.DB
	Stytch *stytchapi.API
	Router *gin.Engine
}

type envName string

const (
	EnvDev  envName = "dev"
	EnvTest envName = "test"
	EnvProd envName = "prod"
)

func (env Env) SqlExecute(query string) (sql.Result, error) {
	statement, err := env.DB.Prepare(query)
	if err != nil {
		fmt.Println("Failed to prepare SQL " + query)
		return nil, err
	}
	result, err := statement.Exec()
	if err != nil {
		fmt.Println("Failed to execute SQL " + query)
		return nil, err
	}
	return result, nil
}

func Connect(name envName) (*Env, error) {
	fmt.Println("app env:", name)

	env := Env{
		Name: name,
	}

	var db *sql.DB
	var err error
	if env.Name == EnvProd {
		fmt.Println("Connecting to prod database")

		// get connection data from SSM
		sess := session.Must(session.NewSession())
		svc := ssm.New(sess, aws.NewConfig().WithRegion("us-west-2"))
		path := fmt.Sprintf("/icc/%s/database/", env.Name)
		decrypt := true
		input := ssm.GetParametersByPathInput{
			Path:           &path,
			WithDecryption: &decrypt,
		}
		out, err := svc.GetParametersByPath(&input)
		if err != nil {
			return nil, err
		}
		params := out.Parameters
		var c SqlConnection
		for i := 0; i < len(params); i++ {
			name := *params[i].Name
			value := *params[i].Value
			switch {
			case strings.HasSuffix(name, "host"):
				c.Host = value
			case strings.HasSuffix(name, "port"):
				c.Port = value
			case strings.HasSuffix(name, "user"):
				c.User = value
			case strings.HasSuffix(name, "password"):
				c.Password = value
			case strings.HasSuffix(name, "name"):
				c.Name = value
			}
		}

		db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=true&parseTime=true", c.User, c.Password, c.Host, c.Port, c.Name))
		if err != nil {
			return nil, err
		}
	} else {
		fmt.Println("Connecting to dev database")
		db, err = sql.Open("mysql", "tcp(localhost:3306)/?parseTime=true")
		if err != nil {
			return nil, err
		}

	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(time.Minute * 4)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	env.DB = db

	env.Stytch = env.initStytch()
	if env.Name != EnvTest {
		env.Router = gin.Default()
	}

	return &env, nil
}

func (env Env) initStytch() *stytchapi.API {
	stytchProjectID := os.Getenv("STYTCH_PROJECT_ID")
	stytchSecret := os.Getenv("STYTCH_SECRET")

	var stytchClient *stytchapi.API
	if env.Name == EnvProd {
		stytchClient = stytchapi.NewAPIClient(stytch.EnvLive, stytchProjectID, stytchSecret)
	} else {
		stytchClient = stytchapi.NewAPIClient(stytch.EnvTest, stytchProjectID, stytchSecret)
	}
	return stytchClient
}

func TestSetup(t *testing.T, parallel bool, pathToDotEnv string) *Env {
	if parallel {
		t.Parallel()
	}
	err := godotenv.Load(pathToDotEnv)
	if err != nil {
		t.Error("Failed to load ../.env. " + err.Error())
	}
	e, err := Connect(EnvTest)
	if err != nil {
		t.Error("Failed to connect services. " + err.Error())
	}
	return e
}
