// SPDX-FileCopyrightText: Winni Neessen <wn@neessen.dev>
//
// SPDX-License-Identifier: MIT

package config

import (
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
}
