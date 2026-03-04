package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
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

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8089"
	}

	logger.WithField("port", port).Info("Server starting")

	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		logger.WithError(err).Fatal("Server failed to start")
	}

}
