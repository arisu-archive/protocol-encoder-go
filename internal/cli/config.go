package cli

import (
	"fmt"
	"strings"

	"github.com/arisu-archive/protocol-encoder-go/pkg/encoder"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// initConfig reads in config file and ENV variables if set
func InitConfig(cfgFile string, logger *logrus.Logger) (*encoder.Config, error) {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}
	viper.SetEnvPrefix("BA")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		logger.WithField("config", viper.ConfigFileUsed()).Info("Using config file")
	}

	// Validate required fields
	var cfg encoder.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set logger
	cfg.Logger = logger

	// Validate config
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Log loaded config (excluding binary path for security)
	logger.WithField("encoders", len(cfg.Encoders)).Info("Configuration loaded")
	return &cfg, nil
}

// SetupEncoder creates a new encoder with the given logger.
func SetupEncoder(commonCfg encoder.EmulatorConfig, cfg encoder.EncoderConfig, logger *logrus.Logger) (*encoder.Encoder, error) {
	// Create encoder config
	enc, err := encoder.NewEncoder(commonCfg, cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create encoder: %w", err)
	}

	return enc, nil
}
