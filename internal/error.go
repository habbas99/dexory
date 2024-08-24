package internal

import (
	"errors"
)

var ErrEntityNotFound = errors.New("entity not found in database")
var ErrComparisonCaseNotSupported = errors.New("comparison case not supported")
