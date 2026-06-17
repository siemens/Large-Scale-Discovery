/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2023.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	"os"
	"testing"
)

func TestWindowsExportTrustStore(t *testing.T) {

	// Prepare and run test cases
	type args struct {
		outputFile string
		appendFile bool
		version    string
		store      string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "generate-store",
			args:    args{outputFile: "customtest.pem", appendFile: false, version: "gentest", store: "CA"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := windowsExportTrustStore(tt.args.outputFile, tt.args.appendFile, tt.args.version, tt.args.store); (got != nil) != tt.wantErr {
				t.Errorf("windowsExportTrustStore() = '%v', wantErr = '%v'", got, tt.wantErr)
			}
			_ = os.Remove(tt.args.outputFile)
		})
	}
}
