// Copyright The gotf Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"

	"github.com/unguiculus/gotf/pkg/config"
	"github.com/unguiculus/gotf/pkg/sh"
	"github.com/unguiculus/gotf/pkg/tf"
)

func main() {
	var cfgFile string

	flagSet := flag.NewFlagSet("gotf", flag.ExitOnError)
	flagSet.StringVarP(&cfgFile, "config", "c", "gotf.yaml", "Config file to be used")

	args := os.Args[1:]
	if len(args) <= 2 {
		usage(flagSet)
		os.Exit(1)
	}

	err := flagSet.Parse(args[:2])
	if err != nil {
		usage(flagSet)
		os.Exit(1)
	}

	cfg, err := config.Load(cfgFile)
	if err != nil {
		usage(flagSet)
		os.Exit(1)
	}

	shell := sh.Shell{}
	tf := terraform.NewTerraform(cfg, shell)
	exitCode, err := tf.Run(args[2:]...)

	if err != nil {
		fmt.Println("\n", err)
	}
	os.Exit(exitCode)
}

func usage(flagSet *flag.FlagSet) {
	fmt.Println("gotf [options] [Terraform args]")
	fmt.Println()
	flagSet.PrintDefaults()
}
