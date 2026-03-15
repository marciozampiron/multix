package outbound

// ConfigStore is an interface for reading and writing application configurations.
type ConfigStore interface {
	GetString(key string) string
	GetBool(key string) bool
	GetInt(key string) int
	BindEnv(key string, envVar string) error
	WriteConfig() error
}
