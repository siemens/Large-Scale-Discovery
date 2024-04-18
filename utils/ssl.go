/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package utils

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// TlsConfigFactory returns a secure SSL connection configuration
func TlsConfigFactory() *tls.Config {
	// Return TLS config
	// Good cipher suites are maintained by Golang and shouldn't be manually set, unless you really know what you
	// are doing: https://github.com/golang/go/issues/41068. Just make sure to always compile wit a current Golang
	// version.
	return &tls.Config{
		MinVersion:               tls.VersionTLS12,
		MaxVersion:               tls.VersionTLS13,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
	}
}

// PinnedTlsConfigFactory returns an SSL client configuration that is verified by fingerprint matching
// against a provided public key file. This way a secure SSL connection can be established without relying on PKI.
func PinnedTlsConfigFactory(publicKeyPath string) (*tls.Config, error) {

	// Read public key
	b, errRead := os.ReadFile(publicKeyPath)
	if errRead != nil {
		return nil, errRead
	}

	// Generate fingerprint of broker's public key
	fingerprint, _ := pem.Decode(b)
	if fingerprint == nil {
		return nil, fmt.Errorf("could not prepare fingerprint from '%s'", publicKeyPath)
	}

	// Define certificate fingerprint checking routine
	checkFingerprintFunc := func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		for _, rawCert := range rawCerts {
			if bytes.Equal(rawCert, fingerprint.Bytes) { // Fingerprint matching
				return nil
			}
		}
		return fmt.Errorf("invalid certificate fingerprint") // Fingerprint not matching
	}

	// Create the tls conf
	tlsConfig := &tls.Config{
		InsecureSkipVerify:    true,                 // We'll verify the public key fingerprint instead of relying on PKI
		VerifyPeerCertificate: checkFingerprintFunc, // Verify broker's public key fingerprint
	}

	// Return config as everything went fine
	return tlsConfig, nil
}
