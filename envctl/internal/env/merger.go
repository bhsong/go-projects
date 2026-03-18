package env

import (
	"fmt"
	"os"
)

func Merge(base, override map[string]string) map[string]string {
	result := make(map[string]string, len(base))

	for k, v := range base {
		result[k] = v
	}

	for k, v := range override {
		result[k] = v
	}

	return result
}

func Apply(env map[string]string, overrideExisting bool) error {
	for k, v := range env {
		_, exists := os.LookupEnv(k)
		if exists && !overrideExisting {
			continue
		}

		if err := os.Setenv(k, v); err != nil {
			return fmt.Errorf("env.Apply: %s 설정 실패: %w", k, err)
		}
	}

	return nil
}
