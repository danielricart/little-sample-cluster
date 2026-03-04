package api

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type Server struct {
	Logger *log.Logger
}

var (
	HelloRegex = regexp.MustCompile("^/hello/[a-zA-Z]+$")
)

func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("health handler called")

	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		log.Error(fmt.Errorf("failed to write health response: %w", err))
	}
}

func (s *Server) HelloHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("hello handler called")

	if r.Method != http.MethodGet && r.Method != http.MethodPut {
		s.Logger.Error(fmt.Errorf("method %s not allowed", r.Method))
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	if !HelloRegex.MatchString(r.URL.Path) {
		s.Logger.Error(fmt.Errorf("invalid path: %s", r.URL.Path))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	username := strings.Split(r.URL.Path, "/")[2]

	if r.Method == http.MethodPut {
		var dateOfBirth DateOfBirth
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			s.Logger.Error(fmt.Errorf("failed to read body: %w", err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {

			}
		}(r.Body)
		err = json.Unmarshal(bodyBytes, &dateOfBirth)
		if err != nil {
			s.Logger.Error(fmt.Errorf("failed to unmarshal body: %w", err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
		dateOfBirth.Username = username

		err = dateOfBirth.Put()
		if err != nil {
			s.Logger.Error(fmt.Errorf("failed to put date of birth: %w", err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusNoContent)
		
	}
}
