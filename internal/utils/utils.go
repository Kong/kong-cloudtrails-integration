package utils

import (
	"os"
	"strconv"
)

func GetEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

func GetEnvInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		v, err := strconv.Atoi(value)
		if err != nil {
			return defaultVal
		}
		return v
	}

	return defaultVal
}
