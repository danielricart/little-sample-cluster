package e2eTest

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"little-sample-cluster/pkg/api"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

type E2eTestSuite struct {
	suite.Suite
	ctx context.Context
}

// TestMain sets up and tears down the required dependency environment
func TestMain(m *testing.M) {
	//ctx := context.Background()
	logger := log.New()
	logger.SetLevel(log.DebugLevel)

	// Run tests
	code := m.Run()

	os.Exit(code)
}

func TestHealthIntegration(t *testing.T) {
	server := api.Server{Logger: log.New()}

	cases := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   string
	}{
		{"GET returns 200", http.MethodGet, http.StatusOK, "OK"},
		{"PUT returns 405", http.MethodPut, http.StatusMethodNotAllowed, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/health", nil)
			w := httptest.NewRecorder()

			server.HealthHandler(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, w.Body.String())
			}
		})
	}
}

func TestHelloGetIntegration(t *testing.T) {
	server := api.Server{Logger: log.New()}
	cases := []struct {
		testName       string
		username       string
		method         string
		expectedStatus int
		expectedBody   string
	}{
		{"valid username, in 3 days", "asdasda", http.MethodGet, http.StatusOK, fmt.Sprintf(`{"message":"Hello, %s! Your birthday is in %d day(s)"}`, "asdasda", 3)},
		{"valid username, today", "asdasda", http.MethodGet, http.StatusOK, fmt.Sprintf(`{"message":"Hello, %s! Happy birthday!"}`, "asdasda")},
		{"invalid username pattern", "asd123", http.MethodGet, http.StatusBadRequest, ""},
	}
	
	for _, tc := range cases {
		t.Run(tc.testName, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, fmt.Sprintf("/hello/%s", tc.username), nil)
			w := httptest.NewRecorder()

			server.HelloHandler(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, w.Body.String())
			}
		})
	}
}

func TestHelloPutIntegration(t *testing.T) {
	server := api.Server{Logger: log.New()}
	pastDate := time.Now().AddDate(-1, 0, -1).Format("2006-01-02")
	futureDate := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	cases := []struct {
		testName             string
		username             string
		requestBody          string
		method               string
		expectedStatus       int
		expectedResponseBody string
	}{
		{"valid username missing body", "asdasda", ``, http.MethodPut, http.StatusBadRequest, ""},
		{"valid username, wrong date", "asdasda", fmt.Sprintf(`{"dateOfBirth": "%s"}`, futureDate), http.MethodPut, http.StatusBadRequest, ""},
		{"invalid username pattern", "asd123", fmt.Sprintf(`{"dateOfBirth": "%s"}`, pastDate), http.MethodPut, http.StatusBadRequest, ""},
		{"valid username pattern", "asdasda", fmt.Sprintf(`{"dateOfBirth": "%s"}`, pastDate), http.MethodPut, http.StatusNoContent, ""},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, fmt.Sprintf("/hello/%s", tc.username), strings.NewReader(tc.requestBody))
			w := httptest.NewRecorder()

			server.HelloHandler(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.expectedResponseBody != "" {
				assert.Equal(t, tc.expectedResponseBody, w.Body.String())
			}
		})
	}
}
