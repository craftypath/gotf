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

package sh

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

type Shell struct{}

func (s Shell) Execute(env map[string]string, workingDir string, cmd string, args ...string) error {
	log.Println()
	log.Println("Terraform command-line:")
	log.Println("-----------------------")
	log.Println(cmd, strings.Join(args, " "))
	log.Println()
	log.Println("Terraform environment:")
	log.Println("----------------------")

	c := exec.Command(cmd, args...)
	c.Dir = workingDir
	c.Env = os.Environ()
	for k, v := range env {
		c.Env = append(c.Env, k+"="+v)
		log.Printf("%s=%s\n", k, v)
	}

	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin

	return c.Run()
}
