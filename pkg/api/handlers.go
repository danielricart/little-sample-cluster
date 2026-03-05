package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"little-sample-cluster/pkg/metrics"
	"net/http"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Server struct {
	Logger      *log.Logger
	Database    *sql.DB
	HelloServer *HelloServer
	Metrics     *metrics.Metrics
}

var (
	HelloRegex = regexp.MustCompile("^/hello/.+$")
)

func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	s.Logger.Debug("health handler called")

	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		s.Logger.Error(fmt.Errorf("failed to write health response: %w", err))
	}
}

func (s *Server) HelloGetHandler(w http.ResponseWriter, r *http.Request) {
	s.Logger.Debug("hello handler called")

	//Failing early is safer. it duplicates the switch statement but we don't
	//know what can come with non-supported methods in terms of path, body, and formats.
	if r.Method != http.MethodGet {
		s.Logger.Error(fmt.Errorf("method %s not allowed", r.Method))
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	if !HelloRegex.MatchString(r.URL.Path) {
		s.Logger.Error(fmt.Errorf("invalid path: %s", r.URL.Path))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	username := r.PathValue("username")
	if !IsUsernameValid(username) {
		s.Logger.WithFields(log.Fields{"username": username}).Error("username contains invalid characters or is empty")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	user := DateOfBirth{Username: username}
	birthdayMessage, err := s.HelloServer.Get(&user)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		if s.Metrics != nil {
			s.Metrics.InvalidQueries.Add(1.0)
		}
		return
	}
	if birthdayMessage == nil && err == nil {
		s.Logger.WithFields(log.Fields{"username": username}).Info("username not found")
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		if s.Metrics != nil {
			s.Metrics.InvalidQueries.Add(1.0)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	respBody, err := json.Marshal(birthdayMessage)
	_, err = w.Write([]byte(respBody))
	if err != nil {
		s.Logger.WithFields(log.Fields{"username": username}).Error(fmt.Errorf("failed to write response: %w", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		if s.Metrics != nil {
			s.Metrics.InvalidQueries.Add(1.0)
		}
		if s.Metrics != nil {
			s.Metrics.ValidQueries.Add(1.0)
		}
		return
	}
}

func (s *Server) HelloPutHandler(w http.ResponseWriter, r *http.Request) {
	s.Logger.Debug("hello handler called")

	//Failing early is safer. it duplicates the external routing statement but we don't know about alternative paths or e2e tests
	if r.Method != http.MethodPut {
		s.Logger.Error(fmt.Errorf("method %s not allowed", r.Method))
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	// an additional path validation doesn't harm anyone.
	if !HelloRegex.MatchString(r.URL.Path) {
		s.Logger.Error(fmt.Errorf("invalid path: %s", r.URL.Path))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	username := r.PathValue("username")
	if !IsUsernameValid(username) {
		s.Logger.WithFields(log.Fields{"username": username}).Error("username contains invalid characters or is empty")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var dateOfBirth DateOfBirth
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		s.Logger.Error(fmt.Errorf("failed to read body: %w", err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		if s.Metrics != nil {
			s.Metrics.InvalidQueries.Add(1.0)
		}
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)
	if strings.Compare(string(bodyBytes), username) == 0 {
		s.Logger.WithFields(log.Fields{"username": username}).Error("body is empty")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		if s.Metrics != nil {
			s.Metrics.InvalidQueries.Add(1.0)
		}
		return
	}
	err = json.Unmarshal(bodyBytes, &dateOfBirth)
	if err != nil {
		s.Logger.Error(fmt.Errorf("failed to unmarshal body: %w", err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		if s.Metrics != nil {
			s.Metrics.InvalidQueries.Add(1.0)
		}
		return
	}
	dateOfBirth.Username = username

	err = s.HelloServer.Put(&dateOfBirth)
	if err != nil {
		s.Logger.Error(fmt.Errorf("failed to put date of birth: %w", err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		if s.Metrics != nil {
			s.Metrics.InvalidQueries.Add(1.0)
		}
		return
	}
	if s.Metrics != nil {
		s.Metrics.ValidQueries.Add(1.0)
	}
	w.WriteHeader(http.StatusNoContent)
}
