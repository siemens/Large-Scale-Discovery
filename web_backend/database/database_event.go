/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package database

import (
	"fmt"
	"github.com/microcosm-cc/bluemonday"
	"gorm.io/gorm"
	"large-scale-discovery/utils"
	"strings"
	"time"
)

type Event string

// Definition of some standard event values
const (
	EventLogin       Event = "Login"
	EventDbPassword  Event = "Database Password"
	EventScopeCreate Event = "Scope Created"
	EventViewGrant   Event = "User Granted"
	EventViewToken   Event = "Token Generated"
)

type T_event struct {
	// - Set the JSON ignore flag (json:"-") for sensitive columns that may NEVER be leaked by a JSON response.
	// - Make columns "not null" if possible. Otherwise, use null-types (e.g. sql.NullString).
	// - Avoid 'default' constraints or gorm will replace empty values (0, "", false) with set default values on CREATE!
	// - Define a lower-snake-case json name for every attribute.
	Id          uint64    `gorm:"column:id;primaryKey" json:"-"`
	IdTUser     uint64    `gorm:"column:id_t_user;type:int" json:"-"`
	Email       string    `gorm:"column:email;not null" json:"email"`
	Timestamp   time.Time `gorm:"column:timestamp;default:CURRENT_TIMESTAMP" json:"timestamp"`
	Event       Event     `gorm:"column:event;not null" json:"event"`
	EventDetail string    `gorm:"column:event_detail;default:''" json:"event_detail"`

	User *T_user `gorm:"foreignKey:IdTUser;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"user"`
}

// NewEvent creates an event log entry in the database.
func NewEvent(user *T_user, eventType Event, eventDetail string) error {

	// Check if mandatory user information is available
	if user.Id == 0 || user.Email == "" || !utils.IsPlausibleEmail(user.Email) || eventType == "" {
		return fmt.Errorf("invalid event details")
	}

	// Check whether event type value is valid
	if err := validEvent(eventType); err != nil {
		return err
	}

	// Prepare event entry
	entry := T_event{
		IdTUser:     user.Id,
		Email:       user.Email,
		Timestamp:   time.Now(),
		Event:       eventType,
		EventDetail: eventDetail,
	}

	// Write event to database
	errDb := backendDb.Create(&entry).Error
	if errDb != nil {
		return errDb
	}

	// Return nil as everything went fine
	return nil
}

// BeforeSave is a GORM hook that's executed every time the user object is written to the DB. This should be used to
// do some data sanitization, e.g. to strip illegal HTML tags in user attributes or to convert values to a certain
// format.
func (user *T_event) BeforeSave(tx *gorm.DB) error {

	// Initialize sanitizer
	b := bluemonday.StrictPolicy()

	// Sanitize value
	user.Email = b.Sanitize(user.Email)
	user.Email = strings.ToLower(user.Email) // Standard format of e-mail addresses shall be lower case.
	tx.Statement.SetColumn("email", user.Email)

	// Sanitize value
	user.Event = Event(b.Sanitize(string(user.Event)))
	tx.Statement.SetColumn("event", user.Event)

	// Sanitize value
	user.EventDetail = b.Sanitize(user.EventDetail)
	tx.Statement.SetColumn("event_detail", user.EventDetail)

	// Return nil as everything went fine
	return nil
}

func GetEvents(eventType Event, since time.Time) ([]T_event, error) {

	// Check whether event type value is valid
	if err := validEvent(eventType); err != nil {
		return nil, err
	}

	// Prepare query result
	var events []T_event

	// Query related events form database
	q := backendDb.
		Preload("User").
		Where("event = ?", eventType).
		Order("timestamp ASC")

	// Add where condition
	if !since.IsZero() {
		q = q.Where("timestamp >= ?", since)
	}

	// Execute query
	errDb := q.Find(&events).Error
	if errDb != nil {
		return nil, errDb
	}

	// Return list of related events
	return events, nil
}

// validEvent checks whether an event is an existing one
func validEvent(event Event) error {

	// Check whether event type value is valid
	switch event {
	case EventLogin, EventDbPassword, EventScopeCreate, EventViewGrant, EventViewToken:
		// Ok
	default:
		return fmt.Errorf("invalid event type")
	}

	// Return nil as everything went fine
	return nil
}
