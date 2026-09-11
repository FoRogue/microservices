//go:build tools

// Package main pins developer tool dependencies so `go mod tidy` keeps them
// in go.sum even though they are only invoked via `go run`, never imported.
package main

import (
	_ "github.com/swaggo/swag/cmd/swag"
)
