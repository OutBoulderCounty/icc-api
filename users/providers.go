package users

import (
	"database/sql"
	"errors"
)

func ApproveProvider(userID int64, approved bool, db *sql.DB) error {
	_, err := db.Exec("update users set approvedProvider = ? where id = ?", approved, userID)
	if err != nil {
		return errors.New("error updating user. " + err.Error())
	}
	return nil
}
