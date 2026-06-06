package config

type Config struct {
	Whitelist map[string]bool
}

func New() *Config {
	return &Config{
		Whitelist: make(map[string]bool),
	}
}

func (c *Config) IsWhitelisted(moduleName string) bool {
	return c.Whitelist[moduleName]
}
