package db

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestLoadConfigFromEnv(t *testing.T) {
	cases := []struct {
		name     string
		env      map[string]string
		expected Config
	}{
		{
			name: "custom values from environment",
			env: map[string]string{
				"DB_HOST":     "testhost",
				"DB_USERNAME": "testuser",
				"DB_PASSWORD": "testpass",
				"DB_NAME":     "testdb",
				"DB_PORT":     "3307",
			},
			expected: Config{Host: "testhost", Username: "testuser", Password: "testpass", Database: "testdb", Port: "3307"},
		},
		{
			name:     "defaults when no environment variables set",
			env:      map[string]string{},
			expected: Config{Host: "localhost", Username: "root", Password: "", Database: "", Port: "3306"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				assert.NoError(t, os.Setenv(k, v))
			}
			t.Cleanup(func() {
				for k := range tc.env {
					assert.NoError(t, os.Unsetenv(k))
				}
			})

			config := LoadConfigFromEnv()

			assert.Equal(t, tc.expected.Host, config.Host)
			assert.Equal(t, tc.expected.Username, config.Username)
			assert.Equal(t, tc.expected.Password, config.Password)
			assert.Equal(t, tc.expected.Database, config.Database)
			assert.Equal(t, tc.expected.Port, config.Port)
		})
	}
}
