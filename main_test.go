package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	setupLocalDB()
	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestGetHandler(t *testing.T) {
	tt := []struct {
		name        string
		method      string
		target      string
		wantMessage string
		statusCode  int
	}{
		{
			name:        "GET request - success",
			method:      http.MethodGet,
			target:      "/urlinfo/1/google.com",
			wantMessage: `{"message":"URL is not malicious: google.com","status":"OK"}`,
			statusCode:  http.StatusOK,
		},
		{
			name:        "POST request - failure",
			method:      http.MethodPost,
			target:      "/urlinfo/1/google.com",
			wantMessage: `{"message": "Resource not found","status": "NotFound"}`,
			statusCode:  http.StatusNotFound,
		},
		{
			name:        "GET request - forbidden",
			method:      http.MethodGet,
			target:      "/urlinfo/1/badwebsite.com",
			wantMessage: `{"message":"URL is malicious: badwebsite.com","status":"Forbidden"}`,
			statusCode:  http.StatusForbidden,
		},
		{
			name:        "GET request - URL not in database - success",
			method:      http.MethodGet,
			target:      "/urlinfo/1/http://localhost:80/",
			wantMessage: `{"message":"URL is not malicious: http://localhost:80/","status":"OK"}`,
			statusCode:  http.StatusOK,
		},
		{
			name:        "GET request - ftp - success",
			method:      http.MethodGet,
			target:      "/urlinfo/1/ftp://myname@host.dom/%2Fetc/motd",
			wantMessage: `{"message":"URL is not malicious: ftp://myname@host.dom/%2Fetc/motd","status":"OK"}`,
			statusCode:  http.StatusOK,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.target, nil)
			responseRecorder := httptest.NewRecorder()

			wrapperMuxHandler(responseRecorder, req)
			if responseRecorder.Result().StatusCode != tc.statusCode {
				t.Errorf("Want status code %d, got %d", tc.statusCode, responseRecorder.Result().StatusCode)
			}
			if strings.TrimSpace(responseRecorder.Body.String()) != tc.wantMessage {
				t.Errorf("Want message: %s, got message: %s", tc.wantMessage, responseRecorder.Body)
			}
		})
	}
}
