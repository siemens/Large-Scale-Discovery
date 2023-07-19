/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package utils

import (
	"encoding/json"
	"fmt"
	scanUtils "github.com/siemens/GoScans/utils"
	"io/ioutil"
	"net/mail"
	"regexp"
)

type Smtp struct {

	// SMTP settings
	Server     string         `json:"server"`
	Port       uint16         `json:"port"`
	Username   string         `json:"username"`
	Password   string         `json:"password"`
	Subject    string         `json:"subject"`
	Sender     mail.Address   `json:"sender"`
	Recipients []mail.Address `json:"recipients"`

	// Security settings
	OpensslPath         string   `json:"openssl_path"`
	SignatureCertPath   string   `json:"signature_cert"`  // Sender certificate for e-mail signature
	SignatureCert       []byte   `json:"-"`               // Loaded sender certificate for e-mail signature
	SignatureKeyPath    string   `json:"signature_key"`   // Sender private key for e-mail signature
	SignatureKey        []byte   `json:"-"`               // Loaded sender private key for e-mail signature
	EncryptionCertPaths []string `json:"recipient_certs"` // Encryption certificates for the recipients above
	EncryptionCerts     [][]byte `json:"-"`               // Loaded encryption certificates for the recipients above
	TempDir             string   `json:"temp_dir"`
}

// UnmarshalJSON reads a JSON file, validates values and populates the configuration struct
func (s *Smtp) UnmarshalJSON(b []byte) error {

	// Prepare temporary auxiliary data structure to load raw Json data
	type aux Smtp
	var raw aux

	// Unmarshal serialized Json into temporary auxiliary structure
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	// Validate values
	if raw.Sender.Address != "" {
		if !IsPlausibleEmail(raw.Sender.Address) {
			return fmt.Errorf("invalid sender e-mail address")
		}
	}
	if len(raw.Recipients) == 0 {
		return fmt.Errorf("at least one log recipient required")
	}
	for _, recipient := range raw.Recipients {
		if !IsPlausibleEmail(recipient.Address) {
			return fmt.Errorf("invalid recipient e-mail address")
		}
	}
	if raw.Server == "" {
		return fmt.Errorf("invalid SMTP server")
	}
	if !scanUtils.IsValidIp(raw.Server) && !scanUtils.IsValidHostname(raw.Server) {
		return fmt.Errorf("invalid SMTP server")
	}

	// Check if necessary files are defined
	if (len(raw.SignatureCertPath) != 0) != (len(raw.SignatureKeyPath) != 0) {
		return fmt.Errorf("either signature certificate and key required or none of both")
	}

	// Check if OpenSSL path is configured if necessary
	if len(raw.SignatureCertPath) != 0 || len(raw.EncryptionCertPaths) != 0 {
		if len(raw.OpensslPath) == 0 {
			return fmt.Errorf("OpenSSL path not set")
		}
		if errOpenSSL := scanUtils.IsValidExecutable(raw.OpensslPath); errOpenSSL != nil {
			return errOpenSSL
		}
	}

	// Copy loaded Json values to actual
	*s = Smtp(raw)

	// Load signature certificate from file
	if s.SignatureCertPath != "" {
		var errCert error
		s.SignatureCert, errCert = ioutil.ReadFile(s.SignatureCertPath)
		if errCert != nil {
			return fmt.Errorf("unable to load sender certificate: %s", errCert)
		}
	}

	// Load signature key from file
	if s.SignatureKeyPath != "" {
		var errKey error
		s.SignatureKey, errKey = ioutil.ReadFile(s.SignatureKeyPath)
		if errKey != nil {
			return fmt.Errorf("unable to load sender key: %s", errKey)
		}
	}

	// Load administrators encryption certificates from file
	s.EncryptionCerts = make([][]byte, 0, len(s.EncryptionCertPaths))
	for _, p := range s.EncryptionCertPaths {
		cert, errEncCert := ioutil.ReadFile(p)
		if errEncCert != nil {
			return fmt.Errorf("unable to load recipient certificate: %s", errEncCert)
		}
		s.EncryptionCerts = append(s.EncryptionCerts, cert)
	}

	// Return nil as everything is valid
	return nil
}

// IsPlausibleEmail validates whether a given string is a plausible e-mail address
func IsPlausibleEmail(mail string) bool {
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if !re.MatchString(mail) {
		return false
	}
	return true
}
