// +build mage

package main

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

func Lint() error {
	if err := sh.RunV("bash", "-c", "shopt -s globstar; shellcheck **/*.sh"); err != nil {
		return err
	}
	if err := sh.RunV("golangci-lint", "run"); err != nil {
		return err
	}
	if err := sh.RunV("go", "vet", "-v", "./..."); err != nil {
		return err
	}
	if err := sh.RunV("goimports", "-w", "-l", "."); err != nil {
		return err
	}
	if err := sh.RunV("go", "mod", "tidy"); err != nil {
		return err
	}
	return sh.RunV("git", "diff", "--exit-code")
}

func CheckLicenseHeaders() error {
	return sh.RunV("./check_license_headers.sh")
}

func Test() error {
	return sh.RunV("go", "test", "./...", "-race")
}

func Build() error {
	return sh.RunV("goreleaser", "release", "--rm-dist", "--snapshot")
}

func Release() error {
	mg.Deps(Test)
	return sh.RunV("goreleaser", "release", "--rm-dist")
}
