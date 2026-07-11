package config

import "errors"

var ErrRootNotSet = errors.New("root path not set, run 'jgo root <path>' first")
