package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kkyr/fig"
)

type Config struct {
	Units    string `fig:"units" default:"metric"`
	Locale   string `fig:"locale"`
	Database struct {
		BusyTimeout    time.Duration `fig:"busy_timeout" default:"10s"`
		MaxConnections int           `fig:"max_connections" default:"5"`
		Path           string        `fig:"path"`
	}
	Geobus struct {
		CoordinatesFile        string `fig:"coordinates_file"`
		CitynameFile           string `fig:"cityname_file"`
		DisableGeoIP           bool   `fig:"disable_geoip"`
		DisableGeoAPI          bool   `fig:"disable_geoapi"`
		DisableCoordinatesFile bool   `fig:"disable_coordinates_file"`
		DisableCitynameFile    bool   `fig:"disable_cityname_file"`
		DisableICHNAEA         bool   `fig:"disable_ichnaea"`
		DisableGPSD            bool   `fig:"disable_gpsd"`
	} `fig:"geobus"`
	Geocoder struct {
		Provider string `fig:"provider" default:"nominatim"`
		APIKey   string `fig:"apikey"`
	} `fig:"geocoder"`
	Log struct {
		Format string   `fig:"format" default:"json"`
		Output string   `fig:"output" default:"stdout"`
		Level  LogLevel `fig:"level" default:"info"`
	}
	Server struct {
		Address        string        `fig:"bindaddr" default:"127.0.0.1"`
		Port           uint32        `fig:"port" default:"10001"`
		Timeout        time.Duration `fig:"timeout" default:"5s"`
		DevEnvironment bool          `fig:"dev_env"`
	}
	Scheduler struct {
		MaintenanceInterval   time.Duration `fig:"maintenance_interval" default:"1h"`
		WeatherUpdateInterval time.Duration `fig:"weatherdata_update_interval" default:"15m"`
	}
	Weather struct {
		Provider string `fig:"provider" default:"open-meteo"`
		IconSet  string `fig:"iconset" default:"meteocons"`

		// Cold and hot class thresholds (Defaults are based on °C)
		// Defaults are based on suggestions for dangerous driving conditions and uncomfortable heat.
		ColdThreshold float64 `fig:"cold_threshold" default:"2"`
		HotThreshold  float64 `fig:"hot_threshold" default:"30"`
	} `fig:"weather"`
}

// New creates a new instance of Config by reading and loading configuration values. It takes in the file
// path and file name of the configuration file as parameters. It returns a pointer to the Config and an
// error if there was a problem reading or loading the configuration.
func New(path, file string) (*Config, error) {
	config := Config{}
	_, err := os.Stat(fmt.Sprintf("%s/%s", path, file))
	if err != nil {
		return &config, fmt.Errorf("failed to read config: %w", err)
	}

	if err = fig.Load(&config, fig.Dirs(path), fig.File(file), fig.UseEnv("localweather")); err != nil {
		return &config, fmt.Errorf("failed to load config: %w", err)
	}

	return &config, nil
}

// Validate checks and updates the Config object to ensure required fields are set, assigning defaults if necessary.
func (c *Config) Validate() error {
	home, _ := os.UserHomeDir()
	confDir := filepath.Join(home, ".config", "localweather")
	if c.Geobus.CitynameFile == "" {
		c.Geobus.CitynameFile = filepath.Join(confDir, "cityname")
	}

	if c.Geobus.CoordinatesFile == "" {
		c.Geobus.CoordinatesFile = filepath.Join(confDir, "coordinates")
	}

	if c.Database.Path == "" {
		c.Database.Path = filepath.Join(confDir, "storage.db")
	}
	return nil
}

// ListenAddr constructs and returns the server address by combining the server's address and port from
// the configuration.
func (c *Config) ListenAddr() string {
	serverAddr := ":" + fmt.Sprintf("%d", c.Server.Port)
	if c.Server.Address != "" {
		serverAddr = c.Server.Address + serverAddr
	}
	return serverAddr
}
