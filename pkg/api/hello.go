package api

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math"
	"regexp"
	"time"
)

var (
	UserRegex           = regexp.MustCompile("^[a-zA-Z]+$")
	NotBirthdayTemplate = "Hello, %s! Your birthday is in %d day(s)"
	BirthdayTemplate    = "Hello, %s! Happy birthday!"
)

type DateOfBirth struct {
	Username    string `omitempty,json:"username"`
	DateOfBirth string `json:"dateOfBirth"`
	TilBirth    int    `omitempty,json:"tilBirth"`
}

type BirthdayMessage struct {
	Message string `json:"message"`
}

func IsUsernameValid(username string) bool {
	return UserRegex.MatchString(username)
}

// daysTilBirth computes the days that are between today and dateOfBirth of the instance.
// returns 0 if it's today. the amount of days, otherwise.
func (d *DateOfBirth) daysTilBirth() int {
	dt, _ := time.Parse("2006-01-02", d.DateOfBirth)

	t := time.Now()
	todayYear, todayMonth, todayDay := t.Date()
	today := time.Date(todayYear, todayMonth, todayDay, 0, 0, 0, 0, time.UTC)

	_, dateMonth, dateDay := dt.Date()
	date := time.Date(todayYear, dateMonth, dateDay, 0, 0, 0, 0, time.UTC)

	difference := today.Sub(date)
	log.Errorf("difference: %v", difference)
	result := 0.0

	if dateMonth == todayMonth && dateDay == todayDay {
		result = 0.0
	} else if difference < 0 {
		result = math.Abs(math.Ceil(difference.Hours() / 24))
	} else {
		futDate := time.Date(todayYear+1, dateMonth, dateDay, 0, 0, 0, 0, time.UTC)
		futDifference := futDate.Sub(today)
		result = math.Ceil(futDifference.Hours() / 24)
	}
	return int(result)

}

func (d *DateOfBirth) Get() (*BirthdayMessage, error) {
	if d.Username == "" {
		return nil, errors.New("username is empty")
	}
	// TODO: Get from database
	d.DateOfBirth = "2023-03-04"
	d.TilBirth = d.daysTilBirth()

	if d.TilBirth == 0 {
		return &BirthdayMessage{Message: fmt.Sprintf(BirthdayTemplate, d.Username)}, nil
	} else {
		return &BirthdayMessage{Message: fmt.Sprintf(NotBirthdayTemplate, d.Username, d.TilBirth)}, nil
	}
}

func (d *DateOfBirth) Put() error {
	if d.DateOfBirth > time.Now().Format("2006-01-02") {
		return fmt.Errorf("date of birth is set in the future")
	}

	if IsUsernameValid(d.Username) {
		return fmt.Errorf("username contains invalid characters or is empty")
	}

	log.Error("Not implemented")
	return errors.New("Not implemented")
}
