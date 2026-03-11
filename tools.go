//go:build tools

package tools

// This file ensures test/tool dependencies are tracked in go.mod.
// The "tools" build tag is never set during normal builds.
import (
	_ "github.com/DATA-DOG/go-sqlmock"
	_ "github.com/alicebob/miniredis/v2"
	_ "github.com/stretchr/testify"
)
