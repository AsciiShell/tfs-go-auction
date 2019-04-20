package environment

import (
	"os"
	"strconv"
	"time"
)

func GetInt(key string, def int) int {
	result, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return def
	}
	return result
}

func GetStr(key string, def string) string {
	result := os.Getenv(key)
	if result == "" {
		return def
	}
	return result
}

func GetDuration(key string, def time.Duration) time.Duration {
	result, err := time.ParseDuration(os.Getenv(key))
	if err != nil {
		return def
	}
	return result
}
func GetBool(key string, def bool) bool {
	result, err := strconv.ParseBool(os.Getenv(key))
	if err != nil {
		return def
	}
	return result
}
