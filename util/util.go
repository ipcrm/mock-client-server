package util

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

func EnvString(key string, supplied string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return supplied
}

func EnvInt(key string, supplied int) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			log.Fatalf("%s: %v", key, err)
		}
		return v
	}
	return supplied
}

func EnvFloat64(key string, supplied float64) float64 {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			log.Fatalf("%s: %v", key, err)
		}
		return v
	}
	return supplied
}

func HelpString(str string, envVar string) string {
	return fmt.Sprintf("%s, can be set via the environment with %s.", str, envVar)
}
