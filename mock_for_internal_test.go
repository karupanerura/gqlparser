package gqlparser

import (
	"github.com/karupanerura/gqlparser/internal/testutils"
)

var defaultTokenSourceFactory = testutils.NewTestTokenSourceFactory[Token](ErrEndOfToken)
