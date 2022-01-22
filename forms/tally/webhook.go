package tally

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"api/users"
)

type Event struct {
	EventID   string    `json:"eventId"`
	CreatedAt string    `json:"createdAt"`
	Data      EventData `json:"data"`
}

func (e *Event) SaveResponse(db *sql.DB) (*Response, error) {
	fields := e.Data.Fields
	if len(fields) == 0 {
		return nil, errors.New("no fields in event data")
	}
	var userID int64
	var err error
	for i := 0; i < len(fields); i++ {
		if fields[i].Label == "user_id" {
			stringID := fields[i].Value.String
			userID, err = strconv.ParseInt(stringID, 10, 64)
			if err != nil {
				return nil, errors.New("error parsing user ID. " + err.Error())
			}
			break
		}
	}
	if userID == 0 {
		return nil, errors.New("no user ID in event data")
	}
	user, err := users.Get(userID, db)
	if err != nil {
		return nil, errors.New("error getting user. " + err.Error())
	}
	createdAt, err := time.Parse(time.RFC3339Nano, e.CreatedAt)
	if err != nil {
		return nil, errors.New("error parsing created at. " + err.Error())
	}

	response := Response{
		FormID:    e.Data.FormID,
		FormName:  e.Data.FormName,
		EventID:   e.EventID,
		CreatedAt: createdAt,
		User:      user,
		Fields:    fields,
	}
	err = response.Save(db)
	if err != nil {
		return nil, errors.New("error saving response. " + err.Error())
	}
	return &response, nil
}

type Response struct {
	ID        int64       `json:"id"`
	EventID   string      `json:"event_id"`
	FormID    string      `json:"form_id"`
	FormName  string      `json:"form_name"`
	CreatedAt time.Time   `json:"created_at"`
	User      *users.User `json:"user"`
	Fields    []Field     `json:"fields"`
}

func (r *Response) Save(db *sql.DB) error {
	if r.ID != 0 {
		return errors.New("response already saved")
	}
	query := "insert into tally_responses (event_id, form_id, form_name, created_at, user_id, fields) values (?, ?, ?, ?, ?, ?)"
	fields, err := json.Marshal(r.Fields)
	if err != nil {
		return errors.New("error marshalling fields. " + err.Error())
	}
	createdAt := r.CreatedAt.Format("2006-01-02 15:04:05")
	result, err := db.Exec(query, r.EventID, r.FormID, r.FormName, createdAt, r.User.ID, fields)
	if err != nil {
		return errors.New("error saving response. " + err.Error())
	}
	r.ID, err = result.LastInsertId()
	if err != nil {
		return errors.New("error getting last insert ID for response. " + err.Error())
	}
	return nil
}

type EventData struct {
	ResponseID   string  `json:"responseId"`
	SubmissionID string  `json:"submissionId"`
	RespondentID string  `json:"respondentId"`
	FormID       string  `json:"formId"`
	FormName     string  `json:"formName"`
	CreatedAt    string  `json:"createdAt"`
	Fields       []Field `json:"fields"`
}

type Field struct {
	Key     string       `json:"key"`
	Label   string       `json:"label"`
	Type    string       `json:"type"`
	Value   ValueWrapper `json:"value"`
	Options []Option     `json:"options"`
}

type ValueWrapper struct {
	String string
	Int    int64
	Float  float64
	Bool   bool
	List   []string
	IsNull bool
}

func (w *ValueWrapper) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch v := v.(type) {
	case string:
		w.String = v
	case int64:
		w.Int = v
	case float64:
		w.Float = v
	case bool:
		w.Bool = v
	case []interface{}:
		var list []string
		for _, v := range v {
			list = append(list, v.(string))
		}
		w.List = list
	case nil:
		w.IsNull = true
	default:
		return fmt.Errorf("unsupported type. value: %v", v)
	}
	return nil
}

type Option struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}
