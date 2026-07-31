/*
 * Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
 *
 * Copyright (c) Siemens AG, 2016-2026.
 *
 * This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
 * directory or visit <https://opensource.org/licenses/MIT>.
 *
 * SPDX-License-Identifier: MIT
 */

package database

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"gorm.io/gorm"
)

const EventLogin Event = "Login"                  // Successful user login
const EventDbPassword Event = "Database Password" // Database password reset
const EventScopeCreate Event = "Scope Created"    // Scan scope creation
const EventScopeSecret Event = "Scope Secret"     // Scan scope secret reset
const EventViewGrant Event = "User Granted"       // View access grant
const EventViewToken Event = "Token Generated"    // View token creation
const EventDatabaseAdd Event = "Database Added"   // Database server creation
const EventApiToken Event = "API Token"           // API token creation

// Event identifies an auditable user action
type Event string

// T_event represents an auditable user action persisted in the backend database
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

	User *T_user `gorm:"foreignKey:IdTUser;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"user"` // User must be pointer *T_user, because it can be null OnDelete
}

// BeforeSave sanitizes an event before GORM writes it to the database
func (event *T_event) BeforeSave(tx *gorm.DB) error {

	// Initialize sanitizer
	b := bluemonday.StrictPolicy()

	// Sanitize value
	event.Email = b.Sanitize(event.Email)
	event.Email = strings.ToLower(event.Email) // Standard format of e-mail addresses shall be lower case
	tx.Statement.SetColumn("email", event.Email)

	// Sanitize value
	event.Event = Event(b.Sanitize(string(event.Event)))
	tx.Statement.SetColumn("event", event.Event)

	// Sanitize value
	event.EventDetail = b.Sanitize(event.EventDetail)
	tx.Statement.SetColumn("event_detail", event.EventDetail)

	// Return nil as everything went fine
	return nil
}

// LogEvent creates an event log entry in the database
func LogEvent(userEntry *T_user, event Event, eventDetail string) error {

	// Check if mandatory user information is available
	if userEntry.Id == 0 || userEntry.Email == "" || !utils.IsPlausibleEmail(userEntry.Email) || event == "" {
		return fmt.Errorf("invalid event details")
	}

	// Check whether event value is valid
	if errEvent := validEvent(event); errEvent != nil {
		return errEvent
	}

	// Prepare event entry
	entry := T_event{
		IdTUser:     userEntry.Id,
		Email:       userEntry.Email,
		Timestamp:   time.Now(),
		Event:       event,
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

// GetEvents returns optionally limited events of the requested kind at or after the optional timestamp
func GetEvents(event Event, since time.Time, limit *int) ([]T_event, error) {

	// Check whether event value is valid
	if errEvent := validEvent(event); errEvent != nil {
		return nil, errEvent
	}

	// Prepare query result
	var eventEntries []T_event

	// Query related events form database
	q := backendDb.
		Preload("User").
		Where("event = ?", event)

	// Add where condition
	if !since.IsZero() {
		q = q.Where("timestamp >= ?", since)
	}

	// Restrict the query to the newest requested entries
	if limit != nil && *limit >= 0 {
		if *limit == 0 {
			return []T_event{}, nil
		}
		q = q.Order("timestamp DESC").Order("id DESC").Limit(*limit)
	} else {
		q = q.Order("timestamp ASC").Order("id ASC")
	}

	// Execute query
	errDb := q.Find(&eventEntries).Error
	if errDb != nil {
		return nil, errDb
	}

	// Restore chronological order after selecting the newest limited entries
	if limit != nil && *limit > 0 {
		slices.Reverse(eventEntries)
	}

	// Return list of related events
	return eventEntries, nil
}

// GetEventsAll returns all optionally limited events at or after the optional timestamp
func GetEventsAll(since time.Time, limit *int) ([]T_event, error) {

	// Prepare query result
	var eventEntries []T_event

	// Query related events form database
	q := backendDb.
		Model(&T_event{}).
		Preload("User")

	// Add where condition
	if !since.IsZero() {
		q = q.Where("timestamp >= ?", since)
	}

	// Restrict the query to the newest requested entries
	if limit != nil && *limit >= 0 {
		if *limit == 0 {
			return []T_event{}, nil
		}
		q = q.Order("timestamp DESC").Order("id DESC").Limit(*limit)
	} else {
		q = q.Order("timestamp ASC").Order("id ASC")
	}

	// Execute query
	errDb := q.Find(&eventEntries).Error
	if errDb != nil {
		return nil, errDb
	}

	// Restore chronological order after selecting the newest limited entries
	if limit != nil && *limit > 0 {
		slices.Reverse(eventEntries)
	}

	// Return list of related events
	return eventEntries, nil
}

// validEvent checks whether an event is supported
func validEvent(event Event) error {

	// Check whether event value is valid
	switch event {
	case EventLogin, EventDbPassword, EventScopeCreate, EventScopeSecret, EventViewGrant, EventViewToken, EventDatabaseAdd, EventApiToken:
		// Ok
	default:
		return fmt.Errorf("invalid event '%s'", event)
	}

	// Return nil as everything went fine
	return nil
}
