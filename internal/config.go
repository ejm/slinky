package internal

type Config struct {
	DatabaseUrl string `env:"SLINKY_DATABASE_URL" default:"postgresql://localhost:5432/slinky?sslmode=disable"`
	LinkSize    int    `env:"SLINKY_LINK_SIZE" default:"6"`
	MaxRetries  int    `env:"SLINKY_MAX_RETRIES" default:"3"`
}
