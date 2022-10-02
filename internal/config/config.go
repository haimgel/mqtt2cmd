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
func Load(version string, exit func(int), args []string) (*ApplicationConfig, error) {
	processCommandLineArguments(version, exit, args)

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

func processCommandLineArguments(versionStr string, exit func(int), args []string) {
	pflag.StringP("mqtt.broker", "b", "", "MQTT broker (example \"tcp://hostname:1883\")")
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
