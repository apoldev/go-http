package env

import (
	"os"
	"strconv"
)

func LookupEnvStringDefault(key string, defaultValue string) string {
	s := os.Getenv(key)
	if s == "" {
		return defaultValue
	}
	return s
}

func LookupEnvIntDefault(key string, defaultValue int) int {
	s := os.Getenv(key)
	if s == "" {
		return defaultValue
	}
	c, _ := strconv.Atoi(s)
	return c
}
