package gqlparser_test

import (
	"github.com/karupanerura/gqlparser"
	"github.com/karupanerura/gqlparser/internal/testutils"
)

var defaultTokenSourceFactory = testutils.NewTestTokenSourceFactory[gqlparser.Token](gqlparser.ErrEndOfToken)
