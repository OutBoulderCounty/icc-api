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

1. To get a session token, make a POST request to the `/login` endpoint with your email address and the appropriate redirect URL in the body. Make sure the component at this URL makes a POST request to the /authenticate API endpoint with a body containing the token provided in the query parameters. The data you get back from this request will contain a session token. You can find an example implementation in the icc-admin-ui repo.

   ```sh
   http POST http://localhost:8080/login email=<email> redirect_url=<redirect url>
   ```

   You will receive an email to the email address you provided. Clicking on this link will redirect you to the URL you provided.

   To test locally with only the API running, you can pass `http://localhost:8080/localauth` as the redirect URL and get a session token that way.

1. With a session token in hand, you can now make authenticated requests

   ```sh
   http GET http://localhost:8080/forms Authorization:"<session token>"
   ```

   Every time you use your session token, it will be renewed for an additional 7 days. If you do not use your session token for more than 7 days, you will need to login again.

## Available Routes

| Path            | Method | Description                    |
| --------------- | ------ | ------------------------------ |
| `/login`        | POST   | Login endpoint                 |
| `/authenticate` | POST   | Authenticate endpoint          |
| `/localauth`    | GET    | Local authentication endpoint  |
| `/forms`        | GET    | Get all forms                  |
| `/user`         | PUT    | Update user                    |
| `/user`         | GET    | Get user data by session token |
