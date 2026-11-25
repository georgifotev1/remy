package config

import "os"

func GetEnv(v, fallback string) string {
	str := os.Getenv(v)
	if str == "" {
		return fallback
	}

	return str
}
