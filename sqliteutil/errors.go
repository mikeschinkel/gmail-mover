package sqliteutil

import "errors"

var (
	ErrAppNameMustNotBeEmpty = errors.New("app name must not be empty")
	ErrFailedToGetConfigPath = errors.New("failed to get config path")
)
