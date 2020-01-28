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

package terraform

import (
	"fmt"
	"strings"

	"github.com/unguiculus/gotf/pkg/config"
)

type (
	Shell interface {
		Execute(env map[string]string, workingDir string, cmd string, args ...string) error
	}

	Terraform struct {
		config     *config.Config
		params     map[string]string
		moduleDir  string
		shell      Shell
		binaryPath string
	}
)

var commandsWithVars = []string{"apply", "destroy", "plan", "refresh", "import"}

func NewTerraform(config *config.Config, moduleDir string, params map[string]string, shell Shell, binaryPath string) *Terraform {
	return &Terraform{
		config:     config,
		params:     params,
		shell:      shell,
		moduleDir:  moduleDir,
		binaryPath: binaryPath,
	}
}

func (tf *Terraform) Execute(args ...string) error {
	env := map[string]string{}
	stringMapAppend(env, tf.config.Envs)
	tf.appendVarFileArgs(env)
	tf.appendVarArgs(env)
	tf.appendBackendConfigs(env)

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
			env["TF_VAR_backend_"+k] = v
			if sb.Len() > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(fmt.Sprintf("-backend-config=%s=%q", k, v))
		}
		env["TF_CLI_ARGS_init"] = sb.String()
	}
}

func stringMapAppend(target map[string]string, src map[string]string) {
	for k, v := range src {
		target[k] = v
	}
}
