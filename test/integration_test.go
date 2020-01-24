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

package integrationtest

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
)

func Test_e2e(t *testing.T) {
	os.RemoveAll("testdata/.tfstate")
	os.RemoveAll(".terraform")

	tempDir, err := ioutil.TempDir("", "gotf")
	panicOnError(err)
	defer os.RemoveAll(tempDir)

	binary := buildBinary(tempDir)

	output, err := runProcess(binary, "-d", "-c", "testdata/test_config.yaml", "-p", "env=prod", "init", "-no-color", "testdata/testmodule")
	fmt.Println(output)
	assert.NoError(t, err)
	assert.Contains(t, output, "Terraform has been successfully initialized!")

	output, err = runProcess(binary, "-d", "-c", "testdata/test_config.yaml", "-p", "env=prod", "plan", "-no-color", "testdata/testmodule")
	fmt.Println(output)
	assert.NoError(t, err)
	assert.Contains(t, output, "# null_resource.echo will be created")
	assert.Contains(t, output, "Plan: 1 to add, 0 to change, 0 to destroy.")

	output, err = runProcess(binary, "-d", "-c", "testdata/test_config.yaml", "-p", "env=prod", "apply", "-auto-approve", "-no-color", "testdata/testmodule")
	fmt.Println(output)
	assert.NoError(t, err)
	assert.Contains(t, output, `baz = bazvalue
foo = 42
mapvar = {
  entry1 = {
    value1 = testvalue1
    value2 = true
  }
  entry2 = {
    value1 = testvalue2
    value2 = false
  }
}`)
}

func buildBinary(dir string) string {
	fmt.Println("Building application...")
	binary := filepath.Join(dir, "gotf")
	output, err := runProcess("go", "build", "-o", binary, "..")
	if err != nil {
		fmt.Println(output)
		panic(err)
	}
	fmt.Println("Build finished successfully")
	fmt.Printf("Using binary for test: %v\n\n", binary)
	return binary
}

func runProcess(binary string, files ...string) (string, error) {
	output, err := exec.Command(binary, files...).CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

func copyToDir(dst string, src string) {
	panicOnError(copy.Copy(dst, src))
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
