package internal

type Config struct {
	BaseUrl     string `env:"SLINKY_BASE_URL"`
	ListenAddr  string `env:"SLINKY_LISTEN_ADDR" default:":8080"`
	DatabaseUri string `env:"SLINKY_DATABASE_URI" default:"postgresql://localhost:5432/slinky?sslmode=disable"`
	LinkSize    int    `env:"SLINKY_LINK_SIZE" default:"6"`
	MaxRetries  int    `env:"SLINKY_MAX_RETRIES" default:"3"`
}
