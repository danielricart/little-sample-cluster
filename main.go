package main

import (
	"fmt"
	"little-sample-cluster/pkg/api"
	"little-sample-cluster/pkg/db"
	"little-sample-cluster/pkg/metrics"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
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

	// Load database configuration
	config := db.LoadConfigFromEnv()
	logger.WithFields(log.Fields{
		"host":     config.Host,
		"database": config.Database,
		"port":     config.Port,
	}).Info("Database configuration loaded")

	// Connect to database
	database, err := config.Connect()
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer database.Close()

	logger.Info("Database connection established")

	logger.Info("Starting Hello Birthday...")
	promMetrics, promHandler := metrics.NewMetrics(logger)

	server := api.Server{
		Logger:      logger,
		Database:    database,
		HelloServer: api.NewHelloServer(database, logger),
		Metrics:     promMetrics,
	}

	http.HandleFunc("/health", server.HealthHandler)
	http.HandleFunc("/hello", server.HelloHandler)
	http.Handle("/metrics", *promHandler)
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8089"
	}

	logger.WithField("port", port).Info("Server starting")

	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		logger.WithError(err).Fatal("Server failed to start")
	}

}
