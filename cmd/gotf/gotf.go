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

package gotf

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/unguiculus/gotf/pkg/config"
	"github.com/unguiculus/gotf/pkg/opts"
	"github.com/unguiculus/gotf/pkg/sh"
	terraform "github.com/unguiculus/gotf/pkg/tf"
)

var (
	version = "dev"
	gitCommit = "HEAD"
	buildDate = "unknown"
)

func Execute() {
	var cfgFile string
	params := opts.NewMapOpts()
	var debug bool

	fullVersion := fmt.Sprintf("%s (commit=%s, date=%s)", version, gitCommit, buildDate)
	cmd := &cobra.Command{
		Use:     "gotf [flags] [Terraform args]",
		Short:   "gotf is a Terraform wrapper facilitating configurations for various environments",
		Long:   fmt.Sprintf(`
  ___   __  ____  ____
 / __) /  \(_  _)(  __)
( (_ \(  O ) )(   ) _)
 \___/ \__/ (__) (__)   %s

gotf is a Terraform wrapper facilitating configurations for various environments
`, fullVersion),
		Version: fullVersion,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("no arguments for Terraform specified")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			Run(debug, cfgFile, params.GetAll(), args...)
		},
	}

	cmd.Flags().StringVarP(&cfgFile, "config", "c", "gotf.yaml", "Config file to be used")
	cmd.Flags().VarP(params, "params", "p", "Params for templating in the config file. May be specified multiple times")
	cmd.Flags().BoolVarP(&debug, "debug", "d", false, "Print additional debug output")
	cmd.Flags().SetInterspersed(false)
	cmd.SetVersionTemplate("{{ .Version }}\n")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func Run(debug bool, cfgFile string, params map[string]string, args ...string) {
	cfg, err := config.Load(cfgFile, params)

	shell := sh.NewShell(debug)
	tf := terraform.NewTerraform(cfg, params, shell)
	exitCode, err := tf.Run(args...)

	if err != nil {
		fmt.Println("\n", err)
	}
	os.Exit(exitCode)
}
