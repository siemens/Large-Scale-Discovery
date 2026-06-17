/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package scopedb

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgconn"
	"github.com/siemens/GoScans/ssl"
	scanUtils "github.com/siemens/GoScans/utils"
	managerdb "github.com/siemens/Large-Scale-Discovery/manager/database"
	"github.com/siemens/Large-Scale-Discovery/utils"
	"gorm.io/gorm"
)

// PrepareSslResult creates an info entry in the database to indicate an active process.
func PrepareSslResult(
	logger scanUtils.Logger,
	scanScope *managerdb.T_scan_scope,
	idsTDiscoveryService []uint64,
	scanStarted sql.NullTime,
	scanIp string,
	scanHostname string,
	wg *sync.WaitGroup,
) {

	// Log potential panics before letting them move on
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf(fmt.Sprintf("Panic: %s%s", r, scanUtils.StacktraceIndented("\t")))
			panic(r)
		}
	}()

	// Notify the WaitGroup that we're done
	defer wg.Done()

	// Log action
	logger.Debugf("Preparing SSL/TLS info entry.")

	// Return if there were no results to be created
	if len(idsTDiscoveryService) == 0 {
		logger.Debugf("No discovery services for SSL/TLS info entries provided.")
		return
	}

	// Open scope's database
	scopeDb, errHandle := managerdb.GetScopeDbHandle(logger, scanScope)
	if errHandle != nil {
		logger.Errorf("Could not open scope database: %s", errHandle)
		return
	}

	// Iterate service IDs to create database info entries for
	infoEntries := make([]managerdb.T_ssl, 0, len(idsTDiscoveryService))
	for _, idTDiscoveryService := range idsTDiscoveryService {

		// Prepare info entry
		infoEntries = append(infoEntries, managerdb.T_ssl{
			IdTDiscoveryService: idTDiscoveryService,
			ColumnsScan: managerdb.ColumnsScan{
				ScanStarted:  scanStarted,
				ScanFinished: sql.NullTime{Valid: false},
				ScanStatus:   scanUtils.StatusRunning,
				ScanIp:       scanIp,
				ScanHostname: scanHostname,
			},
		})
	}

	// Create info entries, if not yet existing. It might exist from a previous scan attempt (if the scan crashed).
	// This check which is applied here is specific to postgres.
	// Use a new gorm session and force a limit on how many Entries can be batched, as we otherwise might
	// exceed the database's limit of 65535 parameters
	errDb := scopeDb.
		Session(&gorm.Session{CreateBatchSize: managerdb.MaxBatchSizeSsl}).
		Create(&infoEntries).Error

	// Extract Postgres error if existing
	var errPg *pgconn.PgError
	errors.As(errDb, &errPg)

	// Handle database errors
	if errPg != nil && errPg.Code == "23505" { // Code for Postgres' unique constraint violation

		// Fall back to inserting the entries one by one to ensure as many entries as possible being added to the db
		for _, entry := range infoEntries {
			errDb2 := scopeDb.Create(&entry).Error

			// Extract Postgres error if existing
			var errPg2 *pgconn.PgError
			errors.As(errDb2, &errPg2)

			// Handle database errors
			if errPg2 != nil && errPg2.Code == "23505" { // Code for Postgres' unique constraint violation
				logger.Debugf("SSL info entry %d already existing.", entry.IdTDiscoveryService)
			} else if errDb2 != nil {
				logger.Errorf("SSL info entry %d could not be created: %s", entry.IdTDiscoveryService, errDb2)
			}
		}
	} else if errDb != nil {
		logger.Errorf("SSL info entries could not be created: %s", errDb)
	}
}

// SaveSslResult parses a result and adds its data into the database. Furthermore, it sets the "scan_finished"
// timestamp to indicate a completed process.
func SaveSslResult(
	logger scanUtils.Logger,
	scanScope *managerdb.T_scan_scope,
	idTDiscoveryService uint64, // T_discovery_services ID this result belongs to
	result *ssl.Result,
) error {

	// Log action
	logger.Debugf("Saving SSL/TLS scan result.")

	// Open scope's database
	scopeDb, errHandle := managerdb.GetScopeDbHandle(logger, scanScope)
	if errHandle != nil {
		return fmt.Errorf("could not open scope database: %s", errHandle)
	}

	// Start transaction on the scoped db. The new Transaction function will commit if the provided function
	// returns nil and rollback if an error is returned.
	return scopeDb.Transaction(func(txScopeDb *gorm.DB) error {
		// Check the condition of the info entry stored in the database. It got created the first time the target was
		// taken from the cache (info entry should never be removed from the scope database again!). However,
		// 		A) the scan could have crashed before and been restarted by the broker after its timeout
		// 		B) the broker could have been down for a prolonged time, restarting the scan after an assumed timeout.
		// 		   Both scan instances might deliver results in arbitrary order.
		//
		// Case A: The scan_finished timestamp is not set, because this is the first scan instances delivering results
		// 		    => Save results, set scan_finished timestamp
		// Case B: The scan_finished timestamp is already set, because this is a later scan instance delivering results
		// 			=> Discard results (to prevent later manipulation by client), but update scan_finished timestamp (to
		// 			   allow users proper investigations in case of IDS alerts).

		// Prepare memory for result
		var infoEntry = managerdb.T_ssl{}
		var infoEntryCount int64

		// Query info entry
		errDb := txScopeDb.Where("id_t_discovery_service = ?", idTDiscoveryService).
			First(&infoEntry).
			Count(&infoEntryCount).Error
		if errDb != nil {
			return fmt.Errorf("could not find associated info entry in scope db: %s", errDb)
		}

		// Check if info entry was found
		if infoEntryCount != 1 {
			logger.Infof("Dropping SSL/TLS result of vanished discovery result.") // Scan database might have been reset or cleaned up
			return nil
		}

		// Add scan results, if this is the first result delivered (scan_finished not yet set)
		if !infoEntry.ScanFinished.Valid {

			// Add scan results
			errAdd := addTlsResults(txScopeDb, infoEntry.Id, idTDiscoveryService, result)
			if errAdd != nil {
				return fmt.Errorf("could not insert result into scope db: %s", errAdd)
			}

			// Add scan status to info entry. Some scan modules might have even more data stored in the info column
			infoEntry.ScanStatus = result.Status
		}

		// Update finished timestamp. Set or update it to the current time.
		infoEntry.ScanFinished = sql.NullTime{Time: time.Now(), Valid: true}

		// Update db with changes
		errDb2 := txScopeDb.Model(&infoEntry).
			Where("id_t_discovery_service = ?", idTDiscoveryService).
			Updates(infoEntry).Error
		if errDb2 != nil {
			return fmt.Errorf("could not update entry in scope db: %s", errDb2)
		}

		// Return nil as everything went fine
		return nil
	})
}

func addTlsResults(txScopeDb *gorm.DB, tSslInfoId uint64, idTDiscoveryService uint64, result *ssl.Result) error {

	// Prepare entry slices
	resultCiphers := make([]managerdb.T_ssl_cipher, 0, 10)
	resultCertificates := make([]managerdb.T_ssl_certificate, 0, 5)
	resultSettings := make([]managerdb.T_ssl_setting, 0, len(result.Data))
	resultIssues := make([]managerdb.T_ssl_issue, 0, len(result.Data))

	// Iterate scan results
	for _, sslData := range result.Data {

		// Sanitize some result strings that might contain invalid UTF8-sequences
		sslData.Vhost = utils.ValidUtf8String(sslData.Vhost)

		// Convert list of supported elliptic curves to string
		supportedECs := "id:name"
		for _, curve := range sslData.Curves.SupportedCurves {
			supportedECs += fmt.Sprintf("%s%d:%s", DbValueSeparator, curve.OpenSslNid, curve.Name)
		}

		// Convert list of rejected elliptic curves to string
		rejectedECs := "id:name"
		for _, curve := range sslData.Curves.RejectedCurves {
			rejectedECs += fmt.Sprintf("%s%d:%s", DbValueSeparator, curve.OpenSslNid, curve.Name)
		}

		// Prepare cipher suites results
		for _, cipher := range sslData.Ciphers {

			// Prepare SSL cipher data
			resultCiphers = append(resultCiphers, managerdb.T_ssl_cipher{
				IdTDiscoveryService:     idTDiscoveryService,
				IdTSsl:                  tSslInfoId,
				Vhost:                   sslData.Vhost,
				CipherId:                cipher.Id,
				IanaName:                cipher.IanaName,
				OpensslName:             cipher.OpensslName,
				SupportsECDHKEyExchange: sslData.Curves.SupportEcdhKeyExchange,
				SupportedEllipticCurves: supportedECs,
				RejectedEllipticCurves:  rejectedECs,
				ProtocolVersion:         cipher.Protocol.String(),
				KeyExchange:             cipher.KeyExchange.String(),
				KeyExchangeBits:         cipher.KeyExchangeBits,
				KeyExchangeStrength:     cipher.KeyExchangeStrength,
				KeyExchangeInfo:         strings.Join(cipher.KeyExchangeInfo, DbValueSeparator),
				ForwardSecrecy:          cipher.ForwardSecrecy,
				Authentication:          cipher.Authentication.String(),
				Encryption:              cipher.Encryption.String(),
				EncryptionMode:          cipher.EncryptionMode.String(),
				EncryptionBits:          cipher.EncryptionBits,
				EncryptionStrength:      cipher.EncryptionStrength,
				BlockCipher:             cipher.BlockCipher,
				BlockSize:               cipher.BlockSize,
				StreamCipher:            cipher.StreamCipher,
				Mac:                     cipher.Mac.String(),
				MacBits:                 cipher.MacBits,
				MacStrength:             cipher.MacStrength,
				Prf:                     cipher.Prf.String(),
				PrfBits:                 cipher.PrfBits,
				PrfStrength:             cipher.PrfStrength,
				Export:                  cipher.Export,
				Draft:                   cipher.Draft,
			})
		}

		// Prepare certificates results
		for i, certificateChain := range sslData.Chains {

			// Concat and sanitize some result strings that might contain invalid UTF8-sequences
			validatedBy := utils.ValidUtf8String(strings.Join(certificateChain.ValidatedBy, DbValueSeparator))

			// Prepare certificate data
			for _, certificate := range certificateChain.Certificates {

				// Sanitize some result strings that might contain invalid UTF8-sequences
				certificate.SubjectCN = utils.ValidUtf8String(certificate.SubjectCN)
				certificate.IssuerCN = utils.ValidUtf8String(certificate.IssuerCN)

				// Concat and sanitize some result strings that might contain invalid UTF8-sequences
				subject := utils.ValidUtf8String(strings.Join(certificate.Subject, DbValueSeparator))
				issuer := utils.ValidUtf8String(strings.Join(certificate.Issuer, DbValueSeparator))
				alternativeNames := utils.ValidUtf8String(strings.Join(certificate.AlternativeNames, DbValueSeparator))
				crlUrls := utils.ValidUtf8String(strings.Join(certificate.CrlUrls, DbValueSeparator))
				ocspUrls := utils.ValidUtf8String(strings.Join(certificate.OcspUrls, DbValueSeparator))

				// Prepare certificate data
				resultCertificates = append(resultCertificates, managerdb.T_ssl_certificate{
					IdTDiscoveryService:    idTDiscoveryService,
					IdTSsl:                 tSslInfoId,
					Vhost:                  sslData.Vhost,
					DeploymentId:           uint64(i),
					Type:                   certificate.Type,
					Version:                certificate.Version,
					Serial:                 certificate.Serial.String(),
					ValidChain:             len(certificateChain.ValidatedBy) > 0,
					ChainValidatedBy:       validatedBy,
					ValidChainOrder:        certificateChain.HasValidOrder,
					Subject:                subject,
					SubjectCN:              certificate.SubjectCN,
					Issuer:                 issuer,
					IssuerCN:               certificate.IssuerCN,
					AlternativeNames:       alternativeNames,
					ValidFrom:              sql.NullTime{Time: certificate.ValidFrom, Valid: certificate.ValidFrom != time.Time{}},
					ValidTo:                sql.NullTime{Time: certificate.ValidTo, Valid: certificate.ValidTo != time.Time{}},
					PublicKeyAlgorithm:     certificate.PublicKeyAlgorithm.String(),
					PublicKeyInfo:          certificate.PublicKeyInfo,
					PublicKeyBits:          certificate.PublicKeyBits,
					PublicKeyStrength:      certificate.PublicKeyStrength,
					SignatureAlgorithm:     certificate.SignatureAlgorithm.String(),
					SignatureHashAlgorithm: certificate.SignatureHashAlgorithm.String(),
					CrlUrls:                crlUrls,
					OcspUrls:               ocspUrls,
					KeyUsage:               strings.Join(certificate.KeyUsage, DbValueSeparator),
					ExtendedKeyUsage:       strings.Join(certificate.ExtendedKeyUsage, DbValueSeparator),
					BasicConstraintsValid:  certificate.BasicConstraintsValid,
					Ca:                     certificate.Ca,
					MaxPathLength:          certificate.MaxPathLength,
					Sha1Fingerprint:        certificate.Sha1Fingerprint,
				})
			}
		}

		// Prepare SSL settings results
		sslSettings := sslData.Settings // Temporary reference for shorter code
		resultSettings = append(resultSettings, managerdb.T_ssl_setting{
			IdTDiscoveryService:          idTDiscoveryService,
			IdTSsl:                       tSslInfoId,
			Vhost:                        sslData.Vhost,
			LowestProtocol:               sslSettings.LowestProtocol.String(),
			MinStrength:                  sslSettings.MinStrength,
			Ems:                          sslSettings.Ems,
			TlsFallbackScsv:              sslSettings.TlsFallbackScsv,
			SecureRenegotiation:          sslSettings.SecureRenegotiation,
			SessionResumptionWithId:      sslSettings.SessionResumptionWithId,
			SessionResumptionWithTickets: sslSettings.SessionResumptionWithTickets,
			IsCompliantToMozillaConfig:   sslSettings.IsCompliantToMozillaConfig,
		})

		// Prepare SSL issues results
		sslIssues := sslData.Issues // Temporary reference for shorter code
		resultIssues = append(resultIssues, managerdb.T_ssl_issue{
			IdTDiscoveryService:     idTDiscoveryService,
			IdTSsl:                  tSslInfoId,
			Vhost:                   sslData.Vhost,
			AnyChainInvalid:         sslIssues.AnyChainInvalid,
			AnyChainInvalidOrder:    sslIssues.AnyChainInvalidOrder,
			ClientRenegotiationDos:  sslIssues.ClientRenegotiationDos,
			LowEncryptionStrength:   sslIssues.LowEncryptionStrength,
			ExportSuite:             sslIssues.ExportSuite,
			DraftSuite:              sslIssues.DraftSuite,
			Rc4Enabled:              sslIssues.Rc4Enabled,
			Md2Enabled:              sslIssues.Md2Enabled,
			Md5Enabled:              sslIssues.Md5Enabled,
			Sha1Enabled:             sslIssues.Sha1Enabled,
			Sslv2Enabled:            sslIssues.Sslv2Enabled,
			Sslv3Enabled:            sslIssues.Sslv3Enabled,
			Tlsv1_0Enabled:          sslIssues.Tlsv1_0Enabled,
			Tlsv1_1Enabled:          sslIssues.Tlsv1_1Enabled,
			NoPerfectForwardSecrecy: sslIssues.NoPerfectForwardSecrecy,
			Crime:                   sslIssues.Crime,
			CcsInjection:            sslIssues.CcsInjection,
			EarlyDataSupported:      sslIssues.EarlyDataSupported,
			Beast:                   sslIssues.Beast,
			Heartbleed:              sslIssues.Heartbleed,
			Lucky13:                 sslIssues.Lucky13,
			Poodle:                  sslIssues.Poodle,
			Freak:                   sslIssues.Freak,
			Logjam:                  sslIssues.Logjam,
			Sweet32:                 sslIssues.Sweet32,
			Drown:                   sslIssues.Drown,
			Robot:                   sslIssues.Robot,
		})
	}

	// Insert data into db. Order does NOT matter, as there are no references between the three tables. Empty slices
	// would result in an error, and we don't consider an empty slice an error, as there might simply be no results.
	if len(resultCiphers) > 0 {

		// Use a new gorm session and force a limit on how many Entries can be batched, as we otherwise might
		// exceed the database's limit of 65535 parameters
		errDb := txScopeDb.
			Session(&gorm.Session{CreateBatchSize: managerdb.MaxBatchSizeSslCipher}).
			Create(&resultCiphers).Error
		if errDb != nil {
			return errDb
		}
	}
	if len(resultCertificates) > 0 {
		// Use a new gorm session and force a limit on how many Entries can be batched, as we otherwise might
		// exceed the database's limit of 65535 parameters
		errDb := txScopeDb.
			Session(&gorm.Session{CreateBatchSize: managerdb.MaxBatchSizeSslCertificate}).
			Create(&resultCertificates).Error
		if errDb != nil {
			return errDb
		}
	}
	if len(resultSettings) > 0 {

		// Use a new gorm session and force a limit on how many Entries can be batched, as we otherwise might
		// exceed the database's limit of 65535 parameters
		errDb := txScopeDb.
			Session(&gorm.Session{CreateBatchSize: managerdb.MaxBatchSizeSslSetting}).
			Create(&resultSettings).Error
		if errDb != nil {
			return errDb
		}
	}
	if len(resultIssues) > 0 {

		// Use a new gorm session and force a limit on how many Entries can be batched, as we otherwise might
		// exceed the database's limit of 65535 parameters
		errDb := txScopeDb.
			Session(&gorm.Session{CreateBatchSize: managerdb.MaxBatchSizeSslIssue}).
			Create(&resultIssues).Error
		if errDb != nil {
			return errDb
		}
	}

	// Return nil as everything went fine
	return nil
}
