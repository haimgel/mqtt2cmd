package config

import (
	"fmt"
	"github.com/haimgel/mqtt2cmd/internal/controls"
	"github.com/spf13/pflag"
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

type LoggerConfig struct {
	Path string `mapstructure:"path"`
}

type ApplicationConfig struct {
	Mqtt         MqttConfig        `mapstructure:"mqtt"`
	Switches     []controls.Switch `mapstructure:"switches"`
	LoggerConfig LoggerConfig      `mapstructure:"log"`
}

// Load configuration from command-line options, config file, and environment variables
func Load() (*ApplicationConfig, error) {
	processCommandLineArguments()

	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return nil, err
	}

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

func processCommandLineArguments() {
	pflag.StringP("mqtt.broker", "b", "", "MQTT broker (example \"tcp://hostname:1883\")")
	pflag.StringP("log.path", "l", defaultLogFile(), "Log file path")
	help := pflag.BoolP("help", "h", false, "")
	pflag.Parse()
	if *help {
		pflag.Usage()
		os.Exit(2)
	}
}

func defaultConfigHome() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	err = os.MkdirAll(cfgDir, 0750)
	return filepath.Join(cfgDir, AppName), err
}

func defaultLogFile() string {
	cfgDir, err := defaultConfigHome()
	if err != nil {
		return ""
	}
	return filepath.Join(cfgDir, fmt.Sprintf("%s.log", AppName))
}
