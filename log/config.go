package log

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap/zapcore"
	"large-scale-discovery/utils"
	"net/mail"
	"time"
)

func DefaultLogSettingsFactory() Settings {

	return Settings{
		Console: &ConsoleHandler{
			Enabled: true,
			Level:   zapcore.InfoLevel,
		},
		File: &FileHandler{
			Enabled: true,
			Level:   zapcore.DebugLevel,
			Path:    "./logs/application.log",
			SizeMb:  100,
			History: 10,
		},
		Smtp: &SmtpHandler{
			Enabled:              false,
			Level:                zapcore.WarnLevel,
			LevelPriority:        zapcore.ErrorLevel,
			DelayMinutes:         int((time.Hour * 24).Minutes()),  // 1 Day
			DelayPriorityMinutes: int((time.Minute * 5).Minutes()), // 5 Minutes
			Connector: utils.Smtp{
				Server:   "mail.domain.com",
				Port:     25,
				Username: "",
				Password: "",
				Subject:  "Application Log",
				Sender:   mail.Address{Name: "Large-Scale Discovery", Address: "user1@domain.tld"},
				Recipients: []mail.Address{
					{Name: "User2", Address: "user2@domain.tld"},
					{Name: "User3", Address: "user3@domain.tld"},
				},
				OpensslPath:         opensslPath,
				SignatureCertPath:   "",
				SignatureKeyPath:    "",
				EncryptionCertPaths: []string{},
				TempDir:             ""},
		},
	}
}

//
// JSON structure of configuration
//

type ConsoleHandler struct {
	Enabled bool          `json:"enabled"`
	Level   zapcore.Level `json:"level"` // Zap log levels: -1=debug, 0=info, 1=warn and 2=error
}

// UnmarshalJSON reads a JSON file, validates values and populates the configuration struct
func (h *ConsoleHandler) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw Json data
	type aux ConsoleHandler
	var raw aux

	// Unmarshal serialized Json into temporary auxiliary structure
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	// Validate values
	if raw.Level < zapcore.DebugLevel || raw.Level > zapcore.ErrorLevel {
		return fmt.Errorf("invalid log level (-1=debug, 0=info, 1=warn, 2=error only)")
	}

	// Copy loaded Json values to actual
	*h = ConsoleHandler(raw)

	// Return nil as everything is valid
	return nil
}

type FileHandler struct {
	Enabled bool          `json:"enabled"`
	Level   zapcore.Level `json:"level"` // Zap log levels: -1=debug, 0=info, 1=warn and 2=error

	Path    string `json:"path"`
	SizeMb  int    `json:"size_mb"`
	History int    `json:"history"`
}

// UnmarshalJSON reads a JSON file, validates values and populates the configuration struct
func (h *FileHandler) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw Json data
	type aux FileHandler
	var raw aux

	// Unmarshal serialized Json into temporary auxiliary structure
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	// Validate values
	if raw.Level < zapcore.DebugLevel || raw.Level > zapcore.ErrorLevel {
		return fmt.Errorf("invalid log level (-1=debug, 0=info, 1=warn, 2=error only)")
	}
	if raw.SizeMb <= 0 {
		return fmt.Errorf("invalid file size")
	}
	if raw.History <= 0 {
		return fmt.Errorf("invalid file history")
	}

	// Copy loaded Json values to actual
	*h = FileHandler(raw)

	// Return nil as everything is valid
	return nil
}

type SmtpHandler struct {
	Enabled bool          `json:"enabled"`
	Level   zapcore.Level `json:"level"` // Zap log levels: -1=debug, 0=info, 1=warn and 2=error

	// Log mechanics
	LevelPriority        zapcore.Level `json:"level_priority"`         // The minimum log level to handle with priority (send out faster)
	DelayMinutes         int           `json:"delay_minutes"`          // Serialized integer value
	Delay                time.Duration `json:"-"`                      // Deserialized duration representation
	DelayPriorityMinutes int           `json:"delay_priority_minutes"` // Serialized integer value
	DelayPriority        time.Duration `json:"-"`                      // Deserialized duration representation

	Connector utils.Smtp `json:"connector"`
}

func (h *SmtpHandler) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw Json data
	type aux SmtpHandler
	var raw aux

	// Unmarshal serialized Json into temporary auxiliary structure
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	// Validate values
	if raw.Level < zapcore.DebugLevel || raw.Level > zapcore.ErrorLevel {
		return fmt.Errorf("invalid log level (-1=debug, 0=info, 1=warn, 2=error only)")
	}
	if raw.LevelPriority < zapcore.DebugLevel || raw.LevelPriority > zapcore.ErrorLevel {
		return fmt.Errorf("invalid priority log level (-1=debug, 0=info, 1=warn, 2=error only)")
	}
	if raw.Level > raw.LevelPriority {
		return fmt.Errorf("priority log level may not be less than log level")
	}
	if raw.DelayMinutes <= 0 {
		return fmt.Errorf("invalid smtp delay")
	}
	if raw.DelayPriorityMinutes <= 0 {
		return fmt.Errorf("invalid smtp priority delay")
	}
	if raw.DelayMinutes < raw.DelayPriorityMinutes {
		return fmt.Errorf("delay may not be less than priority delay")
	}

	// Copy loaded Json values to actual config
	*h = SmtpHandler(raw)

	// Set unserializable values
	h.Delay = time.Duration(h.DelayMinutes) * time.Minute
	h.DelayPriority = time.Duration(h.DelayPriorityMinutes) * time.Minute

	// Return nil as everything is valid
	return nil
}

// Settings can be saved to and loaded from a JSON file. It holds settings that are relevant for initializing a
// logger. All the sub logger can either be omitted or disabled by setting Enabled to false
type Settings struct {
	Console *ConsoleHandler `json:"console"`
	File    *FileHandler    `json:"file"`
	Smtp    *SmtpHandler    `json:"smtp"`
}

// UnmarshalJSON reads a JSON file, validates values and populates the configuration struct
func (s *Settings) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw Json data
	type aux Settings
	var raw aux

	// Unmarshal serialized Json into temporary auxiliary structure
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	// Check whether logger settings were there
	if raw.File == nil && raw.Console == nil && raw.Smtp == nil {
		return fmt.Errorf("logger configuration invalid")
	}

	// Copy loaded Json values to actual
	*s = Settings(raw)

	// Return nil as everything is valid
	return nil
}
