// Copyright The gotf Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gotf

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/craftypath/gotf/pkg/gotf"
	"github.com/craftypath/gotf/pkg/opts"
)

func Execute() {
	command := newGotfCommand()
	if err := command.Execute(); err != nil {
		var exitCode int
		if err, ok := err.(*exec.ExitError); ok {
			exitCode = err.ExitCode()
		} else {
			exitCode = 1
		}
		os.Exit(exitCode)
	}
}

func newGotfCommand() *cobra.Command {
	var cfgFile string
	params := opts.NewMapOpts()
	var debug bool
	var moduleDir string
	var skipBackendCheck bool
	var noVars bool

	fullVersion := fmt.Sprintf("%s (commit=%s, date=%s)", gotf.Version, gotf.GitCommit, gotf.BuildDate)
	command := &cobra.Command{
		Use:   "gotf [flags] [Terraform args]",
		Short: "gotf is a Terraform wrapper facilitating configurations for various environments",
		Long: fmt.Sprintf(`
  ___   __  ____  ____
 / __) /  \(_  _)(  __)
( (_ \(  O ) )(   ) _)
 \___/ \__/ (__) (__)   %s

gotf is a Terraform wrapper facilitating configurations for various environments
`, fullVersion),
		Version: fullVersion,
		RunE: func(command *cobra.Command, args []string) error {
			return gotf.Run(gotf.Args{
				Debug:            debug,
				ConfigFile:       cfgFile,
				ModuleDir:        moduleDir,
				Params:           params.GetAll(),
				SkipBackendCheck: skipBackendCheck,
				NoVars:           noVars,
				Args:             args,
			})
		},
	}

	command.Flags().StringVarP(&cfgFile, "config", "c", "gotf.yaml", "Config file to be used")
	command.Flags().VarP(params, "params", "p", "Params for templating in the config file. May be specified multiple times")
	command.Flags().BoolVarP(&debug, "debug", "d", false, "Print additional debug output to stderr")
	command.Flags().StringVarP(&moduleDir, "module-dir", "m", "", "The module directory to run Terraform in")
	command.Flags().BoolVarP(&skipBackendCheck, "skip-backend-check", "s", false, "Skip checking for changed backend configuration")
	command.Flags().BoolVarP(&noVars, "no-vars", "n", false, `Don't add any variables when running Terraform.
This is necessary when running 'terraform apply' with a plan file.`)
	command.Flags().SetInterspersed(false)
	command.SetVersionTemplate("{{ .Version }}\n")
	command.SilenceUsage = true
	if err := command.MarkFlagRequired("module-dir"); err != nil {
		log.Fatalln(err)
	}
	return command
}
