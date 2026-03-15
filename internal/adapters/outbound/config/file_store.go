package config

import (
	"path/filepath"

	"multix/internal/ports/outbound"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

type viperAdapter struct {
	v *viper.Viper
}

func NewFileStore() (outbound.ConfigStore, error) {
	v := viper.New()
	v.SetConfigName(".multix")
	v.SetConfigType("yaml")

	home, err := homedir.Dir()
	if err == nil {
		v.AddConfigPath(filepath.Join(home, ".config", "multix"))
		v.AddConfigPath(home)
	}
	v.AddConfigPath(".")

	v.SetEnvPrefix("MULTIX")
	v.AutomaticEnv()

	_ = v.ReadInConfig()

	return &viperAdapter{v: v}, nil
}

func (va *viperAdapter) GetString(key string) string      { return va.v.GetString(key) }
func (va *viperAdapter) GetBool(key string) bool          { return va.v.GetBool(key) }
func (va *viperAdapter) GetInt(key string) int            { return va.v.GetInt(key) }
func (va *viperAdapter) BindEnv(key, envVar string) error { return va.v.BindEnv(key, envVar) }
func (va *viperAdapter) WriteConfig() error               { return va.v.WriteConfig() }
