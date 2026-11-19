package config

type Config struct {
	ADDR string
}

func Load() Config {
	return Config{
		ADDR: GetEnv("ADDR", "8080"),
	}
}
