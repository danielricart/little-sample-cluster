package e2eTest

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"little-sample-cluster/pkg/api"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
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

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.HealthHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}
