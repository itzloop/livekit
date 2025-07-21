package geoip

import (
	"fmt"
	"net/netip"
	"sync/atomic"

	"github.com/livekit/protocol/logger"
	"github.com/oschwald/geoip2-golang/v2"
)

var (
	GeoIP       *geoip2.Reader
	initialized atomic.Bool
)

const (
	defaultGeoIPPath = "/opt/maximind/geoip.mmdb"
)

type Config struct {
	Path    string `yaml:"path,omitempty"`
	Enabled bool   `yaml:"enabled,omitempty"`
}

// Init reads geoipdb and set up GeoIP reader
// if path is empty defaultGeoIPPath will be used
func Init(cfg Config) error {
	if !cfg.Enabled {
		logger.Infow("GeoIP is not enabled, ignoring")
		return nil
	}

	if initialized.Swap(true) {
		logger.Warnw("GeoIP has been already initialized", nil)
		return nil
	}

	logger.Infow("initializing geoip database")

	if cfg.Path == "" {
		cfg.Path = defaultGeoIPPath
	}

	reader, err := geoip2.Open(cfg.Path)
	if err != nil {
		logger.Errorw("failed to open geoip database", err, "path", cfg.Path)
		return fmt.Errorf("failed to open geoip database: %w", err)
	}

	GeoIP = reader

	return nil
}

func Close() {
	if !initialized.Load() {
		return
	}

	if err := GeoIP.Close(); err != nil {
		logger.Errorw("failed to close geoip database", err)
	}
}

func GetASOrganization(address string) string {
	if !initialized.Load() {
		return "geoip not initialized"
	}

	addr, err := netip.ParseAddr(address)
	if err != nil {
		logger.Errorw("geoip: failed to parse address", err, "addr", address)
		return "invalid address"
	}

	record, err := GeoIP.Enterprise(addr)
	if err != nil {
		logger.Infow("geoip: failed to query for address", err, "addr", address)
		return "unknown"
	}

	if !record.HasData() {
		return "not found"
	}

	return record.Traits.AutonomousSystemOrganization
}
