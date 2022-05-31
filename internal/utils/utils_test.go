package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv(t *testing.T) {
	expected := "hello"
	os.Setenv("Test", expected)
	result := GetEnv("Test", "")

	assert.Equal(t, result, expected)
}

func TestGetEnvDefault(t *testing.T) {

	result := GetEnv("TestTwo", "")
	assert.Equal(t, result, "")
}

func TestGetEnvInt(t *testing.T) {
	expected := 1
	os.Setenv("Value", "1")
	result := GetEnvInt("Value", 1)
	assert.Equal(t, result, expected)
}

func TestGetEnvIntDefault(t *testing.T) {
	result := GetEnvInt("ValueTwo", 2)
	assert.Equal(t, result, 2)
}
