package env

import (
	"os"
	"strconv"
	"strings"
)

func envClean(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func String(key, defaultValue string) string {
	envValue := envClean(key)
	if envValue == "" {
		return defaultValue
	}
	return envValue
}

func Int(key string, defaultValue int) int {
	envValue := envClean(key)
	if envValue == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(envValue)
	if err != nil {
		return defaultValue
	}
	return intValue
}

func Float(key string, defaultValue float64) float64 {
	envValue := envClean(key)
	if envValue == "" {
		return defaultValue
	}
	v, err := strconv.ParseFloat(envValue, 64)
	if err != nil {
		return defaultValue
	}
	return v
}

func Array(key, separator string, defaultValue []string) []string {
	envValue := envClean(key)
	if envValue == "" {
		return defaultValue
	}
	return strings.Split(envValue, separator)
}
