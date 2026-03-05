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

	//Failing early is safer. it duplicates the switch statement but we don't
	//know what can come with non-syupported methods in terms of path, body, and formats.
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
	if !IsUsernameValid(username) {
		log.WithFields(log.Fields{"username": username}).Error("username contains invalid characters or is empty")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		user := DateOfBirth{Username: username}
		birthdayMessage, err := s.HelloServer.Get(&user)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			s.Metrics.InvalidQueries.Add(1.0)
			return
		}
		if birthdayMessage == nil && err == nil {
			log.WithFields(log.Fields{"username": username}).Info("username not found")
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			s.Metrics.InvalidQueries.Add(1.0)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		respBody, err := json.Marshal(birthdayMessage)
		_, err = w.Write([]byte(respBody))
		if err != nil {
			log.WithFields(log.Fields{"username": username}).Error(fmt.Errorf("failed to write response: %w", err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			s.Metrics.InvalidQueries.Add(1.0)
			return
		}

	case http.MethodPut:
		var dateOfBirth DateOfBirth
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			s.Logger.Error(fmt.Errorf("failed to read body: %w", err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			s.Metrics.InvalidQueries.Add(1.0)
			return
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {

			}
		}(r.Body)
		if strings.Compare(string(bodyBytes), username) == 0 {
			log.WithFields(log.Fields{"username": username}).Error("body is empty")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			s.Metrics.InvalidQueries.Add(1.0)
			return
		}
		err = json.Unmarshal(bodyBytes, &dateOfBirth)
		if err != nil {
			s.Logger.Error(fmt.Errorf("failed to unmarshal body: %w", err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			s.Metrics.InvalidQueries.Add(1.0)
			return
		}
		dateOfBirth.Username = username

		err = s.HelloServer.Put(&dateOfBirth)
		if err != nil {
			s.Logger.Error(fmt.Errorf("failed to put date of birth: %w", err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			s.Metrics.InvalidQueries.Add(1.0)
			return
		}
		s.Metrics.ValidQueries.Add(1.0)
		w.WriteHeader(http.StatusNoContent)

	}
}
