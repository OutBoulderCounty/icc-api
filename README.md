# Inclusive Care CO API

## Prerequisites

- [Install Go](https://golang.org/doc/install)
- [Install air](https://github.com/cosmtrek/air#installation)
- [Install Planetscale CLI](https://docs.planetscale.com/reference/planetscale-environment-setup) and login

  ```sh
  pscale login
  ```

- Set up Stytch environment variables in .env file at the root of this repository

  ```env
  STYTCH_PROJECT_ID="<project id>"
  STYTCH_SECRET="<secret>"
  ```

- Clone this repo

## How to run

1. Open a connection to the dev branch of the database

   ```sh
   pscale connect icc-dev dev
   ```

   This will create a connection to the database `icc-dev` on branch `dev`

1. In a separate terminal, `cd` into this repo and run `air`. This will provide a hot reloading environment

   ```sh
   cd icc-api
   air
   ```

   When air is done building, you will have a server running on port 8080

## Logging in

I use [httpie](https://httpie.io/cli) to make requests in the examples below, but these could be translated to curl or any other tool.

1. To get a session token, make a POST request to the `/login` endpoint with your email address in the body

   ```sh
   http POST http://localhost:8080/login email=<email>
   ```

   You will receive an email to the address you provided. Click the link in the email to get a session token. Copy the session token value from the JSON response.

1. With a session token in hand, you can now make authenticated requests

   ```sh
   http GET http://localhost:8080/forms Authorization:"<session token>"
   ```

   Every time you use your session token, it will be renewed for an additional 7 days. If you do not use your session token for more than 7 days, you will need to login again.
