package cliutil

// Config represents any config object that can be passed to commands

type Config interface {
	Config()
	SetValues(map[string]any)
}
