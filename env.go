package main

import (
	"os"
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
		name, val, _ := strings.Cut(kv, "=")
		env[name] = val
	}
	return env
}
