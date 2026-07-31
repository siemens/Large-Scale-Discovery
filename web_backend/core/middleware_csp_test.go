/*
 * Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
 *
 * Copyright (c) Siemens AG, 2016-2026.
 *
 * This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
 * directory or visit <https://opensource.org/licenses/MIT>.
 *
 * SPDX-License-Identifier: MIT
 */

package core

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	scanUtils "github.com/siemens/GoScans/utils"
)

// TestMwCspNonce_SameNonceInHeaderAndHTML verifies that the CSP header and SPA shell use the same nonce
func TestMwCspNonce_SameNonceInHeaderAndHTML(t *testing.T) {
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "index.html")
	const templateHTML = `<!doctype html><html><head><meta name="csp-nonce" content="__CSP_NONCE__"></head></html>`
	if errWriteFile := os.WriteFile(indexPath, []byte(templateHTML), 0600); errWriteFile != nil {
		t.Fatalf("write test SPA shell: %v", errWriteFile)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(ctx *gin.Context) {
		SetContextStorage(ctx, &ContextStorage{Logger: scanUtils.NewTestLogger()})
		ctx.Next()
	})
	router.Use(MwSetSecurityHeaders())
	router.NoRoute(func(ctx *gin.Context) {
		content, errContent := os.ReadFile(indexPath)
		if errContent != nil {
			ctx.Status(http.StatusInternalServerError)
			return
		}
		content = []byte(strings.ReplaceAll(string(content), "__CSP_NONCE__", ctx.GetString("csp-nonce")))
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", content)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", recorder.Code)
	}
	nonce := cspNonce(recorder.Header().Get("Content-Security-Policy"))
	if nonce == "" {
		t.Fatal("CSP header did not contain a nonce")
	}
	if strings.Contains(recorder.Body.String(), "__CSP_NONCE__") {
		t.Fatal("SPA shell still contains the CSP nonce placeholder")
	}
	if !strings.Contains(recorder.Body.String(), `content="`+nonce+`"`) {
		t.Fatalf("SPA shell did not receive the CSP header nonce %q", nonce)
	}
}

// TestMwCspNonce_DiffersPerRequestAndUsesEnforcingHeader verifies that enforcing headers use unique per-request nonces
func TestMwCspNonce_DiffersPerRequestAndUsesEnforcingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(ctx *gin.Context) {
		SetContextStorage(ctx, &ContextStorage{Logger: scanUtils.NewTestLogger()})
		ctx.Next()
	})
	router.Use(MwSetSecurityHeaders())
	router.GET("/ping", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })

	requestNonce := func() string {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/ping", nil)
		router.ServeHTTP(recorder, request)
		if reportOnly := recorder.Header().Get("Content-Security-Policy-Report-Only"); reportOnly != "" {
			t.Fatalf("enforced mode emitted report-only CSP: %q", reportOnly)
		}
		return cspNonce(recorder.Header().Get("Content-Security-Policy"))
	}

	first := requestNonce()
	second := requestNonce()
	if first == "" || second == "" {
		t.Fatal("could not extract a nonce from the enforcing CSP headers")
	}
	if first == second {
		t.Fatalf("consecutive requests reused CSP nonce %q", first)
	}
}

// cspNonce extracts the style nonce from a content security policy
func cspNonce(policy string) string {
	const prefix = "'nonce-"
	start := strings.Index(policy, prefix)
	if start < 0 {
		return ""
	}
	start += len(prefix)
	end := strings.Index(policy[start:], "'")
	if end < 0 {
		return ""
	}
	return policy[start : start+end]
}
