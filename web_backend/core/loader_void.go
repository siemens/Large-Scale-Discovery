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
	scanUtils "github.com/siemens/GoScans/utils"
	"large-scale-discovery/web_backend/database"
)

// init automatically registers the loader implemented in this file. If you don't want this loader,
// just remove it. You can also add your own loader by adding a file with your dedicated implementation.
func init() {

	// Register loader for initialization
	loaders = append(loaders, NewLoaderVoid(nil))
}

type LoaderVoid struct {
	domains []string
}

// NewLoaderVoid generates a new loader with the user domains it is responsible for. Everything
// else of the loader will be initialized later during core initialization, with the actual config values.
func NewLoaderVoid(domains []string) *LoaderVoid {
	return &LoaderVoid{
		domains: domains,
	}
}

// Domains returns the user domains this loader got registered for
func (l *LoaderVoid) Domains() []string {
	return l.domains
}

// Init validates loader settings and initializes the loader
func (l *LoaderVoid) Init(conf map[string]interface{}) error {

	// Return nil as everything went fine
	return nil
}

// RefreshUser updates user attributes according to the implemented rules. This may be used to load/update user
// details from a remote repository. Changes are not yet committed! Might return one of FOUR kinds of error:
//     - A temporary error: Indicating a remote connection error. You may continue with cached data or return a
//       temporary error to the user.
//     - An internal error: Indicating an unexpected error. You should not continue, but return a generic error
//       message to the user.
//     - A public error (string): Indicating an error that is relevant information for the user, you may want to
//       return this message back to the user.
//       ATTENTION: If a public error message is returned, it also always comes in tandem with an detailed internal
//       error, which might be useful for additional logging.
func (l *LoaderVoid) RefreshUser(logger scanUtils.Logger, user *database.T_user) (errTemporary error, errInternal error, errPublic string) {

	// Fall back, if necessary, to user-dedicated company to avoid unintended groups
	if len(user.Company) == 0 {
		user.Company = user.Email
	}

	// Return nil as everything went fine
	return nil, nil, ""
}
