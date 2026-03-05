package db

import (
	"database/sql"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
)

type User struct {
	Username    string
	DateOfBirth time.Time
}

type UserRepository interface {
	GetBirthDateByUsername(username string) (*time.Time, error)
	InsertOrUpdateUsernameAndBirthDate(username string, birthDate time.Time) error
}

type UserRepositoryImpl struct {
	db     *sql.DB
	logger *log.Logger
}

func NewUserRepository(db *sql.DB, logger *log.Logger) *UserRepositoryImpl {
	return &UserRepositoryImpl{db: db, logger: logger}
}

func (u *UserRepositoryImpl) GetBirthDateByUsername(username string) (*time.Time, error) {
	// get the most recent user and limit to 1 result.
	//given that there are constraints for usernames being unique, this will return always 1 or 0.
	query := "SELECT username, date_of_birth FROM users WHERE username = ? limit 1"

	user := &User{}

	err := u.db.QueryRow(query, username).Scan(&user.Username, &user.DateOfBirth)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		u.logger.Error(err)
		return nil, err
	}
	return &user.DateOfBirth, nil
}

func (u *UserRepositoryImpl) InsertOrUpdateUsernameAndBirthDate(username string, birthDate time.Time) error {
	query := "INSERT INTO users (username, date_of_birth) VALUES (?, ?) ON DUPLICATE KEY UPDATE date_of_birth = values(date_of_birth)"

	_, err := u.db.Exec(query, username, birthDate)
	if err != nil {
		u.logger.Debug(err)
		return err
	}
	return nil
}
