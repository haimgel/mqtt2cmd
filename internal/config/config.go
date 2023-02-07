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
	AppId        string            `mapstructure:"app-id"`
	Mqtt         MqttConfig        `mapstructure:"mqtt"`
	Switches     []controls.Switch `mapstructure:"switches"`
	LoggerConfig LoggerConfig      `mapstructure:"log"`
}

// Load configuration from command-line options, config file, and environment variables
func Load(version string, exit func(int), args []string) (*ApplicationConfig, error) {
	processCommandLineArguments(version, exit, args)

	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return nil, err
	}

	viper.SetEnvPrefix(strings.ToUpper(AppName))
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	viper.SetConfigFile(viper.GetString("config"))
	viper.SetDefault("app-id", AppName)
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

func processCommandLineArguments(versionStr string, exit func(int), args []string) {
	pflag.StringP("config", "c", defaultConfigFile(), "Configuration file path")
	pflag.StringP("mqtt.broker", "b", "tcp://localhost:1883", "MQTT broker")
	pflag.StringP("log.path", "l", defaultLogFile(), "Log file path")
	helpFlag := pflag.BoolP("help", "h", false, "This help message")
	versionFlag := pflag.BoolP("version", "v", false, "Show version")
	_ = pflag.CommandLine.Parse(args)
	if *helpFlag {
		pflag.Usage()
		exit(2)
	}
	if *versionFlag {
		fmt.Printf("%s version %s\n", AppName, versionStr)
		exit(0)
	}
}

func defaultConfigHome() (string, error) {
	cfgDir, err := os.UserConfigDir()
	return filepath.Join(cfgDir, AppName), err
}

func defaultConfigFile() string {
	configDir, err := defaultConfigHome()
	if err != nil {
		return ""
	}
	return filepath.Join(configDir, "config.yaml")
}

func defaultLogFile() string {
	cfgDir, err := defaultConfigHome()
	if err != nil {
		return ""
	}
	return filepath.Join(cfgDir, fmt.Sprintf("%s.log", AppName))
}
