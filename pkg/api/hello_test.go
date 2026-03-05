package api

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsUsernameValid(t *testing.T) {
	cases := []struct {
		name     string
		username string
		expected bool
	}{
		{"only letters lowercase", "abc", true},
		{"only letters uppercase", "ABC", true},
		{"mixed case", "AbCdEf", true},
		{"contains numbers", "abc123", false},
		{"contains special chars", "abc!", false},
		{"empty string", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, IsUsernameValid(tc.username))
		})
	}
}

func TestDaysTilBirth(t *testing.T) {
	today := time.Now()

	cases := []struct {
		name        string
		dateOfBirth string
		expected    int
	}{
		{
			name:        "birthday is today",
			dateOfBirth: today.Format("2006-01-02"),
			expected:    0,
		},
		{
			name:        "birthday is tomorrow",
			dateOfBirth: today.AddDate(-1, 0, 1).Format("2006-01-02"),
			expected:    1,
		},
		{
			name:        "birthday was yesterday",
			dateOfBirth: today.AddDate(-1, 0, -1).Format("2006-01-02"),
			expected:    364,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := &DateOfBirth{DateOfBirth: tc.dateOfBirth}
			assert.Equal(t, tc.expected, d.daysTilBirth())
		})
	}
}

func TestUsernameFromRequest(t *testing.T) {
	t.Run("reads username from path variables", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/hello/from-path-value", nil)
		req.SetPathValue("username", "from-path-value")

		assert.Equal(t, "from-path-value", usernameFromRequest(req))
	})

	t.Run("falls back to URL path when not mux-routed", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/hello/from-url", nil)

		assert.Equal(t, "from-url", usernameFromRequest(req))
	})

	t.Run("rejects nested paths", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/hello/user/extra", nil)

		assert.Equal(t, "", usernameFromRequest(req))
	})
}
