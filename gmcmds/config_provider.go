package gmcmds

import (
	"github.com/mikeschinkel/gmover/cliutil"
	"github.com/mikeschinkel/gmover/gmover"
)

var _ gmover.ConfigProvider = (*configProvider)(nil)

// configProvider implements the scout.ConfigProvider interface
type configProvider struct {
}

func (cp *configProvider) config() *Config {
	sc, ok := cp.GetConfig().(*Config)
	if !ok {
		panic("Can't get config from ConfigProvider")
	}
	return sc
}

// GetConfig returns the global config instance
func (cp *configProvider) GetConfig() cliutil.Config {
	return GetConfig()
}

// GlobalFlagSet returns the global flag set
func (cp *configProvider) GlobalFlagSet() *cliutil.FlagSet {
	return GlobalFlagSet
}

// NewConfigProvider creates a new configProvider instance
func NewConfigProvider() gmover.ConfigProvider {
	return &configProvider{}
}
