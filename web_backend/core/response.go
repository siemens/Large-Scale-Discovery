/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package core

import (
	"github.com/gin-gonic/gin"
	"github.com/siemens/Large-Scale-Discovery/web_backend/config"
	"net/http"
	"time"
)

type Token struct {
	AccessToken string    `json:"access_token"`
	Expire      time.Time `json:"expire"`
}

type BaseResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Body    interface{} `json:"body"`
	Token   *Token      `json:"token"`
}

// Respond responds with 200 OK, the base response data and optional given response body data. The client might always
// look out for error flag and message.
func Respond(context *gin.Context, error bool, message string, body interface{}) {

	// Get user from context storage
	user := GetContextUser(context)

	// Get config
	conf := config.GetConfig()

	// Prepare response body
	resp := BaseResponse{
		Error:   error,
		Message: message,
		Body:    body,
	}

	// Attach fresh JWT token to response if possible
	if user != nil {

		// Generate fresh JWT token if hard limit wasn't reached yet
		threshold := time.Now().Add(-conf.Jwt.Refresh)
		if user.LastLogin.After(threshold) {

			// Generate JWT token and attach to response
			accessToken, expires := createJwt(user.Id, user.LogoutCount)
			resp.Token = &Token{
				AccessToken: accessToken,
				Expire:      expires,
			}
		}
	}

	// Send JSON response
	context.JSON(http.StatusOK, resp)
}

// RespondAuthError returns a 401 UNAUTHORIZED http error code and should be used whenever there is an
// authentication or authorization issue. Additionally, the response body will contain an error flag and a generic
// error message, which may be used by the client. Don't forget to abort further processing after calling this!
func RespondAuthError(context *gin.Context) {
	respondError(context, http.StatusUnauthorized, "Unauthorized")
}

// RespondInternalError returns a 400 BAD REQUEST http error code and should be used whenever there is an internal
// issue. Additionally, the response body will contain an error flag and a generic error message, which may be used
// by the client. Don't forget to issue a critical log message first and abort further processing afterwards!
func RespondInternalError(context *gin.Context) {
	respondError(context, http.StatusBadRequest, "Bad Request")
}

// RespondTemporaryError returns a 200 OK status code but an error flag together with an error message. This return
// code must be used if there is a temporary issue, e.g. some component is down.
func RespondTemporaryError(context *gin.Context) {
	respondError(context, http.StatusServiceUnavailable, "Temporarily Unavailable")
}

//
// DO NOT USE BELOW FUNCTIONS. In order to ensure a unified response behaviour across the complete backend, use the
// predefined public functions from above. They will also ensure that, in case of an error, only generic information
// is returned to the client. If necessary, feel free to add a new public function for a new kind of response.
//

// respondError returns a custom http error code. Additionally, the response body will contain an error flag and a
// generic error message, which may be used by the client.
func respondError(context *gin.Context, errorCode int, errorMessage string) {

	// Prepare error message
	resp := BaseResponse{
		Error:   true,
		Message: errorMessage,
	}

	// Build and return error
	context.JSON(errorCode, resp)

	// Response built, abort further handling
	context.Abort()
}
