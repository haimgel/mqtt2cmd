package config

import (
	"github.com/haimgel/mqtt-buttons/internal/controls"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

const AppName = "mqtt2cmd"

type MqttConfig struct {
	Broker   string  `mapstructure:"broker"`
	User     *string `mapstructure:"username"`
	Password *string `mapstructure:"password"`
}

type ApplicationConfig struct {
	Mqtt     MqttConfig        `mapstructure:"mqtt"`
	Switches []controls.Switch `mapstructure:"switches"`
}

// Load configuration from command-line options, config file, and environment variables
func Load() (*ApplicationConfig, error) {
	viper.SetEnvPrefix(strings.ToUpper(AppName))
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	configDir, err := defaultConfigHome()
	if err != nil {
		return nil, err
	}
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(configDir)
	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	config := ApplicationConfig{}
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func defaultConfigHome() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cfgDir, AppName), nil
}
