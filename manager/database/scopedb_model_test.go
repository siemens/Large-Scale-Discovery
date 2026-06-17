/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package database

import "testing"

// TestTableName_Discovery verifies T_discovery.TableName returns the expected table name.
func TestTableName_Discovery(t *testing.T) {
	if got := (T_discovery{}).TableName(); got != "t_discovery" {
		t.Errorf("T_discovery.TableName() = '%v', want = 't_discovery'", got)
	}
}

// TestTableName_Nfs verifies T_nfs.TableName returns the expected table name.
func TestTableName_Nfs(t *testing.T) {
	if got := (T_nfs{}).TableName(); got != "t_nfs" {
		t.Errorf("T_nfs.TableName() = '%v', want = 't_nfs'", got)
	}
}

// TestTableName_Smb verifies T_smb.TableName returns the expected table name.
func TestTableName_Smb(t *testing.T) {
	if got := (T_smb{}).TableName(); got != "t_smb" {
		t.Errorf("T_smb.TableName() = '%v', want = 't_smb'", got)
	}
}

// TestTableName_Ssh verifies T_ssh.TableName returns the expected table name.
func TestTableName_Ssh(t *testing.T) {
	if got := (T_ssh{}).TableName(); got != "t_ssh" {
		t.Errorf("T_ssh.TableName() = '%v', want = 't_ssh'", got)
	}
}

// TestTableName_Ssl verifies T_ssl.TableName returns the expected table name.
func TestTableName_Ssl(t *testing.T) {
	if got := (T_ssl{}).TableName(); got != "t_ssl" {
		t.Errorf("T_ssl.TableName() = '%v', want = 't_ssl'", got)
	}
}

// TestTableName_Webcrawler verifies T_webcrawler.TableName returns the expected table name.
func TestTableName_Webcrawler(t *testing.T) {
	if got := (T_webcrawler{}).TableName(); got != "t_webcrawler" {
		t.Errorf("T_webcrawler.TableName() = '%v', want = 't_webcrawler'", got)
	}
}

// TestTableName_Webenum verifies T_webenum.TableName returns the expected table name.
func TestTableName_Webenum(t *testing.T) {
	if got := (T_webenum{}).TableName(); got != "t_webenum" {
		t.Errorf("T_webenum.TableName() = '%v', want = 't_webenum'", got)
	}
}

// TestTableName_Nuclei verifies T_nuclei.TableName returns the expected table name.
func TestTableName_Nuclei(t *testing.T) {
	if got := (T_nuclei{}).TableName(); got != "t_nuclei" {
		t.Errorf("T_nuclei.TableName() = '%v', want = 't_nuclei'", got)
	}
}
