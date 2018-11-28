package main

import (
	"os"
	"strings"
)

func setDefaultEnv(defaultEnv map[string]string) {
	for k, v := range defaultEnv {
		if _, ok := os.LookupEnv(k); !ok {
			os.Setenv(k, v)
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
