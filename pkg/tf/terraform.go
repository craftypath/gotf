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

package terraform

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/craftypath/gotf/pkg/config"
)

type (
	Shell interface {
		Execute(env map[string]string, workingDir string, cmd string, args ...string) error
	}

	Terraform struct {
		config           *config.Config
		params           map[string]string
		moduleDir        string
		noVars           bool
		skipBackendCheck bool
		shell            Shell
		binaryPath       string
	}
)

var commandsWithVars = []string{"apply", "destroy", "plan", "refresh", "import"}

func NewTerraform(config *config.Config, moduleDir string, params map[string]string, skipBackendCheck bool, noVars bool, shell Shell, binaryPath string) *Terraform {
	return &Terraform{
		config:           config,
		params:           params,
		shell:            shell,
		moduleDir:        moduleDir,
		skipBackendCheck: skipBackendCheck,
		noVars:           noVars,
		binaryPath:       binaryPath,
	}
}

func (tf *Terraform) Execute(args ...string) error {
	env := map[string]string{}
	stringMapAppend(env, tf.config.Envs)
	if !tf.noVars {
		tf.appendVarFileArgs(env)
		tf.appendVarArgs(env)
		tf.appendBackendConfigs(env)
	}

	if !tf.skipBackendCheck {
		if err := tf.checkBackendConfig(args...); err != nil {
			return err
		}
	}
	return tf.shell.Execute(env, tf.moduleDir, tf.binaryPath, args...)
}

func (tf *Terraform) appendVarArgs(env map[string]string) {
	for k, v := range tf.config.Vars {
		env["TF_VAR_"+k] = v
	}
}

func (tf *Terraform) appendVarFileArgs(env map[string]string) {
	varFiles := tf.config.VarFiles
	if len(varFiles) > 0 {
		sb := strings.Builder{}
		for _, f := range varFiles {
			if sb.Len() > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(fmt.Sprintf("-var-file=%q", f))
		}
		varFilesArgs := sb.String()
		for _, cmd := range commandsWithVars {
			env["TF_CLI_ARGS_"+cmd] = varFilesArgs
		}
	}
}

func (tf *Terraform) appendBackendConfigs(env map[string]string) {
	configs := tf.config.BackendConfigs
	if len(configs) > 0 {
		sb := strings.Builder{}
		for k, v := range configs {
			if sb.Len() > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(fmt.Sprintf("-backend-config=%s=%q", k, v))
		}
		env["TF_CLI_ARGS_init"] = sb.String()
	}
}

func (tf *Terraform) checkBackendConfig(args ...string) error {
	if len(args) >= 2 {
		if args[0] == "init" {
			for _, arg := range args[1:] {
				if arg == "-reconfigure" {
					return nil
				}
			}
		}
	}

	backendFile := filepath.Join(tf.moduleDir, ".terraform", "terraform.tfstate")
	b, err := ioutil.ReadFile(backendFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var backendJSON map[string]interface{}
	if err := json.Unmarshal(b, &backendJSON); err != nil {
		return err
	}

	sb := strings.Builder{}
	for k, v := range tf.config.BackendConfigs {
		b := backendJSON["backend"].(map[string]interface{})
		c := b["config"].(map[string]interface{})
		currentVal := c[k]
		if v != currentVal {
			sb.WriteString(fmt.Sprintf("%s: got=%v, want=%v\n", k, currentVal, v))
		}
	}

	if sb.Len() > 0 {
		sb.WriteString("\nRun terraform init -reconfigure!\n")
		return fmt.Errorf("configured backend does not match current environment\n\n%s", sb.String())
	}

	return nil
}

func stringMapAppend(target map[string]string, src map[string]string) {
	for k, v := range src {
		target[k] = v
	}
}
