// +build mage

package main

import (
	"github.com/magefile/mage/sh"
)

func Lint() error {
	if err := sh.Run("golangci-lint", "run"); err != nil {
		return err
	}
	if err := sh.Run("go", "vet", "-v", "./..."); err != nil {
		return err
	}
	if err := sh.Run("goimports", "-w", "-l", "."); err != nil {
		return err
	}
	if err := sh.Run("go", "mod", "tidy"); err != nil {
		return err
	}
	return sh.Run("git", "diff", "--exit-code")
}

func Test() error {
	return sh.Run("go", "test", "./...", "-race")
}

func Build() error {
	return sh.Run("goreleaser", "release", "--rm-dist", "--snapshot")
}

func Release() error {
	return sh.Run("goreleaser", "release", "--rm-dist")
}
