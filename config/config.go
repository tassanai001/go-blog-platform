package config

import (
    "os"
)

type Config struct {
    Server   ServerConfig
    MongoDB  MongoDBConfig
    JWT      JWTConfig
    SMTP     SMTPConfig
    BaseURL  string
}

type ServerConfig struct {
    Port string
}

type MongoDBConfig struct {
    URI      string
    Database string
}

type JWTConfig struct {
    Secret string
}

type SMTPConfig struct {
    Host     string
    Port     string
    Username string
    Password string
    FromEmail string
}

func LoadConfig() *Config {
    return &Config{
        Server: ServerConfig{
            Port: getEnvOrDefault("SERVER_PORT", "8080"),
        },
        MongoDB: MongoDBConfig{
            URI:      getEnvOrDefault("MONGODB_URI", "mongodb://localhost:27017"),
            Database: getEnvOrDefault("MONGODB_DATABASE", "blog_platform"),
        },
        JWT: JWTConfig{
            Secret: getEnvOrDefault("JWT_SECRET", "your-256-bit-secret"),
        },
        SMTP: SMTPConfig{
            Host:     getEnvOrDefault("SMTP_HOST", "smtp.gmail.com"),
            Port:     getEnvOrDefault("SMTP_PORT", "587"),
            Username: getEnvOrDefault("SMTP_USERNAME", ""),
            Password: getEnvOrDefault("SMTP_PASSWORD", ""),
            FromEmail: getEnvOrDefault("SMTP_FROM_EMAIL", "noreply@yourblog.com"),
        },
        BaseURL: getEnvOrDefault("BASE_URL", "http://localhost:8080"),
    }
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
