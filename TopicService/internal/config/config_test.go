package config

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Сохраняем оригинальные переменные окружения
	originalEnv := os.Environ()
	defer func() {
		// Восстанавливаем окружение после теста
		os.Clearenv()
		for _, env := range originalEnv {
			keyVal := strings.SplitN(env, "=", 2)
			if len(keyVal) == 2 {
				os.Setenv(keyVal[0], keyVal[1])
			}
		}
	}()

	tests := []struct {
		name     string
		setup    func()
		expected *Config
	}{
		{
			name: "default values",
			setup: func() {
				os.Clearenv()
			},
			expected: &Config{
				AppEnv:      "development",
				ServerPort:  "8080",
				AuthService: "",
				DBHost:      "localhost",
				DBPort:      "5432",
				DBUser:      "postgres",
				DBPassword:  "",
				DBName:      "",
				DBSSLMode:   "disable",
				LogLevel:    "info",
				LogFormat:   "text",
				LogOutput:   "",
			},
		},
		{
			name: "custom values from env",
			setup: func() {
				os.Clearenv()
				os.Setenv("APP_ENV", "production")
				os.Setenv("SERVER_PORT", "3000")
				os.Setenv("DB_HOST", "db.example.com")
				os.Setenv("DB_NAME", "testdb")
				os.Setenv("LOG_LEVEL", "debug")
				os.Setenv("LOG_FORMAT", "json")
				os.Setenv("LOG_COLOR", "false")
			},
			expected: &Config{
				AppEnv:      "production",
				ServerPort:  "3000",
				AuthService: "",
				DBHost:      "db.example.com",
				DBPort:      "5432",
				DBUser:      "postgres",
				DBPassword:  "",
				DBName:      "testdb",
				DBSSLMode:   "disable",
				LogLevel:    "debug",
				LogFormat:   "json",
				LogOutput:   "",
			},
		},
		{
			name: "partial custom values",
			setup: func() {
				os.Clearenv()
				os.Setenv("DB_USER", "custom_user")
				os.Setenv("DB_PASSWORD", "secret")
				os.Setenv("LOG_OUTPUT", "file.log")
			},
			expected: &Config{
				AppEnv:      "development",
				ServerPort:  "8080",
				AuthService: "",
				DBHost:      "localhost",
				DBPort:      "5432",
				DBUser:      "custom_user",
				DBPassword:  "secret",
				DBName:      "",
				DBSSLMode:   "disable",
				LogLevel:    "info",
				LogFormat:   "text",
				LogOutput:   "file.log",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			cfg, err := Load()
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, cfg)
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		defaultValue  string
		setup         func()
		expectedValue string
	}{
		{
			name:          "env var exists",
			key:           "TEST_KEY",
			defaultValue:  "default",
			setup:         func() { os.Setenv("TEST_KEY", "custom") },
			expectedValue: "custom",
		},
		{
			name:          "env var not exists",
			key:           "NON_EXISTENT_KEY",
			defaultValue:  "default",
			setup:         func() {},
			expectedValue: "default",
		},
		{
			name:          "empty env var",
			key:           "EMPTY_KEY",
			defaultValue:  "default",
			setup:         func() { os.Setenv("EMPTY_KEY", "") },
			expectedValue: "",
		},
		{
			name:          "boolean env var",
			key:           "BOOL_KEY",
			defaultValue:  "true",
			setup:         func() { os.Setenv("BOOL_KEY", "false") },
			expectedValue: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Сохраняем и восстанавливаем окружение
			originalValue, exists := os.LookupEnv(tt.key)
			defer func() {
				if exists {
					os.Setenv(tt.key, originalValue)
				} else {
					os.Unsetenv(tt.key)
				}
			}()

			tt.setup()
			value := getEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expectedValue, value)
		})
	}
}
