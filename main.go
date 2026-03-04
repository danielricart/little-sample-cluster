package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"little-sample-cluster/pkg/api"
	"net/http"
	"os"
)

func main() {
	// Initialize logger
	logger := log.New()
	logger.SetFormatter(&log.JSONFormatter{})
	logger.SetLevel(log.InfoLevel)
	logger.SetReportCaller(true)

	// Check if DEBUG environment variable is set
	if os.Getenv("DEBUG") == "true" {
		logger.SetLevel(log.TraceLevel)
	}

	logger.Info("Starting Hello Birthday...")

	server := api.Server{
		Logger: logger,
	}

	http.HandleFunc("/health", server.HealthHandler)
	http.HandleFunc("/hello", server.HelloHandler)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8089"
	}

	logger.WithField("port", port).Info("Server starting")

	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		logger.WithError(err).Fatal("Server failed to start")
	}

}
