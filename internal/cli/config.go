package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/arisu-archive/protocol-encoder-go/pkg/encoder"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// initConfig reads in config file and ENV variables if set
func InitConfig(cfgFile string, logger *logrus.Logger) {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		logger.WithField("config", viper.ConfigFileUsed()).Info("Using config file")
	}
}

// SetupEncoder creates a new encoder with the given logger.
func SetupEncoder(logger *logrus.Logger) (*encoder.Encoder, error) {
	offset, err := parseOffset()
	if err != nil {
		return nil, fmt.Errorf("failed to get offset: %w", err)
	}

	enc, err := encoder.NewEncoder(&encoder.Config{
		Binary:   viper.GetString("binary"),
		Offset:   offset,
		PoolSize: viper.GetInt("pool"),
		Logger:   logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create encoder: %w", err)
	}

	return enc, nil
}

// parseOffset returns the offset from Viper
func parseOffset() (uint64, error) {
	offsetStr := strings.TrimSpace(viper.GetString("offset"))
	if offsetStr == "" {
		return 0, fmt.Errorf("offset cannot be empty")
	}

	offset, err := strconv.ParseUint(offsetStr, 0, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse offset '%s': %w", offsetStr, err)
	}

	return offset, nil
}
