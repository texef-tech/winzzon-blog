package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Port    string
	Env     string

	DatabaseURL string
	RedisURL    string

	JWTSecret      string
	JWTExpiryHours int

	AdminUsername string
	AdminPassword string

	S3Endpoint  string
	S3Bucket    string
	S3AccessKey string
	S3SecretKey string
	S3CDNURL    string

	MeiliURL    string
	MeiliAPIKey string

	AllowedOrigins string
	SiteURL        string
	SiteName       string
}

func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

func Load() *Config {
	_ = godotenv.Load()

	viper.AutomaticEnv()

	viper.SetDefault("PORT", "8080")
	viper.SetDefault("ENV", "development")
	viper.SetDefault("JWT_EXPIRY_HOURS", 24)

	cfg := &Config{
		Port: viper.GetString("PORT"),
		Env:  viper.GetString("ENV"),

		DatabaseURL: viper.GetString("DATABASE_URL"),
		RedisURL:    viper.GetString("REDIS_URL"),

		JWTSecret:      viper.GetString("JWT_SECRET"),
		JWTExpiryHours: viper.GetInt("JWT_EXPIRY_HOURS"),

		AdminUsername: viper.GetString("ADMIN_USERNAME"),
		AdminPassword: viper.GetString("ADMIN_PASSWORD"),

		S3Endpoint:  viper.GetString("S3_ENDPOINT"),
		S3Bucket:    viper.GetString("S3_BUCKET"),
		S3AccessKey: viper.GetString("S3_ACCESS_KEY"),
		S3SecretKey: viper.GetString("S3_SECRET_KEY"),
		S3CDNURL:    viper.GetString("S3_CDN_URL"),

		MeiliURL:    viper.GetString("MEILI_URL"),
		MeiliAPIKey: viper.GetString("MEILI_API_KEY"),

		AllowedOrigins: viper.GetString("ALLOWED_ORIGINS"),
		SiteURL:        viper.GetString("SITE_URL"),
		SiteName:       viper.GetString("SITE_NAME"),
	}

	required := map[string]string{
		"DATABASE_URL": cfg.DatabaseURL,
		"REDIS_URL":    cfg.RedisURL,
		"JWT_SECRET":   cfg.JWTSecret,
		"SITE_URL":     cfg.SiteURL,
		"SITE_NAME":    cfg.SiteName,
	}
	for key, val := range required {
		if val == "" {
			log.Fatalf("required environment variable %s is not set", key)
		}
	}

	return cfg
}
