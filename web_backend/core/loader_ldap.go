/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
	scanUtils "github.com/siemens/GoScans/utils"
	"github.com/siemens/Large-Scale-Discovery/_build"
	"github.com/siemens/Large-Scale-Discovery/web_backend/database"
)

// init automatically registers the loader implemented in this file. If you don't want this loader,
// just remove it. You can also add your own loader by adding a file with your dedicated implementation.
func init() {
	if _build.DevMode {
	} else {

		// Register loader for initialization
		loaders = append(loaders, NewLoaderLdap([]string{}))
	}
}

var errNotFound = fmt.Errorf("LDAP entry not existing")
var errManyFound = fmt.Errorf("LDAP entry is ambiguous")
var msgUnknownUser = "Unknown LDAP entry."
var msgInvalidUser = "Invalid LDAP entry."

type errConnection struct {
	reason error
}

func (e errConnection) Error() string { return fmt.Sprintf("LDAP not reachable: %s", e.reason) }

type LoaderLdap struct {
	domains      []string
	ldapHost     string
	ldapPort     int
	ldapUser     string
	ldapPass     string
	ldapTimeout  time.Duration
	ldapCertPool *x509.CertPool
}

// NewLoaderLdap generates a new loader with the user domains it is responsible for. Everything
// else of the loader will be initialized later during core initialization, with the actual config values.
func NewLoaderLdap(domains []string) *LoaderLdap {
	return &LoaderLdap{
		domains: domains,
	}
}

// Domains returns the user domains this loader got registered for
func (l *LoaderLdap) Domains() []string {
	return l.domains
}

// Init validates loader settings and initializes the loader
func (l *LoaderLdap) Init(conf map[string]interface{}) error {

	// Check if config exists
	if conf == nil {
		return fmt.Errorf("loader configuration missing")
	}

	// Get and cast configuration values
	ldapCertificatePath, ldapCertificatePathSet := conf["ldap_certificate_path"]
	if !ldapCertificatePathSet {
		return fmt.Errorf("LDAP certificate path not set")
	}
	ldapCertificatePathValue, ldapCertificatePathCast := ldapCertificatePath.(string)
	if !ldapCertificatePathCast {
		return fmt.Errorf("LDAP certificate path value must be string")
	}
	if len(ldapCertificatePathValue) == 0 {
		return fmt.Errorf("LDAP certificate path empty")
	}
	ldapHost, ldapHostSet := conf["ldap_host"]
	if !ldapHostSet {
		return fmt.Errorf("LDAP host not set")
	}
	ldapHostValue, ldapHostCast := ldapHost.(string)
	if !ldapHostCast {
		return fmt.Errorf("LDAP host value must be string")
	}
	if len(ldapHostValue) == 0 {
		return fmt.Errorf("LDAP host empty")
	}
	ldapPort, okPortSet := conf["ldap_port"]
	ldapPortValue, okPortCast := ldapPort.(float64) // Go will assume float64 to accommodate possible values
	if !okPortCast {
		return fmt.Errorf("LDAP port value must be integer")
	}
	if !okPortSet {
		return fmt.Errorf("LDAP port not set")
	}
	ldapUser, okUserSet := conf["ldap_user"]
	if !okUserSet {
		return fmt.Errorf("LDAP user not set")
	}
	ldapUserValue, okUserCast := ldapUser.(string)
	if !okUserCast {
		return fmt.Errorf("LDAP user value must be string")
	}
	if len(ldapUserValue) == 0 {
		return fmt.Errorf("LDAP user not configured")
	}
	ldapPass, okPasswordSet := conf["ldap_password"]
	if !okPasswordSet {
		return fmt.Errorf("LDAP password not set")
	}
	ldapPassValue, okPasswordCast := ldapPass.(string)
	if !okPasswordCast {
		return fmt.Errorf("LDAP password value must be string")
	}
	if len(ldapPassValue) == 0 {
		return fmt.Errorf("LDAP password empty")
	}
	ldapTimeout, okTimeoutSet := conf["ldap_timeout_seconds"]
	ldapTimeoutValue, okTimeoutCast := ldapTimeout.(float64) // Go will assume float64 to accommodate possible values
	if !okTimeoutCast {
		return fmt.Errorf("LDAP timeout value must be integer")
	}
	if !okTimeoutSet {
		return fmt.Errorf("LDAP timeout not set")
	}

	// Validate and convert configuration arguments (they all come as strings)
	if ldapPortValue < 0 || ldapPortValue > 65535 {
		return fmt.Errorf("LDAP port invalid")
	}
	if ldapTimeoutValue <= 0 {
		return fmt.Errorf("LDAP timeout invalid")
	}
	ldapTimeoutDuration := time.Second * time.Duration(ldapTimeoutValue)

	// Prepare loader details
	ldapCaCertPem, errRead := os.ReadFile(ldapCertificatePathValue)
	if errRead != nil {
		return fmt.Errorf("could not read LDAP certificate: %s", errRead)
	}
	ldapCertPool := x509.NewCertPool()
	if ok := ldapCertPool.AppendCertsFromPEM(ldapCaCertPem); !ok {
		return fmt.Errorf("could not parse LDAP certificate")
	}

	// Initialize loader
	l.ldapHost = ldapHostValue
	l.ldapPort = int(ldapPortValue)
	l.ldapUser = ldapUserValue
	l.ldapPass = ldapPassValue
	l.ldapTimeout = ldapTimeoutDuration
	l.ldapCertPool = ldapCertPool

	// Test LDAP connection
	_, _, _, _, _, _, _, _, err := l.getLdap("")

	// Check if there is a fundamental issue
	if _, conErr := err.(errConnection); conErr {
		return err
	} else if !errors.Is(err, errNotFound) && !errors.Is(err, errManyFound) {
		return err
	}

	// Return nil as everything went fine
	return nil
}

// RefreshUser updates user attributes according to the implemented rules. This may be used to load/update user
// details from a remote repository. Changes are not yet committed! Might return one of FOUR kinds of error:
//   - A temporary error: Indicating a remote connection error. You may continue with cached data or return a
//     temporary error to the user.
//   - A public error (string): Indicating an error that is relevant information for the user, you may want to
//     return this message back to the user.
//     ATTENTION: If a public error message is returned, it also always comes in tandem with a detailed internal
//     error, which might be useful for additional logging.
//   - An internal error: Indicating an unexpected error. You should not continue, but return a generic error
//     message to the user.
func (l *LoaderLdap) RefreshUser(logger scanUtils.Logger, user *database.T_user) (errTemporary error, errPublic string, errInternal error) {

	// Log action
	logger.Debugf("Loading LDAP data.")

	// Query LDAP for user values
	userCompany, userDepartment, userStatus, userType,
		userName, userSurname, userGender, userCertificate, err := l.getLdap(user.Email)
	if err != nil {
		if _, conErr := err.(errConnection); conErr {
			return err, "", nil
		} else if errors.Is(err, errNotFound) || errors.Is(err, errManyFound) {
			return nil, msgUnknownUser, err
		} else { // Something is wrong, log and return
			return nil, "", err
		}
	}

	// Update user status to inactive if the user type changed to something invalid
	if userType == "T" { // I = internal account, X = external account, T = team or function account
		return nil, msgInvalidUser, fmt.Errorf("LDAP user type '%s' invalid ", userType)
	}

	// Update user values
	user.Department = userDepartment
	user.Name = userName
	user.Surname = userSurname
	user.Gender = userGender
	user.Certificate = userCertificate

	// Update user active flag, but only if it wasn't already disabled. If it was already disabled, leave it disabled
	// because it might have been disabled manually be an administrator.
	if user.Active {

		// A = active employment, R = dormant employment, G = AÜG-employee, F = none-corporate entries
		if userStatus != "A" && userStatus != "" {
			logger.Infof("User '%s' has invalid contract status '%s'.", user.Email, userStatus)
			user.Active = false
		}
	}

	// Set user company or fall back to user-dedicated company to avoid unintended groups
	if len(userCompany) > 0 {

		// Don't allow loaders to return e-mail addresses as company names, otherwise there could be a
		// security-relevant collision with users where company equals e-mail address (users not part of any
		// other company).
		if strings.Contains(userCompany, "@") {
			return nil, "", fmt.Errorf("forbidden company name '%s' returned", userCompany)
		}

		// Build company string to assign user to
		user.Company = userCompany
	} else {
		user.Company = user.Email
	}

	// Return nil as everything went fine
	return nil, "", nil
}

// getLdap queries user attributes for a given email address via LDAP
func (l *LoaderLdap) getLdap(email string) (
	userCompany string,
	userDepartment string,
	userStatus string,
	userType string,
	userName string,
	userSurname string,
	userGender string,
	userCert []byte,
	err error,
) {
	// Prepare memory
	userCert = make([]byte, 0)

	// Prepare the LDAP options.
	opts := []ldap.DialOpt{
		ldap.DialWithDialer(&net.Dialer{Timeout: l.ldapTimeout}),     // DialWithDialer updates net.Dialer in DialContext.
		ldap.DialWithTLSConfig(&tls.Config{RootCAs: l.ldapCertPool}), // Use the LDAP certificate as root
	}

	// Try to establish LDAPs connection
	conn, errDial := ldap.DialURL(fmt.Sprintf("ldaps://%s:%d", l.ldapHost, l.ldapPort), opts...)
	if errDial != nil {
		err = errConnection{errDial}
		return
	}

	// Make sure connection gets closed on exit
	defer func() { _ = conn.Close() }()

	// Bind LDAP connection with authentication
	errBind := conn.Bind(l.ldapUser, l.ldapPass)
	if errBind != nil {
		err = errConnection{errBind}
		return
	}

	// Define according LDAP attribute names
	attrCompany := "organizationName"
	attrDepartment := "department"
	attrUserStatus := "contractstatus" // Status of user: A = active employment, R = dormant employment, G = AÜG-employee, F = none-corporate entries
	attrUserType := "userType"         // Kind of user: I = internal account, X = external account, T = team or function account
	attrUserName := "gn"
	attrUserSurname := "sn"
	attrUserGender := "gender"
	attrUserCert := "userCertificate;binary"

	// Define entry type attribute
	attrRecordType := "recordType" // Kind of entry: H = main entry, additional entry

	// Prepare search request
	searchRequest := ldap.NewSearchRequest(
		"", // The base dn to search
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(&(objectClass=Person)(mail=%s))", email), // The filter to apply
		[]string{
			attrCompany,
			attrDepartment,
			attrUserStatus,
			attrUserType, attrUserName,
			attrUserSurname,
			attrUserGender,
			attrUserCert,
			attrRecordType,
		}, // Attributes to retrieve
		nil,
	)

	// Execute search
	var sr *ldap.SearchResult
	sr, err = conn.Search(searchRequest)
	if err != nil {
		return
	}

	// Abort if nothing was found
	if len(sr.Entries) < 1 { // Zero results
		err = errNotFound
		return
	}

	// Prepare memory for result entry
	var entry *ldap.Entry = nil

	// Grab suitable entry
	if len(sr.Entries) == 1 {

		// If there is only one result, it can be considered the main entry
		entry = sr.Entries[0]
	} else {

		// If there are multiple results, only proceed with main entry
		for _, ldapEntry := range sr.Entries {
			entryType := ldapEntry.GetAttributeValue(attrRecordType)
			if entryType == "H" { // H = Main Entry
				entry = ldapEntry
				break
			}
		}
	}

	// If result was ambiguous, return error
	if entry == nil { // More than one result, but none marked as main entry
		err = errManyFound
		return
	}

	// Extract attributes
	userCompany = entry.GetAttributeValue(attrCompany)
	userDepartment = entry.GetAttributeValue(attrDepartment)
	userStatus = entry.GetAttributeValue(attrUserStatus)
	userType = entry.GetAttributeValue(attrUserType)
	userName = entry.GetAttributeValue(attrUserName)
	userSurname = entry.GetAttributeValue(attrUserSurname)
	userGender = entry.GetAttributeValue(attrUserGender)

	// Get certificates (there might be more) and select most plausible one
	certs := entry.GetAttributeValues(attrUserCert)
	for _, cert := range certs {

		// Convert back to bytes
		bCert := []byte(cert)

		// Parse certificate or skip to next one if it couldn't be parsed
		c, errParse := x509.ParseCertificate(bCert)
		if errParse != nil {
			continue // Try next one
		}

		// Select certificate, in case we can't find a better one
		if len(userCert) == 0 {
			userCert = bCert
		}

		// Check whether certificate is valid for data encryption usage
		if x509.KeyUsageDataEncipherment&c.KeyUsage == x509.KeyUsageDataEncipherment {

			// Select certificate
			userCert = bCert

			// Best option found, stop search
			break
		}
	}

	// Return nil as everything went fine
	return userCompany, userDepartment, userStatus, userType, userName, userSurname, userGender, userCert, err
}
