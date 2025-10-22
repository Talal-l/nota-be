package types

import (
	"github.com/anotik/anocore/pkg/errs"
)

type GenericResponse[T any] struct {
	Data T              `json:"data"`
	Err  *errs.ApiError `json:"err"`
}

type Validator = map[string]string

type AppConfig struct {
	BaseUrl           string `env:"BASE_URL"`
	Port              string `env:"PORT" envDefault:"8080"`
	Host              string `env:"HOST" envDefault:"0.0.0.0"`
	DBName            string `env:"DB_NAME"`
	DBUser            string `env:"DB_USER"`
	DBPassword        string `env:"DB_PASSWORD"`
	DBUserAdmin       string `env:"DB_USER_ADMIN"`
	DBPasswordAdmin   string `env:"DB_PASSWORD_ADMIN"`
	DBPort            string `env:"DB_PORT" envDefault:"1433"`
	DBHost            string `env:"DB_HOST"`
	Mode              string `env:"MODE" envDefault:"dev"`                     // dev, stg, prod
	MaxAttachmentSize int    `env:"MAX_ATTACHMENT_SIZE" envDefault:"35000000"` // in MB
}
