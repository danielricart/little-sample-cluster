package api

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"regexp"
	"time"
)

var (
	UserRegex = regexp.MustCompile("^[a-zA-Z]+$")
)

type DateOfBirth struct {
	Username    string `omitempty,json:"username"`
	DateOfBirth string `json:"dateOfBirth"`
}

func (d *DateOfBirth) Put() error {
	if d.DateOfBirth > time.Now().Format("2006-01-02") {
		return fmt.Errorf("date of birth is set in the future")
	}

	if UserRegex.MatchString(d.Username) {
		return fmt.Errorf("username contains invalid characters or is empty")
	}

	log.Error("Not implemented")
	return errors.New("Not implemented")
}
