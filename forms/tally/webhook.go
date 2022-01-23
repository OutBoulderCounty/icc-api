package tally

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
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
	var formID int64
	var err error
	for i := 0; i < len(fields); i++ {
		field := fields[i]
		label := field.Label
		value := field.Value
		if label == "user_id" {
			userID, err = strconv.ParseInt(value.(string), 10, 64)
			if err != nil {
				return nil, errors.New("error parsing user ID. " + err.Error())
			}
		} else if label == "form_id" {
			formID, err = strconv.ParseInt(value.(string), 10, 64)
			if err != nil {
				return nil, errors.New("error parsing form ID. " + err.Error())
			}
		}
	}
	if userID == 0 {
		return nil, errors.New("no user ID in event data")
	}
	createdAt, err := time.Parse(time.RFC3339Nano, e.CreatedAt)
	if err != nil {
		return nil, errors.New("error parsing created at. " + err.Error())
	}

	response := Response{
		FormID:    formID,
		EventID:   e.EventID,
		CreatedAt: createdAt,
		UserID:    userID,
		Fields:    fields,
	}
	err = response.Save(db)
	if err != nil {
		return nil, errors.New("error saving response. " + err.Error())
	}
	return &response, nil
}

type Response struct {
	ID        int64     `json:"id"`
	EventID   string    `json:"event_id"`
	FormID    int64     `json:"form_id"`
	CreatedAt time.Time `json:"created_at"`
	UserID    int64     `json:"user"`
	Fields    []Field   `json:"fields"`
}

func (r *Response) Save(db *sql.DB) error {
	if r.ID != 0 {
		return errors.New("response already saved")
	}
	query := "insert into tally_responses (event_id, form_id, created_at, user_id, fields) values (?, ?, ?, ?, ?)"
	fields, err := json.Marshal(r.Fields)
	if err != nil {
		return errors.New("error marshalling fields. " + err.Error())
	}
	createdAt := r.CreatedAt.Format("2006-01-02 15:04:05")
	result, err := db.Exec(query, r.EventID, r.FormID, createdAt, r.UserID, fields)
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
	Key     string      `json:"key"`
	Label   string      `json:"label"`
	Type    string      `json:"type"`
	Value   interface{} `json:"value"`
	Options []Option    `json:"options"`
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
