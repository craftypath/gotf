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

package sh

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Shell struct {
	debug bool
}

func NewShell(debug bool) Shell {
	return Shell{debug: debug}
}

func (s Shell) Execute(env map[string]string, cmd string, args ...string) (int, error) {
	c := exec.Command(cmd, args...)

	c.Env = os.Environ()

	if s.debug {
		fmt.Println()
		fmt.Println("Terraform command-line:")
		fmt.Println("-----------------------")
		fmt.Println(cmd, strings.Join(args, " "))
		fmt.Println()
		fmt.Println("Terraform environment:")
		fmt.Println("----------------------")
	}

	for k, v := range env {
		c.Env = append(c.Env, k+"="+v)
		if s.debug {
			fmt.Printf("%s=%s\n", k, v)
		}
	}

	if s.debug {
		fmt.Println()
	}

	c.Stderr = os.Stdout
	c.Stdout = os.Stderr
	c.Stdin = os.Stdin

	err := c.Run()
	if err == nil {
		return 0, nil
	}

	if err, ok := err.(*exec.ExitError); ok {
		return err.ExitCode(), err
	}

	return 1, err
}
