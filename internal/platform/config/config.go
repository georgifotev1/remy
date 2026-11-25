package config

type Config struct {
	ADDR string
	DB   dbConfig
}

type dbConfig struct {
	URL string
}

func Load() Config {
	return Config{
		ADDR: GetEnv("ADDR", "8080"),
		DB: dbConfig{
			URL: GetEnv("DATABASE_URL", "postgres://remy:remy@localhost:5432/remy?sslmode=disable"),
		},
	}
}
