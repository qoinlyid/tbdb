package tbdb

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/qoinlyid/qore"
	"github.com/spf13/viper"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

// Config defines Cache config.
type Config struct {
	// DependencyPriority defines priority of cache dependency.
	DependencyPriority int `json:"TBDB_DEPENDENCY_PRIORITY" mapstructure:"CACHE_DEPENDENCY_PRIORITY"`

	// ClusterID defines TigerBeetle cluster id.
	ClusterID   uint64        `json:"TBDB_CLUSTER_ID" mapstructure:"TBDB_CLUSTER_ID"`
	clusterIDTB types.Uint128 `json:"-" mapstructure:"-"`

	// Addresses defines TigerBeetle nodes address. Use comma separated to set multi nodes.
	Addresses string `json:"TBDB_ADDRESSES" mapstructure:"TBDB_ADDRESSES"`
}

// Default config.
var defaultConfig = &Config{
	DependencyPriority: 10,
}

// Load config.
func loadConfig() *Config {
	var e error
	config := defaultConfig

	// Get used config from OS env.
	configSource := os.Getenv(qore.CONFIG_USED_KEY)
	if qore.ValidationIsEmpty(configSource) {
		configSource = "OS"
	}
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	switch strings.ToUpper(configSource) {
	case "OS":
		if err := viper.Unmarshal(&config); err != nil {
			e = errors.Join(fmt.Errorf("failed to parse OS env value to config: %w", err))
		}
	default:
		ext := strings.ToLower(filepath.Ext(configSource))
		switch ext {
		case ".env":
			viper.SetConfigFile(configSource)
			viper.SetConfigType("env")
			if err := viper.ReadInConfig(); err != nil {
				e = errors.Join(fmt.Errorf("failed to read env file %s: %w", configSource, err))
			} else {
				if err := viper.Unmarshal(&config); err != nil {
					e = errors.Join(fmt.Errorf("failed to parse env file %s value to config: %w", configSource, err))
				}
			}
		case ".json", ".yml", ".yaml", ".toml":
			viper.SetConfigFile(configSource)
			if err := viper.ReadInConfig(); err != nil {
				e = errors.Join(fmt.Errorf("failed to read config file %s: %w", configSource, err))
			} else {
				if err := viper.Unmarshal(&config); err != nil {
					e = errors.Join(fmt.Errorf("failed to parse config file %s value to config: %w", configSource, err))
				}
			}
		}
	}
	if e != nil {
		log.Printf("dependency config - failed to load config: %s\n", e.Error())
	}

	// Config value modifier.
	if config.DependencyPriority == 0 {
		config.DependencyPriority = defaultConfig.DependencyPriority
	}
	config.clusterIDTB = types.ToUint128(config.ClusterID)
	return config
}
