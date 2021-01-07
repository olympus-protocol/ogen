package bls

import (
	"errors"
	"github.com/olympus-protocol/ogen/pkg/bls/common"
)

// ErrZeroKey describes an error due to a zero secret key.
var ErrZeroKey = common.ErrZeroKey

// ErrInfinitePubKey describes an error due to an infinite public key.
var ErrInfinitePubKey = common.ErrInfinitePubKey

// ErrImplementationNotFound returns when selecting another bls implementation
var ErrImplementationNotFound = errors.New("the implementation requested is not found")
