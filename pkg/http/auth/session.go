package auth

import (
	"context"
	"database/sql"
	"time"

	"github.com/curious-kitten/scratch-post/internal/logger"
	"github.com/segmentio/ksuid"
)

func NewSessionID() string {
	return ksuid.New().String()
}

type Session struct {
	db  *sql.DB
	log logger.Logger
}

// NewSessionHandler creates a structure to handle session authentication
func NewSessionHandler(db *sql.DB, log logger.Logger) *Session {
	return &Session{
		db:  db,
		log: log,
	}
}

// GenerateSecurityString creates a session id for the provided username
func (s *Session) GenerateSecurityString(username string) (string, time.Time, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	sessionID := NewSessionID()
	ctx, close := context.WithTimeout(context.Background(), 5*time.Second)
	defer close()
	stmt := "INSERT INTO sessions (username, sessionid, expirationTime) values ($1, $2, $3)"

	_, err := s.db.ExecContext(ctx, stmt, username, sessionID, expirationTime)
	if err != nil {
		s.log.Debugw("error exeuting query", "query", stmt, "err", err)
		return "", time.Time{}, err
	}
	return sessionID, expirationTime, nil
}

// Validate checks if the provided session id is valid
func (s *Session) Validate(key string) (bool, string, error) {
	ctx, close := context.WithTimeout(context.Background(), 5*time.Second)
	defer close()
	stmt := "select username, expirationtime from sessions WHERE sessionid=$1"
	var username string
	var expTime time.Time
	row := s.db.QueryRowContext(ctx, stmt, key)
	if row.Err() != nil {
		return false, "", row.Err()
	}
	if err := row.Scan(&username, &expTime); err != nil {
		if err == sql.ErrNoRows {
			return false, "", nil
		}
		return false, "", err
	}
	if time.Now().After(expTime) {
		return false, "", nil
	}
	return true, username, nil
}

// Invalidate removes a session ID from the list of accepted session IDs
func (s *Session) Invalidate(session string) error {
	ctx, close := context.WithTimeout(context.Background(), 5*time.Second)
	defer close()
	stmt := "DELETE FROM sessions WHERE sessionid = $1;"
	_, err := s.db.ExecContext(ctx, stmt, session)
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) clearExpiredSessions() {
	ctx, close := context.WithTimeout(context.Background(), 5*time.Second)
	defer close()
	stmt := "DELETE FROM session WHERE expirationTime < $1;"
	_, err := s.db.ExecContext(ctx, stmt, time.Now())
	s.logIfErr(err)
	if err != nil {
		s.logIfErr(err)
	}
	s.logIfErr(err)
}

// Cleanup cleans expired sessions from the store
func (s *Session) Cleanup(cleanInterval time.Duration) {
	go func() {
		timer := time.NewTimer(cleanInterval)
		for range timer.C {
			s.clearExpiredSessions()
		}
	}()
}

func (s *Session) logIfErr(err error) {
	if err != nil {
		s.log.Debugf("not fatal error detected: %v", err.Error())
	}
}
