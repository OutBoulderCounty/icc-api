package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	_ "github.com/go-sql-driver/mysql"
)

type SqlConnection struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type DB struct {
	*sql.DB
}

func (db DB) Execute(query string) (sql.Result, error) {
	statement, err := db.Prepare(query)
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

func Connect(env string) (*DB, error) {
	// get connection data from SSM
	fmt.Println("app env:", env)

	var db *sql.DB
	var err error
	if env == "prod" {
		fmt.Println("Connecting to prod database")

		sess := session.Must(session.NewSession())
		svc := ssm.New(sess, aws.NewConfig().WithRegion("us-west-2"))
		path := fmt.Sprintf("/icc/%s/database/", env)
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

		db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=true", c.User, c.Password, c.Host, c.Port, c.Name))
		if err != nil {
			return nil, err
		}
	} else {
		fmt.Println("Connecting to dev database")
		db, err = sql.Open("mysql", "tcp(localhost:3306)/")
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

	database := DB{db}
	return &database, nil
}
