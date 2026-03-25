package config_editor

import "errors"

var (
	ErrConfigNotLoaded = errors.New("config not loaded")
	ErrSectionNotFound = errors.New("section not found")
	ErrKeyNotFound     = errors.New("key not found")
)
