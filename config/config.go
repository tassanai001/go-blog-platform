package config

type Config struct {
	MongoDB struct {
		URI      string
		Database string
	}
	JWT struct {
		Secret string
	}
	Server struct {
		Port string
	}
}

func LoadConfig() *Config {
	// In a real application, you would load this from environment variables
	// or a configuration file
	cfg := &Config{}
	cfg.MongoDB.URI = "mongodb://localhost:27017"
	cfg.MongoDB.Database = "blog_platform"
	cfg.JWT.Secret = "your-256-bit-secret" // In production, use a secure secret and store it in environment variables
	cfg.Server.Port = "8080"
	return cfg
}
