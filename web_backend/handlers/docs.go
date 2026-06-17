/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2026.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/siemens/Large-Scale-Discovery/web_backend/docs"
)

// ApiDocs returns the generated OpenAPI/Swagger JSON specification for the API docs endpoint.
var ApiDocs = func() gin.HandlerFunc {

	// Return request handling function
	return func(ctx *gin.Context) {
		ctx.Data(200, "application/json; charset=utf-8", []byte(docs.SwaggerInfo.ReadDoc()))
	}
}
