package config

import "os"

type RuntimeEnvironment string

const (
	RuntimeEnvVar                    = "RUNTIME_ENVIRONMENT"
	Dev           RuntimeEnvironment = "dev"
	Staging       RuntimeEnvironment = "staging"
	Production    RuntimeEnvironment = "production"
)

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetCurrentRuntimeEnvironment() string {
	return GetEnv(RuntimeEnvVar, string(Dev))
}
