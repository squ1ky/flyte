package experiment

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// EnvFile is the parsed project .env: the runner reads ports, credentials and
// timing parameters (RESERVATION_TTL etc.) from the same file docker compose uses.
type EnvFile map[string]string

func LoadEnvFile(path string) (EnvFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open env file: %w", err)
	}
	defer f.Close()

	env := EnvFile{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		env[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"'`)
	}
	return env, scanner.Err()
}

func (e EnvFile) Get(key string) (string, error) {
	v, ok := e[key]
	if !ok || v == "" {
		return "", fmt.Errorf("%s is not set in env file", key)
	}
	return v, nil
}

func (e EnvFile) Duration(key string) (time.Duration, error) {
	v, err := e.Get(key)
	if err != nil {
		return 0, err
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return d, nil
}

// PostgresURL builds a DSN for a service database reached from the host via its external port,
// e.g. prefix "USER_DB" -> USER_DB_USER, USER_DB_PASSWORD, USER_DB_NAME, USER_DB_PORT_EXTERNAL.
func (e EnvFile) PostgresURL(prefix string) (string, error) {
	var parts [4]string
	for i, suffix := range []string{"_USER", "_PASSWORD", "_PORT_EXTERNAL", "_NAME"} {
		v, err := e.Get(prefix + suffix)
		if err != nil {
			return "", err
		}
		parts[i] = v
	}
	return fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", parts[0], parts[1], parts[2], parts[3]), nil
}
