// +build tools

package tools

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/goreleaser/goreleaser"
	_ "github.com/magefile/mage"
	_ "golang.org/x/tools/cmd/goimports"
)
