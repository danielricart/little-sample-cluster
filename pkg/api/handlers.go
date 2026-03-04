package api

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type Server struct {
	Logger *log.Logger
}

func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("health handler called")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		log.Error(fmt.Errorf("failed to write health response: %w", err))
	}
}
