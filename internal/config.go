package internal

type Config struct {
	BaseUrl     string `env:"SLINKY_BASE_URL"`
	ListenAddr  string `env:"SLINKY_LISTEN_ADDR" default:":8080"`
	DatabaseUri string `env:"SLINKY_DATABASE_URI" default:"postgresql://localhost:5432/slinky?sslmode=disable"`
	LinkSize    int    `env:"SLINKY_LINK_SIZE" default:"6"`
	MaxRetries  int    `env:"SLINKY_MAX_RETRIES" default:"3"`
	Discord     struct {
		PoweredBy string `env:"SLINKY_DISCORD_POWERED_BY" default:"Powered by slinky 🔗"`
		Color     string `env:"SLINKY_DISCORD_EMBED_COLOR" default:"#ef6a9a"`
	}
	RequireAuth bool   `env:"SLINKY_REQUIRE_AUTH" default:"true"`
	HmacSecret  string `env:"SLINKY_JWT_HMAC_SECRET"`
}
