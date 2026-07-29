package utils

import "os"

// RestoreEnv sets key to value, or unsets it if value is empty.
// Intended for use in deferred cleanup after os.Setenv calls in tests.
func RestoreEnv(key, value string) {
	if value == "" {
		os.Unsetenv(key)
	} else {
		os.Setenv(key, value)
	}
}
