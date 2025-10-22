package config

import (
	"errors"
	"os"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from a .env file and parses them into a struct of type T.
// The function uses godotenv to load the .env file and env package to parse environment variables.
//
// Parameters:
//   - path: The path to the .env file to load
//   - T: The type parameter representing the struct type to parse environment variables into
//
// Returns:
//   - *T: A pointer to the parsed configuration struct
//   - error: An error if loading the .env file or parsing environment variables fails
//
// Example:
//
//	type Config struct {
//	    DatabaseURL string `env:"DATABASE_URL"`
//	    Port        int    `env:"PORT"`
//	}
//
//	config, err := LoadEnv[Config](".env")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// NOTE: This will override system env with .env file if it exists
func LoadEnv[T any](path string) (T, error) {
	var envNotFound error
	if err := godotenv.Load(path); err != nil {
		envNotFound = err
	}
	_ = envNotFound
	var cfg T
	if err := env.ParseWithOptions(&cfg, env.Options{
		RequiredIfNoDef: true,
		Environment:     env.ToMap(os.Environ()),
	}); err != nil {
		return cfg, errors.Join(envNotFound, err)
	}
	return cfg, nil
}
