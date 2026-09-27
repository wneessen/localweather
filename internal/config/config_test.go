// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package config

import (
	"log/slog"
	"testing"
	"time"
)

const (
	testPath      = "../../testdata"
	testFile      = "config.toml"
	testFileEmpty = "config-empty.toml"
)

func TestNew(t *testing.T) {
	t.Run("New with valid path and file succeeds", func(t *testing.T) {
		conf, err := New(testPath, testFile)
		if err != nil {
			t.Fatalf("failed to create config: %s", err)
		}
		if conf == nil {
			t.Fatalf("config is nil")
		}
	})
	t.Run("New with empty config file uses default values", func(t *testing.T) {
		var (
			defaultUnits                                  = "metric"
			defaultDBBusyTimeout                          = time.Second * 10
			defaultDBMaxConnections                       = 5
			defaultServerAddr                             = "127.0.0.1"
			defaultServerPort                     uint32  = 10001
			defaultServerTimeout                          = time.Second * 5
			defaultSchedulerMaintenanceInterval           = time.Hour * 1
			defaultSchedulerWeatherUpdateInterval         = time.Minute * 15
			defaultWeatherProvider                        = "open-meteo"
			defaultWeatherIconset                         = "meteocons"
			defaultWeatherColdThreshold           float64 = 2
			defaultWeatherHotThreshold            float64 = 30
		)
		conf, err := New(testPath, testFileEmpty)
		if err != nil {
			t.Fatalf("failed to create config: %s", err)
		}
		if conf == nil {
			t.Fatalf("config is nil")
		}
		if conf.Units != defaultUnits {
			t.Errorf("expected units to be %s, got %s", defaultUnits, conf.Units)
		}
		if conf.Database.BusyTimeout != defaultDBBusyTimeout {
			t.Errorf("expected busy timeout to be %s, got %s", time.Second*10, conf.Database.BusyTimeout)
		}
		if conf.Database.MaxConnections != defaultDBMaxConnections {
			t.Errorf("expected max connections to be %d, got %d", 5, conf.Database.MaxConnections)
		}
		if conf.Server.Address != defaultServerAddr {
			t.Errorf("expected server address to be %s, got %s", defaultServerAddr, conf.Server.Address)
		}
		if conf.Server.Port != defaultServerPort {
			t.Errorf("expected server port to be %d, got %d", defaultServerPort, conf.Server.Port)
		}
		if conf.Server.Timeout != defaultServerTimeout {
			t.Errorf("expected server timeout to be %s, got %s", defaultServerTimeout, conf.Server.Timeout)
		}
		if conf.Scheduler.MaintenanceInterval != defaultSchedulerMaintenanceInterval {
			t.Errorf("expected scheduler maintenance interval to be %s, got %s",
				defaultSchedulerMaintenanceInterval, conf.Scheduler.MaintenanceInterval)
		}
		if conf.Scheduler.WeatherUpdateInterval != defaultSchedulerWeatherUpdateInterval {
			t.Errorf("expected scheduler weather update interval to be %s, got %s",
				defaultSchedulerWeatherUpdateInterval, conf.Scheduler.WeatherUpdateInterval)
		}
		if conf.Weather.Provider != defaultWeatherProvider {
			t.Errorf("expected weather provider to be %s, got %s", defaultWeatherProvider, conf.Weather.Provider)
		}
		if conf.Weather.IconSet != defaultWeatherIconset {
			t.Errorf("expected weather iconset to be %s, got %s", defaultWeatherIconset, conf.Weather.IconSet)
		}
		if conf.Weather.ColdThreshold != defaultWeatherColdThreshold {
			t.Errorf("expected weather cold threshold to be %f, got %f", defaultWeatherColdThreshold,
				conf.Weather.ColdThreshold)
		}
		if conf.Weather.HotThreshold != defaultWeatherHotThreshold {
			t.Errorf("expected weather hot threshold to be %f, got %f", defaultWeatherHotThreshold,
				conf.Weather.HotThreshold)
		}
	})
	t.Run("New with no config file path uses env variables", func(t *testing.T) {
		config, err := New("", "")
		if err != nil {
			t.Fatalf("failed to create config with no config file path: %s", err)
		}
		if config == nil {
			t.Fatalf("expected config to be non-nil")
		}
	})
	t.Run("New with invalid config file path fails", func(t *testing.T) {
		_, err := New("invalid", "config.toml")
		if err == nil {
			t.Fatalf("expected error when creating config with invalid config file path")
		}
	})
}

func TestConfig_Validate(t *testing.T) {
	t.Run("Validate sets defaults if not set via config file", func(t *testing.T) {
		config, err := New("", "")
		if err != nil {
			t.Fatalf("failed to create config with no config file path: %s", err)
		}
		if config == nil {
			t.Fatalf("expected config to be non-nil")
		}
		if err = config.Validate(); err != nil {
			t.Errorf("failed to validate config: %s", err)
		}
		if config.Geobus.CitynameFile == "" {
			t.Error("expected validate to set default cityname file path")
		}
		if config.Geobus.CoordinatesFile == "" {
			t.Error("expected validate to set default coordinates file path")
		}
		if config.Database.Path == "" {
			t.Error("expected validate to set default database path")
		}
	})
}

func TestConfig_ListenAddr(t *testing.T) {
	t.Run("ListenAddr returns a valid server listen address", func(t *testing.T) {
		wantAddr := "127.0.0.1:10001"
		config, err := New("", "")
		if err != nil {
			t.Fatalf("failed to create config with no config file path: %s", err)
		}
		if config == nil {
			t.Fatalf("expected config to be non-nil")
		}
		addr := config.ListenAddr()
		if wantAddr != addr {
			t.Errorf("expected ListenAddr to return %s, got %s", wantAddr, addr)
		}
	})
}

func TestLogLevel_SLog(t *testing.T) {
	tests := []struct {
		name string
		has  LogLevel
		want slog.Level
	}{
		{name: "trace", has: LevelTrace, want: slog.LevelDebug},
		{name: "debug", has: LevelDebug, want: slog.LevelDebug},
		{name: "info", has: LevelInfo, want: slog.LevelInfo},
		{name: "warn", has: LevelWarn, want: slog.LevelWarn},
		{name: "error", has: LevelError, want: slog.LevelError},
		{name: "unknown", has: LevelUnknown, want: slog.LevelInfo},
		{name: "default", has: -999, want: slog.LevelInfo},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.has.SLog()
			if got != tc.want {
				t.Errorf("expected slog to return %s, got %s", tc.want, got)
			}
		})
	}
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		has  LogLevel
		want string
	}{
		{want: "trace", has: LevelTrace},
		{want: "debug", has: LevelDebug},
		{want: "info", has: LevelInfo},
		{want: "warn", has: LevelWarn},
		{want: "error", has: LevelError},
		{want: "unknown", has: LevelUnknown},
		{want: "unknown", has: -999},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			got := tc.has.String()
			if got != tc.want {
				t.Errorf("expected string to return %s, got %s", tc.want, got)
			}
		})
	}
}

func TestLogLevel_UnmarshalString(t *testing.T) {
	tests := []struct {
		has     string
		want    LogLevel
		wantErr bool
	}{
		{"trace", LevelTrace, false},
		{"debug", LevelDebug, false},
		{"info", LevelInfo, false},
		{"warn", LevelWarn, false},
		{"error", LevelError, false},
		{"unknown", LevelUnknown, true},
	}

	for _, tc := range tests {
		t.Run(tc.has, func(t *testing.T) {
			level := new(LogLevel)
			if err := level.UnmarshalString(tc.has); err != nil && !tc.wantErr {
				t.Errorf("failed to unmarshal log level string %q: %s", tc.has, err)
			}
			if *level != tc.want {
				t.Errorf("expected unmarshal to return %s, got %s", tc.want, level)
			}
		})
	}
}
