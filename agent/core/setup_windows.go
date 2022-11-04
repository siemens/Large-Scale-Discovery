/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2021.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/siemens/GoScans/smb"
	"github.com/siemens/GoScans/ssl"
	"large-scale-discovery/agent/config"
	"large-scale-discovery/log"
	"os"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"
)

// checkConfigDependant tests OS specific configuration values by trying to initialize scan modules with them. This allows to
// discover invalid configurations at startup, instead of during runtime. Dynamic target arguments are replaced by
// dummy data.
func checkConfigDependant() error {

	// Dummy scan target arguments
	dummyLogger := log.GetLogger().Tagged("checkConfigDependant")
	dummyTarget := "127.0.0.1"
	dummyPort := 0
	dummyOtherNames := []string{"a", "b"}

	// Get config
	conf := config.GetConfig()

	// Run Smb test
	_, errSmb := smb.NewScanner(
		dummyLogger,
		dummyTarget,
		3,
		3,
		[]string{"a", "b", "c"},
		[]string{"a"},
		[]string{".a"},
		time.Date(2008, 01, 01, 00, 00, 00, 00, time.UTC),
		-1,
		true,
		conf.Authentication.Smb.Domain,
		conf.Authentication.Smb.User,
		conf.Authentication.Smb.Password,
	)
	if errSmb != nil {
		return fmt.Errorf("'%s': %s", smb.Label, errSmb)
	}

	// Decide trust store for SSL test
	var sslyzeAdditionalTruststore string
	if len(conf.Modules.Ssl.CustomTruststoreFile) == 0 {
		sslyzeAdditionalTruststore = SslOsTruststoreFile
	} else {
		sslyzeAdditionalTruststore = conf.Modules.Ssl.CustomTruststoreFile
	}

	// Run Ssl test
	_, errSsl := ssl.NewScanner(
		dummyLogger,
		conf.Paths.Sslyze,
		sslyzeAdditionalTruststore, // The ssl scan module will validate this path
		dummyTarget,
		dummyPort,
		dummyOtherNames,
	)
	if errSsl != nil {
		return fmt.Errorf("'%s': %s", ssl.Label, errSsl)
	}

	// Return nil if everything went fine
	return nil
}

// Windows implementation OS trust store generation
func generateTruststoreOs(truststoreOutputFile string) error {

	// Create given path if not existing
	path := filepath.Dir(truststoreOutputFile)
	errMkDirs := os.MkdirAll(path, 0770) // folder needs execute rights to be accessed
	if errMkDirs != nil {
		return fmt.Errorf("could prepare path for trust store: %s", errMkDirs)
	}

	// Export the CA windows store
	errCA := windowsExportTrustStore(truststoreOutputFile, false, "generated", "CA")
	if errCA != nil {
		return fmt.Errorf("could not export trust store: %s", errCA)
	}

	// Export the ROOT windows store
	errROOT := windowsExportTrustStore(truststoreOutputFile, true, "generated", "ROOT")
	if errROOT != nil {
		return fmt.Errorf("could not export trust store: %s", errROOT)
	}

	// Return nil as everything went fine
	return nil
}

// windowsExportTrustStore exports the windows trust store to given file
func windowsExportTrustStore(outputFile string, appendFile bool, version string, certstore string) error {

	const CryptENotFound = 0x80092004

	// Convert 2 utf16 string ptr
	ptrStr, _ := syscall.UTF16PtrFromString(certstore)

	// Open windows cert store
	store, err := syscall.CertOpenSystemStore(0, ptrStr)
	if err != nil {
		return err
	}

	// Make sure cert store is closed at the end
	defer func() { _ = syscall.CertCloseStore(store, 0) }()

	// Initialize certificate memory and certificate counter
	var cert *syscall.CertContext
	certCount := 0

	// Decide file mode
	if !appendFile {
		_ = os.Remove(outputFile)
	}

	// Open/Create custom pem file
	f, errF := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if errF != nil {
		return errF
	}

	// Make sure file gets closed on exit
	defer func() { _ = f.Close() }()

	// Iterate through the certificates
	for {

		cert, err = syscall.CertEnumCertificatesInStore(store, cert)

		if err != nil {
			if errno, ok := err.(syscall.Errno); ok {
				if errno == CryptENotFound {
					break
				}
			}
			return err
		}
		if cert == nil {
			break
		}

		// Copy the buf, since ParseCertificate does not create its own copy.
		buf := (*[1 << 20]byte)(unsafe.Pointer(cert.EncodedCert))[:]
		buf2 := make([]byte, cert.Length)
		copy(buf2, buf)

		// Parsed certificate from the bytes
		_, errC := x509.ParseCertificate(buf2)
		if errC != nil {
			continue
		}

		// Count certificate
		certCount++

		// Append the certificate to the file
		_ = pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: buf2})
		_, _ = f.Write([]byte("\r\n"))

		/*
			YAML generation disabled for now... as it looks like is not needed
			y.Write([]byte(fmt.Sprintf("- subject_name: %s\r\n", c.Subject.CommonName)))
			y.Write([]byte(fmt.Sprintf("  fingerprint:  %s\r\n", HashSha256(c.Raw))))
		*/
	}

	// Return nil as everything went fine
	return nil
}
