package main

import (
	"os"
	"regexp"
	"strings"
)

func setDefaultEnv(defaultEnv map[string]string) {
	for k, v := range defaultEnv {
		if _, ok := os.LookupEnv(k); !ok {
			if err := os.Setenv(k, v); err != nil {
				fatalf("Failed to set environment: %s.", err)
			}
		}
	}
}

func getEnv() map[string]string {
	env := make(map[string]string)
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		env[parts[0]] = parts[1]
	}
	return env
}

func getWaitsFromEnv() *urlsFlag {
	result := new(urlsFlag)

	re, _ := regexp.Compile("^DOCKERIZE_WAIT[0-9]+=.+$")

	for _, kv := range os.Environ() {
		if re.MatchString(kv) {
			parts := strings.SplitN(kv, "=", 2)
			if err := result.Set(parts[1]); err != nil {
				os.Exit(1)
			}
		}
	}

	return result
}
